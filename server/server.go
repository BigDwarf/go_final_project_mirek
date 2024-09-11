package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/MirekKrassilnikov/go_final_project/repeater"
	"net/http"
	"strconv"
	"time"
)

const layout = "20060102"

type Task struct {
	ID      int    `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type Response struct {
	ID    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "../scheduler.db")
	if err != nil {
		http.Error(w, `{"error":"Failed to connect to database"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()
	idStr := r.URL.Query().Get("id")
	switch r.Method {
	case http.MethodPost:
		HandlePost(w, r) // Обработка POST-запросов

	case http.MethodGet:
		getTaskById(w, db, idStr) // Обработка GET-запросов
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPut:
		UpdateTask(w, r)

	case http.MethodDelete:
		DeleteTaskByID(w, r)
	}

}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Проверка обязательного поля id и title
	if task.ID == 0 {
		http.Error(w, `{"error":"ID is required"}`, http.StatusBadRequest)
		return
	}
	if len(task.Title) == 0 {
		http.Error(w, `{"error":"Title is required"}`, http.StatusBadRequest)
		return
	}

	// Проверка формата даты
	_, err = time.Parse(layout, task.Date)
	if err != nil {
		http.Error(w, `{"error":"Invalid date format"}`, http.StatusBadRequest)
		return
	}

	// Подключаемся к базе данных
	db, err := sql.Open("sqlite3", "../scheduler.db")
	if err != nil {
		http.Error(w, `{"error":"Failed to connect to database"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Проверяем, существует ли задача
	var existingTask Task
	err = db.QueryRow("SELECT id FROM scheduler WHERE id = ?", task.ID).Scan(&existingTask.ID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error":"Server error"}`, http.StatusInternalServerError)
		return
	}

	// Выполняем обновление задачи
	updateSQL := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err = db.Exec(updateSQL, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		http.Error(w, `{"error":"Failed to update task"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем пустой JSON в случае успешного обновления
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

/*
	func HandleGet(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("sqlite3", "scheduler.db")
		if err != nil {
			http.Error(w, `{"error":"Failed to connect to database"}`, http.StatusInternalServerError)
			return
		}
		defer db.Close()

		v

		if idStr == "" {
			getAllTasksHandler(w, db)
		} else {
			getTaskById(w, db, idStr)
		}
	}
*/
func GetAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "../scheduler.db")
	if err != nil {
		http.Error(w, `{"error":"Failed to connect to database"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func getTaskById(w http.ResponseWriter, db *sql.DB, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	// Выполняем SQL-запрос для получения задачи по id
	var task Task
	err = db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем заголовок Content-Type для JSON
	w.Header().Set("Content-Type", "application/json")

	// Преобразуем задачу в JSON и отправляем ответ
	json.NewEncoder(w).Encode(task)
}

func HandlePost(w http.ResponseWriter, r *http.Request) {
	var task Task
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Проверка обязательного поля title
	if len(task.Title) == 0 {
		http.Error(w, `{"error":"Title is required"}`, http.StatusBadRequest)
		return
	}

	// Проверка формата даты и установка текущей даты, если дата некорректна
	timeTimeDate, err := time.Parse(layout, task.Date)
	if err != nil {
		http.Error(w, `{"error":"Invalid date format"}`, http.StatusBadRequest)
		return
	}
	if timeTimeDate.Before(time.Now()) {
		// Если дата задачи меньше текущей даты и есть правило повторения
		if task.Repeat != "" {
			nextDate, err := repeater.NextDate(time.Now().Format(layout), task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Invalid repeat rule"}`, http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		} else {
			task.Date = time.Now().Format(layout)
		}
	}
	db, err := sql.Open("sqlite3", "../scheduler.db")
	if err != nil {
		http.Error(w, `{"error":"Failed to connect to database"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()

	insertSQL := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?);`
	result, err := db.Exec(insertSQL, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, `{"error":"Failed to insert task"}`, http.StatusInternalServerError)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve task ID"}`, http.StatusInternalServerError)
		return
	}
	response := Response{
		ID: int(id),
	}
	responseData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, `{"error":"Failed to marshal response"}`, http.StatusInternalServerError)
		return
	}

	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseData)
}

func MainHandle(res http.ResponseWriter, req *http.Request) {
	out := "Hello from server package!"
	res.Write([]byte(out))
}

func MarkAsDone(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	db, err := sql.Open("sqlite3", "../scheduler.db")
	if err != nil {
		http.Error(w, `{"error":"Failed to connect to database"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()
	var task Task
	err = db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		}
		return
	}

	if task.Repeat == "" {
		err = DeleteTaskByID(w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
		return
	}
	now := time.Now().Format("20060102")
	nextDate, err := repeater.NextDate(now, task.Date, task.Repeat)
	if err != nil {
		http.Error(w, "Error with calculating next date", http.StatusInternalServerError)
		return
	}

	updateSQL := `UPDATE scheduler SET date = ? WHERE id = ?`
	_, err = db.Exec(updateSQL, nextDate, task.ID)
	if err != nil {
		http.Error(w, `{"error":"Failed to update task"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем пустой JSON в случае успешного обновления
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{}`))

}

func DeleteTaskByID(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("id")
	if id == "" {
		return fmt.Errorf("task ID is required")
	}

	// Подключаемся к базе данных
	db, err := sql.Open("sqlite3", "../scheduler.db")
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Выполняем удаление задачи
	_, err = db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}

	// Если задача успешно удалена, возвращаем пустой JSON {}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{}`))
	return nil
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"error": message,
	}
	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}
func ApiNextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	/* Парсим текущую дату
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Invalid 'now' date format", http.StatusBadRequest)
		return
	}
	*/
	// Вызываем функцию NextDate
	nextDate, err := repeater.NextDate(nowStr, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}

package server

import (
	"database/sql"
	"encoding/json"
	"errors"
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

type Controller struct {
	DB *sql.DB
}

func (ctl *Controller) TaskHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	switch r.Method {
	case http.MethodPost:
		ctl.HandlePost(w, r) // Обработка POST-запросов
	case http.MethodGet:
		ctl.getTaskById(w, idStr) // Обработка GET-запросов
	case http.MethodPut:
		ctl.UpdateTask(w, r)

	case http.MethodDelete:
		ctl.DeleteTaskByID(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func (ctl *Controller) UpdateTask(w http.ResponseWriter, r *http.Request) {
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

	// Проверяем, существует ли задача
	var existingTask Task
	err = ctl.DB.QueryRow("SELECT id FROM scheduler WHERE id = ?", task.ID).Scan(&existingTask.ID)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"Task not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error":"Server error"}`, http.StatusInternalServerError)
		return
	}

	// Выполняем обновление задачи
	updateSQL := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	_, err = ctl.DB.Exec(updateSQL, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		http.Error(w, `{"error":"Failed to update task"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем пустой JSON в случае успешного обновления
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func (ctl *Controller) GetAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := ctl.DB.Query("SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC")
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

func (ctl *Controller) getTaskById(w http.ResponseWriter, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	// Выполняем SQL-запрос для получения задачи по id
	var task Task
	err = ctl.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
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

func (ctl *Controller) HandlePost(w http.ResponseWriter, r *http.Request) {
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
	insertSQL := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?);`
	result, err := ctl.DB.Exec(insertSQL, task.Date, task.Title, task.Comment, task.Repeat)
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

func (ctl *Controller) MarkAsDone(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	var task Task
	err := ctl.DB.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id).
		Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
		}
		return
	}

	if task.Repeat == "" {
		err = ctl.DeleteTaskByID(w, r)
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
	_, err = ctl.DB.Exec(updateSQL, nextDate, task.ID)
	if err != nil {
		http.Error(w, `{"error":"Failed to update task"}`, http.StatusInternalServerError)
		return
	}

	// Отправляем пустой JSON в случае успешного обновления
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{}`))

}

func (ctl *Controller) DeleteTaskByID(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("id")
	if id == "" {
		return fmt.Errorf("task ID is required")
	}

	// Выполняем удаление задачи
	_, err := ctl.DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
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
func (ctl *Controller) ApiNextDateHandler(w http.ResponseWriter, r *http.Request) {
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

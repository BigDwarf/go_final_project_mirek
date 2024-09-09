package server

import (
	"database/sql"
	"encoding/json"
	"github.com/MirekKrassilnikov/go_final_project/repeater"
	"net/http"
	"strconv"
	"time"
)

const layout = "20060102"

type Task struct {
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
	switch r.Method {
	case http.MethodPost:
		HandlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

		//case http.MethodGet:
		//	HandleGet(w, r)
	}

}
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		http.Error(w, `{"error":"Failed to connect to database"}`, http.StatusInternalServerError)
		return
	}
	// Получаем параметр id из запроса
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	// Преобразуем id в число
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
			http.Error(w, "task not found", http.StatusNotFound)
		} else {
			http.Error(w, "server error", http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем заголовок Content-Type для JSON
	w.Header().Set("Content-Type", "application/json")

	// Преобразуем задачу в JSON и отправляем ответ
	json.NewEncoder(w).Encode(task)
}

/*func HandleGet(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		http.Error(w, `{"error":"Failed to connect to database"}`, http.StatusInternalServerError)
		return
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM scheduler ORDER BY date ASC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		var date time.Time

		// Сканируем поля Date, Title, Comment и Repeat из базы данных
		err := rows.Scan(&date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Преобразуем дату в нужный формат 20060102
		task.Date = date.Format("20060102")

		// Добавляем задачу в срез
		tasks = append(tasks, task)
	}
	response := map[string][]Task{
		"tasks": tasks,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

*/

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
		task.Date = time.Now().Format(layout)
	} else if timeTimeDate.Before(time.Now()) {
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
	db, err := sql.Open("sqlite", "scheduler.db")
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
	w.Write(responseData)
}

func MainHandle(res http.ResponseWriter, req *http.Request) {
	out := "Hello from server package!"
	res.Write([]byte(out))
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

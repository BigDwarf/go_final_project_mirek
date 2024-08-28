package server

import (
	"database/sql"
	"encoding/json"
	"github.com/MirekKrassilnikov/go_final_project/repeater"
	"net/http"
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
	}
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

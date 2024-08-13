package main

import (
	"encoding/json"
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

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func handlePost(w http.ResponseWriter, r *http.Request) {
	var task Task
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Проверка обязательного поля title
	if len(task.Title) < 0 {
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
			nextDate, err := NextDate(time.Now().Format(layout), task.Date, task.Repeat)
			if err != nil {
				http.Error(w, `{"error":"Invalid repeat rule"}`, http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		} else {
			task.Date = time.Now().Format(layout)
		}
	}
	// Добавляем задачу в базу данных
	res, err := db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)", task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, `{"error":"Failed to add task"}`, http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"error":"Failed to retrieve task ID"}`, http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON-ответ с ID задачи
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": id,
	})
}

func MainHandle(res http.ResponseWriter, req *http.Request) {
	out := "Hello from server package!"
	res.Write([]byte(out))
}

func ApiNextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Парсим текущую дату
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Invalid 'now' date format", http.StatusBadRequest)
		return
	}

	// Вызываем функцию NextDate
	nextDate, err := NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Формируем ответ в JSON формате
	response := map[string]string{
		"nextDate": nextDate,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Порт, на котором будет работать сервер
const port = "7540"

// Директория для сервирования файлов
const webDir = "./web"

var db *sql.DB

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
func main() {
	var err error
	db, err = sql.Open("sqlite", "scheduler.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		install = true
	}

	if install {
		db, err := sql.Open("sqlite", "scheduler.db")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		// Создание таблицы scheduler
		createTableSQL := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT,
		title TEXT,
		comment TEXT,
		repeat TEXT
	);
	`

		_, err = db.Exec(createTableSQL)
		if err != nil {
			fmt.Println("Error creating table:", err)
			return
		}

		// Создание индекса по полю date
		createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`

		_, err = db.Exec(createIndexSQL)
		if err != nil {
			fmt.Println("Error creating index:", err)
			return
		}

		fmt.Println("Table and index created successfully")
	} else {
		fmt.Println("Database already exists")
	}

	// Создаем файловый сервер для директории web
	fs := http.FileServer(http.Dir(webDir))
	// Настраиваем обработчик для всех запросов
	http.Handle("/", fs)
	http.HandleFunc("/api/task", TaskHandler)
	http.HandleFunc("/api/nextdate", ApiNextDateHandler)
	// Запускаем сервер на указанном порту
	log.Printf("Starting server on :%s\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Порт, на котором будет работать сервер
const port = "7540"

// Директория для сервирования файлов
const webDir = "./web"

var db *sql.DB

func main() {
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
		createDatabase()
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

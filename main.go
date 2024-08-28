package main

import (
	"database/sql"
	"fmt"
	"github.com/MirekKrassilnikov/go_final_project/createDatabase"
	"github.com/MirekKrassilnikov/go_final_project/server"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"path/filepath"
)

// Порт, на котором будет работать сервер
const port = "7540"

// Директория для сервирования файлов
const webDir = "./web"
const layout = "20060102"

var db *sql.DB

type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func main() {
	db, err := sql.Open("sqlite", "scheduler.db")
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
		createDatabase.CreateDatabase()
	} else {
		fmt.Println("Database already exists")
	}

	// Создаем файловый сервер для директории web
	fs := http.FileServer(http.Dir(webDir))
	// Настраиваем обработчик для всех запросов
	http.Handle("/", fs)
	http.HandleFunc("/api/task", server.TaskHandler)
	http.HandleFunc("/api/nextdate", server.ApiNextDateHandler)
	// Запускаем сервер на указанном порту
	log.Printf("Starting server on :%s\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

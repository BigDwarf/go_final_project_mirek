package main

import (
	"net/http"
	"log"
)
// Порт, на котором будет работать сервер
const port = "7540"

// Директория для сервирования файлов
const webDir = "./web"

func main() {

	// Создаем файловый сервер для директории web
	fs := http.FileServer(http.Dir(webDir))

	// Настраиваем обработчик для всех запросов
	http.Handle("/", fs)

	// Запускаем сервер на указанном порту
	log.Printf("Starting server on :%s\n", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

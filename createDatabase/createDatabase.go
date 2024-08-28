package createDatabase

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
)

func CreateDatabase() {
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
}

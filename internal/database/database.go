package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func InitDB() *sql.DB {
	var err error

	DB, err := sql.Open("sqlite", "./internal/database/migrations/database.db")
	if err != nil {
		log.Fatal("Connect DB Sqlite Error: ", err)
	}

	query := `
		CREATE TABLE IF NOT EXISTS records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			item_id TEXT NOT NULL,
			name TEXT NOT NULL,
			folder_id TEXT NOT NULL,
			date DATE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS user_permission (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			permission_id TEXT NOT NULL,
			email TEXT NOT NULL
		);
	`

	if _, err := DB.Exec(query); err != nil {
		log.Fatal("Create Table Error: ", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal("Ping DB Sqlite Error: ", err)
	}

	fmt.Println("Connected to DB Sqlite")

	return DB
}

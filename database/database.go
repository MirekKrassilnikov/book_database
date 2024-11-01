package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // импортируем драйвер PostgreSQL
)

func createDatabase() {
	// Строка подключения к базе данных postgres
	connectStr := "user=mirekkrassilnikov dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connectStr) // Добавлена запятая между аргументами
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Создание базы данных book_database
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'book_database');").Scan(&exists)
	if err != nil {
		log.Fatalf("Ошибка при проверке существования базы данных: %v", err)
	}

	if !exists {
		_, err = db.Exec("CREATE DATABASE book_database;")
		if err != nil {
			log.Fatalf("Ошибка при создании базы данных: %v", err)
		}
		fmt.Println("База данных book_database создана успешно!")

		// Подключаемся к новой базе данных book_database
		connectStr = "user=mirekkrassilnikov dbname=book_database sslmode=disable"
		db, err = sql.Open("postgres", connectStr)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Создание таблицы rooms в базе данных book_database
		createRoomsTable := `
    CREATE TABLE IF NOT EXISTS rooms (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL
    );`

		_, err = db.Exec(createRoomsTable)
		if err != nil {
			log.Fatalf("Ошибка при создании таблицы rooms: %v", err)
		}
		fmt.Println("Таблица rooms создана успешно!")
	}
}

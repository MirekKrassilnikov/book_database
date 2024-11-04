package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // импортируем драйвер PostgreSQL
)

func CreateDatabase() *sql.DB {
	// Подключение к базе данных postgres для проверки и создания новой базы
	connectStr := "user=mirekkrassilnikov dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connectStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close() // Закрываем соединение после завершения работы функции

	// Проверка существования базы данных book_database
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'book_database');").Scan(&exists)
	if err != nil {
		log.Fatalf("Ошибка при проверке существования базы данных: %v", err)
	}

	if !exists {
		// Создание базы данных book_database
		_, err = db.Exec("CREATE DATABASE book_database;")
		if err != nil {
			log.Fatalf("Ошибка при создании базы данных: %v", err)
		}
		fmt.Println("База данных book_database создана успешно!")
	}

	// Подключаемся к новой базе данных book_database
	connectStr = "user=mirekkrassilnikov dbname=book_database sslmode=disable"
	db, err = sql.Open("postgres", connectStr)
	if err != nil {
		log.Fatal(err)
	}

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

	// Создание таблицы meetings
	createMeetingsTable := `
	CREATE TABLE IF NOT EXISTS meetings (
		id SERIAL PRIMARY KEY,
		room_id INT NOT NULL,
		topic VARCHAR(255) NOT NULL,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP NOT NULL,
		FOREIGN KEY (room_id) REFERENCES rooms(id),
		CHECK (start_time < end_time)
	);`

	_, err = db.Exec(createMeetingsTable)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы meetings: %v", err)
	}
	fmt.Println("Таблица meetings создана успешно!")

	var roomCount int
	err = db.QueryRow("SELECT COUNT(*) FROM rooms;").Scan(&roomCount)
	if err != nil {
		log.Fatalf("Ошибка при подсчете комнат: %v", err)
	}

	// Добавляем комнаты, если их меньше 6
	if roomCount < 6 {
		for i := roomCount + 1; i <= 6; i++ {
			roomName := fmt.Sprintf("Комната %d", i)
			_, err := db.Exec("INSERT INTO rooms (name) VALUES ($1);", roomName)
			if err != nil {
				log.Fatalf("Ошибка при добавлении комнаты %d: %v", i, err)
			}
			fmt.Printf("Комната %d добавлена успешно!\n", i)
		}
	} else {
		fmt.Println("В таблице rooms уже есть 6 или более комнат.")
	}
	rows, err := db.Query("SELECT id, name FROM rooms;")
	if err != nil {
		log.Fatalf("Ошибка при выполнении запроса: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("ID: %d, Name: %s\n", id, name)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	var dbName string
	err = db.QueryRow("SELECT current_database();").Scan(&dbName)
	if err != nil {
		log.Fatalf("Ошибка при получении текущей базы данных: %v", err)
	}
	fmt.Printf("Подключено к базе данных: %s\n", dbName)
	return db // Возвращаем открытое соединение
}

package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // регистрируем драйвер
)

// Глобальный экземпляр подключения к БД
var DB *sql.DB

// SQL-схема: создание таблицы и индекса
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT,
    repeat VARCHAR(128) DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init открывает БД и создаёт таблицу, если файла не существует
func Init(dbFile string) error {
	// Проверяем, существовал ли файл БД до открытия
	_, err := os.Stat(dbFile)
	fileExists := err == nil

	// Открываем (или создаём) файл БД
	// _loc=auto обеспечивает корректную работу с часовыми поясами
	DB, err = sql.Open("sqlite", fmt.Sprintf("%s?_loc=auto", dbFile))
	if err != nil {
		return fmt.Errorf("ошибка открытия БД: %w", err)
	}

	// Проверяем подключение
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Если файла не было — создаём схему
	if !fileExists {
		if _, err := DB.Exec(schema); err != nil {
			return fmt.Errorf("ошибка создания схемы БД: %w", err)
		}
	}

	return nil
}

// Close закрывает подключение к БД (вызывать при завершении сервера)
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
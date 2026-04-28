package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // регистрируем драйвер
)

// Store инкапсулирует подключение к БД и бизнес-логику
type Store struct {
	db *sql.DB
}

// NewStore создаёт новый экземпляр Store с инициализированным подключением
func NewStore(dbFile string) (*Store, error) {
	_, err := os.Stat(dbFile)
	fileExists := err == nil

	// Открываем подключение
	// _loc=auto обеспечивает корректную работу с часовыми поясами
	// _pragma=foreign_keys=on можно добавить при необходимости
	db, err := sql.Open("sqlite", fmt.Sprintf("%s?_loc=auto", dbFile))
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия БД: %w", err)
	}

	// Проверяем подключение и закрываем при ошибке
	// sql.Open не устанавливает соединение сразу — только при первом запросе
	if err := db.Ping(); err != nil {
		_ = db.Close() // освобождаем ресурсы, если пинг не прошёл
		return nil, fmt.Errorf("ошибка подключения к БД: %w", err)
	}

	// Создаём схему, если файла не было
	if !fileExists {
		if _, err := db.Exec(schema); err != nil {
			_ = db.Close() // закрываем при ошибке миграции
			return nil, fmt.Errorf("ошибка создания схемы БД: %w", err)
		}
	}

	return &Store{db: db}, nil
}

// Close закрывает подключение к БД (вызывать при завершении сервера)
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// DB возвращает raw-подключение для случаев, когда нужен прямой доступ к sql.DB
// Используйте с осторожностью: предпочтительнее методы самого Store
func (s *Store) DB() *sql.DB {
	return s.db
}

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
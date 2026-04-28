package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Форматы дат
const (
	DateInputFormat   = "02.01.2006"
	DateStorageFormat = "20060102"
)

// Task представляет задачу
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

// AddTask добавляет задачу
func (s *Store) AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	
	res, err := s.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, fmt.Errorf("ошибка добавления задачи: %w", err)
	}
	
	return res.LastInsertId()
}

// Tasks возвращает список задач
func (s *Store) Tasks(limit int, search string) ([]*Task, error) {
	baseQuery := `SELECT id, date, title, comment, repeat FROM scheduler`
	var query string
	var args []any

	if parsedDate, err := time.Parse(DateInputFormat, search); err == nil {
		dateStr := parsedDate.Format(DateStorageFormat)
		query = baseQuery + ` WHERE date = ? ORDER BY date ASC LIMIT ?`
		args = append(args, dateStr, limit)
	}

	if search != "" && query == "" {
		pattern := "%" + search + "%"
		query = baseQuery + ` WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?`
		args = append(args, pattern, pattern, limit)
	}

	if query == "" {
		query = baseQuery + ` WHERE date >= date('now') ORDER BY date ASC LIMIT ?`
		args = append(args, limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка итерации: %w", err)
	}

	return tasks, nil
}

// GetTask возвращает задачу по ID
func (s *Store) GetTask(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	
	var dbID int64
	var comment, repeat sql.NullString
	task := &Task{}
	
	err := s.db.QueryRow(query, id).Scan(&dbID, &task.Date, &task.Title, &comment, &repeat)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, fmt.Errorf("ошибка получения задачи: %w", err)
	}
	
	task.ID = fmt.Sprintf("%d", dbID)
	if comment.Valid {
		task.Comment = comment.String
	}
	if repeat.Valid {
		task.Repeat = repeat.String
	}
	
	return task, nil
}

// UpdateTask обновляет задачу
func (s *Store) UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	
	res, err := s.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return fmt.Errorf("ошибка обновления: %w", err)
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки результата: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}
	
	return nil
}

// DeleteTask удаляет задачу
func (s *Store) DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	
	res, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления: %w", err)
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}
	
	return nil
}

// UpdateDate обновляет дату задачи
func (s *Store) UpdateDate(next, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	
	res, err := s.db.Exec(query, next, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления даты: %w", err)
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка проверки: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}
	
	return nil
}

// scanTask — вспомогательная функция для сканирования строки
func scanTask(rows *sql.Rows) (*Task, error) {
	var task Task
	var id int64
	var comment, repeat sql.NullString

	if err := rows.Scan(&id, &task.Date, &task.Title, &comment, &repeat); err != nil {
		return nil, fmt.Errorf("ошибка сканирования: %w", err)
	}

	task.ID = fmt.Sprintf("%d", id)
	if comment.Valid {
		task.Comment = comment.String
	}
	if repeat.Valid {
		task.Repeat = repeat.String
	}
	
	return &task, nil
}
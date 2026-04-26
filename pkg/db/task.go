package db

import (
	"database/sql"
	"fmt"
	"time"
)

// Task представляет задачу в планировщике
type Task struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

// AddTask добавляет задачу в таблицу scheduler и возвращает её ID
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) 
	          VALUES (?, ?, ?, ?)`
	
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	
	return res.LastInsertId()
}
// Tasks возвращает список ближайших задач
// limit — максимальное количество записей
// search — опциональный поиск (пустая строка = без фильтра)
func Tasks(limit int, search string) ([]*Task, error) {
	var query string
	var args []any

	baseQuery := `SELECT id, date, title, comment, repeat FROM scheduler`

	if search == "" {
		// Без поиска: ближайшие задачи (дата >= сегодня)
		query = baseQuery + ` WHERE date >= date('now') ORDER BY date ASC LIMIT ?`
		args = append(args, limit)
	} else {
		// Проверяем, является ли search датой в формате 02.01.2006
		if parsedDate, err := time.Parse("02.01.2006", search); err == nil {
			// Поиск по конкретной дате
			dateStr := parsedDate.Format("20060102")
			query = baseQuery + ` WHERE date = ? ORDER BY date ASC LIMIT ?`
			args = append(args, dateStr, limit)
		} else {
			// Поиск подстроки в title или comment
			pattern := "%" + search + "%"
			query = baseQuery + ` WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?`
			args = append(args, pattern, pattern, limit)
		}
	}

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		var id int64
		var comment, repeat sql.NullString

		if err := rows.Scan(&id, &task.Date, &task.Title, &comment, &repeat); err != nil {
			return nil, err
		}

		task.ID = fmt.Sprintf("%d", id)
		if comment.Valid {
			task.Comment = comment.String
		}
		if repeat.Valid {
			task.Repeat = repeat.String
		}
		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}
// GetTask возвращает задачу по её идентификатору
func GetTask(id string) (*Task, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	
	var task Task
	var dbID int64
	var comment, repeat sql.NullString
	
	err := DB.QueryRow(query, id).Scan(&dbID, &task.Date, &task.Title, &comment, &repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("задача не найдена")
		}
		return nil, err
	}
	
	task.ID = fmt.Sprintf("%d", dbID)
	if comment.Valid {
		task.Comment = comment.String
	}
	if repeat.Valid {
		task.Repeat = repeat.String
	}
	
	return &task, nil
}
// UpdateTask обновляет существующую задачу в БД
func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	
	// Проверяем, что запись действительно была обновлена
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}
	
	return nil
}
// DeleteTask удаляет задачу по её идентификатору
func DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	
	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}
	
	// Проверяем, что запись действительно была удалена
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}
	
	return nil
}
func UpdateDate(next string, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	
	res, err := DB.Exec(query, next, id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("задача не найдена")
	}
	
	return nil
}
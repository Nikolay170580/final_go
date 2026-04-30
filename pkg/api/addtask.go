package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"final/pkg/db"
)

const (
	storageDateFormat = "20060102" // формат хранения дат в БД
	tasksDefaultLimit = 50         // лимит задач по умолчанию для GET 
)

// checkDate проверяет и нормализует дату задачи
func checkDate(task *db.Task) (string, error) {
	now := time.Now()
	
	if task.Date == "" {
		task.Date = now.Format(storageDateFormat)
		return "", nil
	}
	
	if _, err := time.Parse(storageDateFormat, task.Date); err != nil {
		return "", fmt.Errorf("неверный формат даты: %w", err)
	}
	
	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return "", fmt.Errorf("ошибка вычисления следующей даты: %w", err)
		}
		return next, nil
	}
	
	return "", nil
}

// isDateInPast возвращает true, если дата date раньше текущей даты
func isDateInPast(now, date time.Time) bool {
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return date.Before(nowDate)
}

// addTaskHandler обрабатывает POST  — создание новой задачи
func addTaskHandler(w http.ResponseWriter, r *http.Request, store *db.Store) {
	var task db.Task
	
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
		return
	}
	
	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}
	
	nextDate, err := checkDate(&task)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат даты или правила повторения"})
		return
	}
	
	now := time.Now()
	taskTime, _ := time.Parse(storageDateFormat, task.Date)
	if isDateInPast(now, taskTime) {
		if task.Repeat == "" {
			task.Date = now.Format(storageDateFormat)
		} else {
			task.Date = nextDate
		}
	}
	
	id, err := store.AddTask(&task)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка сохранения задачи"})
		return
	}
	
	writeJson(w, http.StatusCreated, map[string]string{"id": fmt.Sprintf("%d", id)})
}

// getTaskHandler обрабатывает GET  — получение задачи по ID
func getTaskHandler(w http.ResponseWriter, r *http.Request, store *db.Store) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}
	
	task, err := store.GetTask(id)
	if err != nil {
		writeJson(w, http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		return
	}
	
	writeJson(w, http.StatusOK, task)
}

// updateTaskHandler обрабатывает PUT  — обновление задачи
func updateTaskHandler(w http.ResponseWriter, r *http.Request, store *db.Store) {
	var task db.Task
	
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
		return
	}
	
	if task.ID == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор задачи"})
		return
	}
	
	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}
	
	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format(storageDateFormat)
	}
	
	taskTime, err := time.Parse(storageDateFormat, task.Date)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат даты"})
		return
	}
	
	if task.Repeat != "" {
		if _, err := NextDate(now, task.Date, task.Repeat); err != nil {
			writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат правила повторения"})
			return
		}
	}
	
	if isDateInPast(now, taskTime) {
		if task.Repeat == "" {
			task.Date = now.Format(storageDateFormat)
		} else {
			next, _ := NextDate(now, task.Date, task.Repeat)
			task.Date = next
		}
	}
	
	if err := store.UpdateTask(&task); err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления задачи"})
		return
	}
	
	writeJson(w, http.StatusOK, map[string]string{})
}

// doneTaskHandler обрабатывает POST  — завершение задачи
func doneTaskHandler(w http.ResponseWriter, r *http.Request, store *db.Store) {
	if r.Method != http.MethodPost {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}
	
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}
	
	task, err := store.GetTask(id)
	if err != nil {
		writeJson(w, http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		return
	}
	
	if task.Repeat == "" {
		if err := store.DeleteTask(id); err != nil {
			writeJson(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления задачи"})
			return
		}
		writeJson(w, http.StatusOK, map[string]string{})
		return
	}
	
	now := time.Now()
	nextDate, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат правила повторения"})
		return
	}
	
	if err := store.UpdateDate(nextDate, id); err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления даты"})
		return
	}
	
	writeJson(w, http.StatusOK, map[string]string{})
}

// deleteTaskHandler обрабатывает DELETE  — удаление задачи
func deleteTaskHandler(w http.ResponseWriter, r *http.Request, store *db.Store) {
	if r.Method != http.MethodDelete {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}
	
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}
	
	if err := store.DeleteTask(id); err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления задачи"})
		return
	}
	
	writeJson(w, http.StatusOK, map[string]string{})
}

// taskHandler маршрутизирует запросы по HTTP-методу для эндпоинта 
func taskHandler(w http.ResponseWriter, r *http.Request, store *db.Store) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r, store)
	case http.MethodGet:
		getTaskHandler(w, r, store)
	case http.MethodPut:
		updateTaskHandler(w, r, store)
	case http.MethodDelete:
		deleteTaskHandler(w, r, store)
	default:
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
	}
}
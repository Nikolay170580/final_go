package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"final/pkg/db" 
)


// writeJson отправляет ответ в формате JSON
func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// checkDate проверяет и нормализует дату задачи
// Возвращает следующую дату для повторения (если правило задано)
func checkDate(task *db.Task) (string, error) {
	now := time.Now()
	
	// Если дата не указана — берём сегодня
	if task.Date == "" {
		task.Date = now.Format("20060102")
		return "", nil
	}
	
	// Парсим дату
	_, err := time.Parse("20060102", task.Date)
	if err != nil {
		return "", err
	}
	
	// Если правило повторения указано — валидируем и получаем следующую дату
	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return "", err
		}
		return next, nil
	}
	
	return "", nil
}

// afterNow возвращает true, если дата date раньше текущей даты now
func afterNow(now, date time.Time) bool {
	// Сравниваем только даты, без времени
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return date.Before(nowDate)
}


// addTaskHandler обрабатывает POST /api/task
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	
	// 1. Десериализация JSON
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	
	// 2. Проверка обязательного поля
	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}
	
	// 3. Валидация даты и вычисление следующей даты при повторении
	nextDate, err := checkDate(&task)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат даты или правила повторения"})
		return
	}
	
	// 4. Корректировка даты, если она в прошлом
	now := time.Now()
	taskTime, _ := time.Parse("20060102", task.Date)
	if afterNow(now, taskTime) {
		if task.Repeat == "" {
			// Нет повторения → ставим сегодня
			task.Date = now.Format("20060102")
		} else {
			// Есть повторение → используем вычисленную следующую дату
			task.Date = nextDate
		}
	}
	
	// 5. Добавление в БД
	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	
	// 6. Успешный ответ
	writeJson(w, http.StatusCreated, map[string]string{"id": fmt.Sprintf("%d", id)})
}


// TasksResp — формат ответа со списком задач
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler обрабатывает GET /api/tasks
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем только GET
	if r.Method != http.MethodGet {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}

	// Параметр поиска (для бонуса)
	search := r.URL.Query().Get("search")

	// Получаем задачи из БД (лимит 50)
	tasks, err := db.Tasks(50, search)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Возвращаем пустой слайс, а не null
	if tasks == nil {
		tasks = []*db.Task{}
	}

	writeJson(w, http.StatusOK, TasksResp{Tasks: tasks})
}


// getTaskHandler обрабатывает GET 
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем id из query-параметра
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}
	
	// Получаем задачу из БД
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	
	// Возвращаем задачу
	writeJson(w, http.StatusOK, task)
}

// updateTaskHandler обрабатывает PUT 
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	
	// 1. Десериализация JSON
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	
	// 2. Проверка обязательного id
	if task.ID == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор задачи"})
		return
	}
	
	// 3. Проверка обязательного заголовка
	if task.Title == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан заголовок задачи"})
		return
	}
	
	// 4. Валидация даты (аналогично addTaskHandler)
	now := time.Now()
	if task.Date == "" {
		task.Date = now.Format("20060102")
	}
	
	taskTime, err := time.Parse("20060102", task.Date)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат даты"})
		return
	}
	
	// Валидация правила повторения
	if task.Repeat != "" {
		_, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат правила повторения"})
			return
		}
	}
	
	// 5. Корректировка даты, если она в прошлом
	if taskTime.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
		if task.Repeat == "" {
			task.Date = now.Format("20060102")
		} else {
			next, _ := NextDate(now, task.Date, task.Repeat)
			task.Date = next
		}
	}
	
	// 6. Обновление в БД
	if err := db.UpdateTask(&task); err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	
	// 7. Успешный ответ (пустой объект)
	writeJson(w, http.StatusOK, map[string]string{})
}

// doneTaskHandler обрабатывает POST 
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Только POST
	if r.Method != http.MethodPost {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}
	
	// Получаем id из query-параметра
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}
	
	// Получаем текущие данные задачи
	task, err := db.GetTask(id)
	if err != nil {
		writeJson(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	
	// Если правило повторения пустое — удаляем задачу
	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJson(w, http.StatusOK, map[string]string{})
		return
	}
	
	// Если задача периодическая — вычисляем следующую дату
	now := time.Now()
	nextDate, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Неверный формат правила повторения"})
		return
	}
	
	// Обновляем только дату в БД
	if err := db.UpdateDate(nextDate, id); err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	
	// Успешный ответ
	writeJson(w, http.StatusOK, map[string]string{})
}

// deleteTaskHandler обрабатывает DELETE 
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Только DELETE
	if r.Method != http.MethodDelete {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}
	
	// Получаем id из query-параметра
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		return
	}
	
	// Удаляем задачу из БД
	if err := db.DeleteTask(id); err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	
	// Успешный ответ
	writeJson(w, http.StatusOK, map[string]string{})
}

// taskHandler маршрутизирует запросы  по HTTP-методу
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
		
	case http.MethodGet:
		getTaskHandler(w, r)
		
	case http.MethodPut:
		updateTaskHandler(w, r)
		
	case http.MethodDelete:  
		deleteTaskHandler(w, r)
	
	default:
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
	}
}
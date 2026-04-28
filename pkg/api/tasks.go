package api

import (
	"net/http"

	"final/pkg/db"
)

// TasksResp — формат ответа со списком задач
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler обрабатывает GET /api/tasks
// Возвращает список задач с поддержкой поиска и лимита
func tasksHandler(w http.ResponseWriter, r *http.Request, store *db.Store) {
	if r.Method != http.MethodGet {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}

	search := r.URL.Query().Get("search")

	// Используем константу вместо магического числа 50
	tasks, err := store.Tasks(tasksDefaultLimit, search)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "Ошибка получения задач"})
		return
	}

	if tasks == nil {
		tasks = []*db.Task{}
	}

	writeJson(w, http.StatusOK, TasksResp{Tasks: tasks})
}
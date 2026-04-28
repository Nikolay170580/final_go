package api

import (
	"encoding/json"
	"net/http"
)

// writeJson отправляет ответ в формате JSON
func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Логируем ошибку, но не ломаем ответ
		// В продакшене лучше использовать logger
	}
}
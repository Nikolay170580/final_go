package api

import (
	"encoding/json"
	"net/http"
)

// writeJson отправляет ответ в формате JSON
func writeJson(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
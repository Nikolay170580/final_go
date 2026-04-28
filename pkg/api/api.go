package api

import (
	"net/http"
	"final/pkg/db"
	"final/pkg/server/config"
)

func Init(cfg *config.Config, store *db.Store) {
	// Публичные эндпоинты
	http.HandleFunc("/api/signin", func(w http.ResponseWriter, r *http.Request) {
		signinHandler(w, r, cfg)
	})
	http.HandleFunc("/api/nextdate", NextDateHandler)

	// Защищённые эндпоинты
	http.HandleFunc("/api/task", AuthMiddleware(cfg)(func(w http.ResponseWriter, r *http.Request) {
		taskHandler(w, r, store)
	}))
	http.HandleFunc("/api/tasks", AuthMiddleware(cfg)(func(w http.ResponseWriter, r *http.Request) {
		tasksHandler(w, r, store)
	}))
	http.HandleFunc("/api/task/done", AuthMiddleware(cfg)(func(w http.ResponseWriter, r *http.Request) {
		doneTaskHandler(w, r, store)
	}))
}
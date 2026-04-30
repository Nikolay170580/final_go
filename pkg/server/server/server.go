package server

import (
	"log"
	"net/http"
	"final/pkg/server/config"
)

func Start(cfg *config.Config) error {
	// Раздача статики из директории web
	fs := http.FileServer(http.Dir(cfg.WebDir))
	http.Handle("/", fs)

	addr := ":" + cfg.Port
	log.Printf("Сервер запущен на порту %s, раздает файлы из %s", cfg.Port, cfg.WebDir)

	return http.ListenAndServe(addr, nil)
}
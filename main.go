package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"time"
	
	"final/pkg/api"
	"final/pkg/server/config"
	"final/pkg/db"
)

func main() {
	cfg := config.Load()

	// Инициализация БД
	if err := db.Init(cfg.DBFile); err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer db.Close()

	//Регистрируем API-обработчики
	api.Init()

	// Раздача статики
	 fs := http.FileServer(http.Dir(cfg.WebDir))
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Пропускаем все /api/* — их уже обработал api.Init()
        if strings.HasPrefix(r.URL.Path, "/api/") {
            http.NotFound(w, r)
            return
        }
        fs.ServeHTTP(w, r)
    })


	// Создаём http.Server для контроля завершения
	srv := &http.Server{
		Addr: ":" + cfg.Port,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Printf("Сервер запущен: http://localhost:%s", cfg.Port)
		log.Printf("БД: %s", cfg.DBFile)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Получен сигнал завершения, останавливаем сервер...")

	// Корректное завершение с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке: %v", err)
	}

	log.Println(" Сервер остановлен")
}
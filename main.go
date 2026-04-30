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
	// Загружаем конфигурацию ОДИН раз при старте
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Инициализация БД через NewStore (вместо db.Init)
	store, err := db.NewStore(cfg.DBFile)
	if err != nil {
		log.Fatalf("Ошибка инициализации БД: %v", err)
	}
	defer store.Close() // гарантированно закрываем при выходе

	// Передаём и конфиг, и store в Init
	api.Init(cfg, store)

	// Раздача статики
	fs := http.FileServer(http.Dir(cfg.WebDir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	})

	// Подписка на сигналы (до запуска сервера)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	errChan := make(chan error, 1)

	srv := &http.Server{
		Addr: ":" + cfg.Port,
	}

	// Запуск сервера в горутине
	go func() {
		log.Printf("Сервер запущен: http://localhost:%s", cfg.Port)
		log.Printf(" БД: %s", cfg.DBFile)
		if cfg.TodoPassword == "" {
			log.Println(" Аутентификация ОТКЛЮЧЕНА (пустой TODO_PASSWORD)")
		}
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка сервера: %v", err)
			errChan <- err
		}
	}()

	// Ждём ИЛИ сигнал ОС, ИЛИ ошибку сервера
	select {
	case sig := <-quit:
		log.Printf("Получен сигнал завершения: %v", sig)
	case err := <-errChan:
		log.Printf("Сервер завершился с ошибкой: %v", err)
	}

	log.Println("Остановка сервера...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()  
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке: %v", err)
	}

	log.Println("Сервер остановлен")
}
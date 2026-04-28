package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Сервер
	Port   string
	WebDir string
	
	// База данных
	DBFile string
	
	// Аутентификация
	JWTSecret    []byte
	TodoPassword string
	TokenTTL     time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Порт 
	cfg.Port = "7540"
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p <= 65535 {
			cfg.Port = envPort
		}
	}

	// Web-директория 
	cfg.WebDir = "web"
	if envDir := os.Getenv("TODO_WEB_DIR"); envDir != "" {
		cfg.WebDir = envDir
	}

	// Файл БД 
	cfg.DBFile = "scheduler.db"
	if envDB := os.Getenv("TODO_DBFILE"); envDB != "" {
		cfg.DBFile = envDB
	}

	//  Аутентификация: пароль 
	cfg.TodoPassword = os.Getenv("TODO_PASSWORD")
	authEnabled := cfg.TodoPassword != ""

	// JWT Secret: требуется ТОЛЬКО если включена аутентификация
	if authEnabled {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			return nil, fmt.Errorf("ошибка загрузки конфигурации: при включённой аутентификации требуется JWT_SECRET")
		}
		cfg.JWTSecret = []byte(jwtSecret)
	}
	// Если аутентификация отключена: JWTSecret остаётся пустым []byte{}
	// Middleware должен пропускать запросы без проверки, если TodoPassword == ""

	// TTL токена 
	cfg.TokenTTL = 8 * time.Hour
	if envTTL := os.Getenv("TOKEN_TTL_HOURS"); envTTL != "" {
		if h, err := strconv.Atoi(envTTL); err == nil && h > 0 {
			cfg.TokenTTL = time.Duration(h) * time.Hour
		}
	}

	return cfg, nil
}
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port   string
	WebDir string
	DBFile string
}

func Load() *Config {
	// Порт
	port := "7540"
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p <= 65535 {
			port = envPort
		}
	}

	// Директория со статикой
	webDir := "web"
	if envDir := os.Getenv("TODO_WEB_DIR"); envDir != "" {
		webDir = envDir
	}

	// Файл БД
	dbFile := "scheduler.db"
	if envDB := os.Getenv("TODO_DBFILE"); envDB != "" {
		dbFile = envDB
	}

	return &Config{
		Port:   port,
		WebDir: webDir,
		DBFile: dbFile,
	}
}
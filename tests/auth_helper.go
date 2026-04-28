package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
)

type signInReq struct {
	Password string `json:"password"`
}

type signInResp struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// GetAuthToken выполняет вход через /api/signin и возвращает JWT-токен
func GetAuthToken(t *testing.T) string {
	password := os.Getenv("TODO_PASSWORD")
	if password == "" {
		// Аутентификация отключена — токен не требуется
		return ""
	}

	body, err := json.Marshal(signInReq{Password: password})
	if err != nil {
		t.Fatalf("Ошибка сериализации запроса авторизации: %v", err)
	}

	resp, err := http.Post(getURL("api/signin"), "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Ошибка HTTP-запроса к /api/signin: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Ошибка чтения ответа /api/signin: %v", err)
	}

	var result signInResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		t.Fatalf("Ошибка парсинга JSON ответа /api/signin: %v", err)
	}

	if result.Error != "" {
		t.Fatalf("Ошибка авторизации: %s (статус: %d)", result.Error, resp.StatusCode)
	}

	if result.Token == "" {
		t.Fatal("Пустой токен в ответе /api/signin")
	}

	return result.Token
}

// EnsureToken инициализирует глобальный Token, если аутентификация включена
func EnsureToken(t *testing.T) {
	if Token == "" && os.Getenv("TODO_PASSWORD") != "" {
		Token = GetAuthToken(t)
	}
}
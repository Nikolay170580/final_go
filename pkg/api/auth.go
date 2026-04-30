package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"final/pkg/server/config"
)

type Claims struct {
	PasswordHash string `json:"phash"`
	jwt.RegisteredClaims
}

type SignInRequest struct {
	Password string `json:"password"`
}

type SignInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func generateToken(password string, cfg *config.Config) (string, error) {
	if len(cfg.JWTSecret) == 0 {
		return "", errors.New("JWT_SECRET не настроен")
	}

	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	claims := Claims{
		PasswordHash: hashStr,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.TokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(cfg.JWTSecret)
}

func validateToken(tokenStr, currentPassword string, cfg *config.Config) bool {
	if currentPassword == "" {
		return true
	}
	if len(cfg.JWTSecret) == 0 {
		return false
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неподдерживаемый метод подписи")
		}
		return cfg.JWTSecret, nil
	})
	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return false
	}

	currentHash := sha256.Sum256([]byte(currentPassword))
	currentHashStr := hex.EncodeToString(currentHash[:])
	return claims.PasswordHash == currentHashStr
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	cookie, err := r.Cookie("token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func signinHandler(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	if r.Method != http.MethodPost {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Ошибка чтения запроса"})
		return
	}

	if req.Password != cfg.TodoPassword {
		writeJson(w, http.StatusUnauthorized, SignInResponse{Error: "Неверный пароль"})
		return
	}

	token, err := generateToken(req.Password, cfg)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, SignInResponse{Error: "Ошибка генерации токена"})
		return
	}

	writeJson(w, http.StatusOK, SignInResponse{Token: token})
}

func AuthMiddleware(cfg *config.Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if cfg.TodoPassword == "" {
				next(w, r)
				return
			}

			tokenStr := extractToken(r)
			if tokenStr == "" {
				writeJson(w, http.StatusUnauthorized, map[string]string{"error": "Требуется аутентификация"})
				return
			}

			if !validateToken(tokenStr, cfg.TodoPassword, cfg) {
				writeJson(w, http.StatusUnauthorized, map[string]string{"error": "Неверный токен"})
				return
			}

			next(w, r)
		}
	}
}
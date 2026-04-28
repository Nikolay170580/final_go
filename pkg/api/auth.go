package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"final/pkg/server/config"
)

// Claims — полезная нагрузка JWT-токена
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

// generateToken создаёт JWT
func generateToken(password string, cfg *config.Config) (string, error) {
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

// validateToken проверяет токен
func validateToken(tokenStr, currentPassword string, cfg *config.Config) bool {
	if currentPassword == "" {
		return true
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
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

// extractToken извлекает токен из заголовка или куки
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

// signinHandler — обработка входа
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

// AuthMiddleware — фабрика мидлвара
func AuthMiddleware(cfg *config.Config) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if cfg.TodoPassword == "" {
				next(w, r)
				return
			}

			tokenStr := extractToken(r)
			if tokenStr == "" {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !validateToken(tokenStr, cfg.TodoPassword, cfg) {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}

			next(w, r)
		}
	}
}
package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims — полезная нагрузка JWT-токена
type Claims struct {
	PasswordHash string `json:"phash"` // хэш пароля для валидации
	jwt.RegisteredClaims
}

// Secret key для подписи токенов (в продакшене хранить в env!)
var jwtSecret = []byte("change-me-in-production")

// SignInRequest — формат запроса на вход
type SignInRequest struct {
	Password string `json:"password"`
}

// SignInResponse — формат ответа
type SignInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

// generateToken создаёт JWT-токен с хэшем пароля
func generateToken(password string) (string, error) {
	// Создаём хэш пароля (SHA-256)
	hash := sha256.Sum256([]byte(password))
	hashStr := hex.EncodeToString(hash[:])

	claims := Claims{
		PasswordHash: hashStr,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// validateToken проверяет JWT-токен и сверяет хэш пароля
func validateToken(tokenStr, currentPassword string) bool {
	// Если пароль не задан — аутентификация не требуется
	if currentPassword == "" {
		return true
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return false
	}

	// Сверяем хэш текущего пароля с хэшем в токене
	currentHash := sha256.Sum256([]byte(currentPassword))
	currentHashStr := hex.EncodeToString(currentHash[:])

	return claims.PasswordHash == currentHashStr
}

// signinHandler обрабатывает POST 
func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJson(w, http.StatusMethodNotAllowed, map[string]string{"error": "Метод не поддерживается"})
		return
	}

	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJson(w, http.StatusBadRequest, map[string]string{"error": "Ошибка чтения запроса"})
		return
	}

	// Получаем правильный пароль из окружения
	correctPassword := os.Getenv("TODO_PASSWORD")

	// Сравниваем
	if req.Password != correctPassword {
		writeJson(w, http.StatusUnauthorized, SignInResponse{Error: "Неверный пароль"})
		return
	}

	// Генерируем токен
	token, err := generateToken(req.Password)
	if err != nil {
		writeJson(w, http.StatusInternalServerError, SignInResponse{Error: "Ошибка генерации токена"})
		return
	}

	writeJson(w, http.StatusOK, SignInResponse{Token: token})
}
// Auth — middleware для проверки аутентификации
func Auth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        correctPassword := os.Getenv("TODO_PASSWORD")
        if correctPassword == "" {
            next(w, r)
            return
        }

        // 1. Пробуем взять токен из заголовка: Authorization: Bearer <token>
        authHeader := r.Header.Get("Authorization")
        var tokenStr string
        
        if strings.HasPrefix(authHeader, "Bearer ") {
            tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
        } else {
            // 2. Или из куки (для обратной совместимости)
            cookie, err := r.Cookie("token")
            if err != nil {
                http.Error(w, "Authentication required", http.StatusUnauthorized)
                return
            }
            tokenStr = cookie.Value
        }

        if !validateToken(tokenStr, correctPassword) {
            http.Error(w, "Authentication required", http.StatusUnauthorized)
            return
        }

        next(w, r)
    }
}
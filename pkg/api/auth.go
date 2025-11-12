package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	applog "github.com/ljnsur/todosay/pkg/log"
)

var jwtKey []byte

type Claims struct {
	PasswordHash string `json:"phash"`
	jwt.RegisteredClaims
}

// Инициализация jwtKey при старте сервера
func InitAuth() {
	pass := os.Getenv("TODO_PASSWORD")
	if pass != "" {
		jwtKey = []byte(pass)
		applog.Printf("InitAuth: JWT-ключ инициализирован")
	} else {
		applog.Printf("InitAuth: TODO_PASSWORD не задан, аутентификация отключена")
	}
}

// Обработчик входа
func signInHandler(w http.ResponseWriter, r *http.Request) {
	applog.Printf("signIn: начало, method=%s remote=%s", r.Method, r.RemoteAddr)
	if r.Method != http.MethodPost {
		applog.Printf("signIn: неверный метод %s", r.Method)
		http.Error(w, "разрешены только POST запросы", http.StatusMethodNotAllowed)
		return
	}

	pass := os.Getenv("TODO_PASSWORD")
	if pass == "" {
		// Просто пускаем (для тестов)
		applog.Printf("signIn: TODO_PASSWORD пуст — выдаётся пустой токен для %s", r.RemoteAddr)
		writeJson(w, http.StatusOK, map[string]string{"token": ""})
		return
	}

	var req struct {
		Password string `json:"password"`
	}

	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		applog.Printf("signIn: ошибка чтения тела запроса: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(buf.Bytes(), &req); err != nil {
		applog.Printf("signIn: неверный JSON: %v (body=%q)", err, buf.String())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Password != pass {
		applog.Printf("signIn: неверный пароль, попытка с %s", r.RemoteAddr)
		writeJson(w, http.StatusUnauthorized, map[string]string{"error": "Неверный пароль"})
		return
	}

	// Создаём токен
	expirationTime := time.Now().Add(8 * time.Hour)
	claims := &Claims{
		PasswordHash: pass,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		applog.Printf("signIn: не удалось сгенерировать токен: %v", err)
		writeJson(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		return
	}

	applog.Printf("signIn: токен выдан для %s, срок до %s", r.RemoteAddr, expirationTime.Format(time.RFC3339))
	writeJson(w, http.StatusOK, map[string]string{"token": tokenString})
}

// Middleware для проверки JWT
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		applog.Printf("auth: начало проверки авторизации для %s %s", r.Method, r.RemoteAddr)
		pass := os.Getenv("TODO_PASSWORD")
		if pass == "" {
			// Если пароль не задан — пропускаем без проверки
			applog.Printf("auth: TODO_PASSWORD пуст — пропуск проверки для %s", r.RemoteAddr)
			next(w, r)
			return
		}

		cookie, err := r.Cookie("token")
		if err != nil {
			applog.Printf("auth: отсутствует cookie token от %s: %v", r.RemoteAddr, err)
			writeJson(w, http.StatusUnauthorized, map[string]string{"error": "требуется аутентификация"})
			return
		}

		tokenStr := cookie.Value

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			applog.Printf("auth: ошибка парсинга токена от %s: %v", r.RemoteAddr, err)
			writeJson(w, http.StatusUnauthorized, map[string]string{"error": "невалидный токен"})
			return
		}

		if !token.Valid {
			applog.Printf("auth: невалидный токен от %s", r.RemoteAddr)
			writeJson(w, http.StatusUnauthorized, map[string]string{"error": "невалидный токен"})
			return
		}

		// Важно: если пароль изменился — старый токен должен стать невалидным
		if claims.PasswordHash != pass {
			applog.Printf("auth: несоответствие пароля в токене от %s", r.RemoteAddr)
			writeJson(w, http.StatusUnauthorized, map[string]string{"error": "невалидный токен"})
			return
		}

		applog.Printf("auth: аутентификация успешно пройдена для %s", r.RemoteAddr)
		next(w, r)
	}
}

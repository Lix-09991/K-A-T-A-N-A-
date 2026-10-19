package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Lix-09991/neon-saga-api/internal/models"
)

type contextKey string

const UserContextKey contextKey = "user"

// RequireAuth — проверяет JWT в заголовке Authorization: Bearer <token>
func RequireAuth(jwtSecret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Пробуем Authorization header
			tokenStr := ""
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				// 2. Fallback на cookie
				cookie, err := r.Cookie("token")
				if err == nil {
					tokenStr = cookie.Value
				}
			}

			if tokenStr == "" {
				writeJSONError(w, http.StatusUnauthorized, "требуется авторизация")
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				writeJSONError(w, http.StatusUnauthorized, "невалидный токен")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "невалидные claims")
				return
			}

			sub, _ := claims["sub"].(string)
			userID, err := uuid.Parse(sub)
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, "невалидный sub")
				return
			}

			user := &models.User{
				ID:       userID,
				Email:    asString(claims["email"]),
				Username: asString(claims["username"]),
				Role:     models.Role(asString(claims["role"])),
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext — достаёт юзера из контекста
func UserFromContext(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(UserContextKey).(*models.User)
	return u, ok
}

func asString(v interface{}) string {
	s, _ := v.(string)
	return s
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

package middleware

import (
	"net/http"

	"github.com/Lix-09991/neon-saga-api/internal/models"
)

// RequireRole — пропускает только пользователей с указанной ролью.
// Использовать ПОСЛЕ RequireAuth.
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	allowed := make(map[models.Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok {
				writeJSONError(w, http.StatusUnauthorized, "требуется авторизация")
				return
			}
			if _, ok := allowed[user.Role]; !ok {
				writeJSONError(w, http.StatusForbidden, "недостаточно прав")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Solomiam356/witness-backend/internal/auth"
	"github.com/Solomiam356/witness-backend/internal/database"
)

type contextKey string

const (
UserIDKey contextKey = "userID"
UserRoleKey contextKey = "userRole"
)
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Відсутній токен авторизації", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Некоректний формат заголовка Authorization (має бути Bearer <token>)", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		claims, err := auth.IsTokenValid(tokenStr)
		if err != nil {
			http.Error(w, "Невалідний або прострочений токен", http.StatusUnauthorized)
			return
		}

		var hasActiveSession bool

		query := `SELECT EXISTS(SELECT 1 FROM sessions WHERE user_id = $1 AND revoked_at IS NULL)`

		err = database.DB.QueryRow(r.Context(), query, claims.UserID).Scan(&hasActiveSession)
		if err != nil {
			http.Error(w, "Помилка перевірки безпеки сесії: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if !hasActiveSession {
			http.Error(w, "Сесія анульована. Будь ласка, увійдіть в систему знову", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				http.Error(w, "Доступ заборонено: роль не визначено", http.StatusForbidden)
				return
			}

			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Доступ заборонено: недостатньо прав", http.StatusForbidden)
		})

	}
}

func RequireVerifiedEmail(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(UserIDKey).(string)
		if !ok || userID == "" {
			http.Error(w, "Неавторизований доступ", http.StatusUnauthorized)
			return
		}

		var emailVerified bool
		query := `SELECT email_verified FROM users WHERE id = $1 AND deleted_at IS NULL`

		err := database.DB.QueryRow(context.Background(), query, userID).Scan(&emailVerified)
		if err != nil {
			http.Error(w, "Користувача не знайдено", http.StatusUnauthorized)
			return
		}

		if !emailVerified {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error": "Будь ласка, підтвердіть вашу електронну пошту, щоб отримати доступ до цієї дії."}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
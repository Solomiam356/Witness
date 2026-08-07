package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Solomiam356/witness-backend/internal/auth"
	"github.com/Solomiam356/witness-backend/internal/database"
	"github.com/Solomiam356/witness-backend/internal/middleware"
	"github.com/Solomiam356/witness-backend/internal/service"
)

type AuthHandler struct {
	sessionSrv *service.SessionService
}

func NewAuthHandler(sessionSrv *service.SessionService) *AuthHandler {
	return &AuthHandler{sessionSrv: sessionSrv}
}

func (h *AuthHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req auth.SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON запит", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		http.Error(w, "Усі поля (email, password, display_name) є обов'язковими", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		http.Error(w, "Пароль має бути не менше як 6 символів", http.StatusBadRequest)
		return
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Помилка при обробці пароля", http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO users (email, password_hash, display_name, role, email_verified)
		VALUES ($1, $2, $3, 'user', false)
		RETURNING id
	`

	var userID string
	err = database.DB.QueryRow(context.Background(), query, req.Email, hashedPassword, req.DisplayName).Scan(&userID)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			http.Error(w, "Користувач з таким Email вже існує", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Помилка збереження в базу: %v", err), http.StatusInternalServerError)
		return
	}

	rawToken, err := auth.GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Помилка генерації токена верифікації", http.StatusInternalServerError)
		return
	}

	tokenHash := auth.HashToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	tokenQuery := `
	INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3)
	`

	_, err = database.DB.Exec(context.Background(), tokenQuery, userID, tokenHash, expiresAt)
	if err != nil {
		http.Error(w, "Помилка збереження токена верифікації", http.StatusInternalServerError)
		return
	}

	go func() {
		emailSvc := service.NewEmailService(os.Getenv("RESEND_API_KEY"))
		if emailErr := emailSvc.SendVerificationEmail(req.Email, rawToken); emailErr != nil {
			fmt.Printf("Помилка відправки листа верифікації: %v\n", emailErr)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Реєстрація пройшла успішно! Перевірте вашу пошту для активації акаунту.",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON запит", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	var dbUserID string
	var dbPasswordHash string
	var dbRole string

	query := "SELECT id, password_hash, role FROM users WHERE email = $1 AND deleted_at IS NULL"
	err := database.DB.QueryRow(context.Background(), query, req.Email).Scan(&dbUserID, &dbPasswordHash, &dbRole)
	if err != nil {
		http.Error(w, "Неправильний email або password", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPasswordHash(req.Password, dbPasswordHash) {
		http.Error(w, "Неправильний email або password", http.StatusUnauthorized)
		return
	}

	accessToken, err := auth.GenerateToken(dbUserID, dbRole)
	if err != nil {
		http.Error(w, "Помилка генерації токена безпеки", http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Помилка генерації токена оновлення", http.StatusInternalServerError)
		return
	}

	rfHash := auth.HashToken(refreshToken)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	deviceInfo := r.UserAgent()

	sessionQuery := `
		INSERT INTO sessions (user_id, refresh_token_hash, device_info, expires_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = database.DB.Exec(context.Background(), sessionQuery, dbUserID, rfHash, deviceInfo, expiresAt)
	if err != nil {
		http.Error(w, "Помилка збереження сесії", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"message":       "Вхід виконано успішно!",
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req auth.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON", http.StatusBadRequest)
		return
	}

	rfHash := auth.HashToken(req.RefreshToken)

	var userID string
	var expiresAt time.Time
	var revokedAt *time.Time
	var userRole string

	query := `
		SELECT s.user_id, s.expires_at, s.revoked_at, u.role
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.refresh_token_hash = $1
	`
	err := database.DB.QueryRow(context.Background(), query, rfHash).Scan(&userID, &expiresAt, &revokedAt, &userRole)
	if err != nil {
		http.Error(w, "Невалідний або відкликаний токен", http.StatusUnauthorized)
		return
	}

	if revokedAt != nil || time.Now().After(expiresAt) {
		http.Error(w, "Токен оновлення застарів або був анульований", http.StatusUnauthorized)
		return
	}

	newAccessToken, err := auth.GenerateToken(userID, userRole)
	if err != nil {
		http.Error(w, "Помилка генерації токена", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		http.Error(w, "Відсутній токен верифікації", http.StatusBadRequest)
		return
	}

	tokenHash := auth.HashToken(rawToken)

	var userID string
	var expiresAt time.Time

	query := `
	SELECT user_id, expires_at
	FROM email_verification_tokens
	WHERE token_hash = $1
	`

	err := database.DB.QueryRow(context.Background(), query, tokenHash).Scan(&userID, &expiresAt)
	if err != nil {
		http.Error(w, "Недійсний або вже використаний токен верифікації", http.StatusUnauthorized)
		return
	}

	if time.Now().After(expiresAt) {
		http.Error(w, "Термін дії токена закінчився. Запросіть новий лист.", http.StatusUnauthorized)
		return
	}

	updateUserQuery := `UPDATE users SET email_verified = true WHERE id = $1`
	_, err = database.DB.Exec(context.Background(), updateUserQuery, userID)
	if err != nil {
		http.Error(w, "Помилка оновлення статусу користувача", http.StatusInternalServerError)
		return
	}

	deleteTokenQuery := `DELETE FROM email_verification_tokens WHERE token_hash = $1`
	_, _ = database.DB.Exec(context.Background(), deleteTokenQuery, tokenHash)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Електронну пошту успішно підтверджено! Тепер ви можете увійти в акаунт.",
	})
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req auth.ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON запит", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		http.Error(w, "Email є обов'язковим", http.StatusBadRequest)
		return
	}

	var userID string
	var emailVerified bool
	userQuery := `SELECT id, email_verified FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := database.DB.QueryRow(context.Background(), userQuery, req.Email).Scan(&userID, &emailVerified)
	if err != nil {
		http.Error(w, "Користувача з таким email не знайдено", http.StatusNotFound)
		return
	}

	if emailVerified {
		http.Error(w, "Ця електронна адреса вже підтверджена", http.StatusBadRequest)
		return
	}

	_, _ = database.DB.Exec(context.Background(), `DELETE FROM email_verification_tokens WHERE user_id = $1`, userID)

	rawToken, err := auth.GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Помилка генерації нового токена", http.StatusInternalServerError)
		return
	}

	tokenHash := auth.HashToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	tokenQuery := `
	INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
	VALUES ($1, $2, $3)
	`

	_, err = database.DB.Exec(context.Background(), tokenQuery, userID, tokenHash, expiresAt)
	if err != nil {
		http.Error(w, "Помилка збереження токена", http.StatusInternalServerError)
		return
	}

	go func() {
		emailSvc := service.NewEmailService(os.Getenv("RESEND_API_KEY"))
		if emailErr := emailSvc.SendVerificationEmail(req.Email, rawToken); emailErr != nil {
			fmt.Printf("Помилка повторної відправки листа: %v\n", emailErr)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Новий лист для підтвердження успішно надіслано на вашу пошту!",
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req auth.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON запит", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		http.Error(w, "Email є обов'язковим", http.StatusBadRequest)
		return
	}

	var userID string
	userQuery := `SELECT id FROM users WHERE email = $1 AND deleted_at IS NULL`
	err := database.DB.QueryRow(context.Background(), userQuery, req.Email).Scan(&userID)

	if err == nil {
		_, _ = database.DB.Exec(context.Background(), `DELETE FROM password_reset_tokens WHERE user_id = $1`, userID)

		rawToken, genErr := auth.GenerateRefreshToken()
		if genErr == nil {
			tokenHash := auth.HashToken(rawToken)
			expiresAt := time.Now().Add(15 * time.Minute)

			tokenQuery := `
				INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
				VALUES ($1, $2, $3)
			`

			_, _ = database.DB.Exec(context.Background(), tokenQuery, userID, tokenHash, expiresAt)

			go func() {
				fmt.Printf("🔒 СИРИЙ ТОКЕН СКИДАННЯ для %s: %s\n", req.Email, rawToken)
			}()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Якщо такий email зареєстрований, інструкції зі скидання пароля були надіслані.",
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req auth.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON запит", http.StatusBadRequest)
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	if req.Token == "" || req.NewPassword == "" {
		http.Error(w, "Токен та новий пароль є обов'язковими", http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 6 {
		http.Error(w, "Пароль має бути не менше як 6 символів", http.StatusBadRequest)
		return
	}

	tokenHash := auth.HashToken(req.Token)

	var userID string
	var expiresAt time.Time

	query := `
		SELECT user_id, expires_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`

	err := database.DB.QueryRow(context.Background(), query, tokenHash).Scan(&userID, &expiresAt)
	if err != nil {
		http.Error(w, "Недійсний або вже використаний токен скидання", http.StatusUnauthorized)
		return
	}

	if time.Now().After(expiresAt) {
		http.Error(w, "Термін дії токена закінчився. Запросіть новий.", http.StatusUnauthorized)
		return
	}

	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Помилка обробки нового пароля", http.StatusInternalServerError)
		return
	}

	updateQuery := `UPDATE users SET password_hash = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err = database.DB.Exec(context.Background(), updateQuery, hashedPassword, userID)
	if err != nil {
		http.Error(w, "Помилка оновлення пароля", http.StatusInternalServerError)
		return
	}

	_, _ = database.DB.Exec(context.Background(), `DELETE FROM password_reset_tokens WHERE token_hash = $1`, tokenHash)
	_, _ = database.DB.Exec(context.Background(), `UPDATE sessions SET revoked_at = CURRENT_TIMESTAMP WHERE user_id = $1 AND revoked_at IS NULL`, userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Пароль успішно змінено! Тепер ви можете увійти з новим паролем.",
	})
}

func (h *AuthHandler) GetMySessions(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Користувач не авторизований", http.StatusUnauthorized)
		return
	}

	sessions, err := h.sessionSrv.GetUserSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, "Помилка при отриманні сесій", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}
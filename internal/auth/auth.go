package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Solomiam356/witness-backend/internal/database"
	"github.com/Solomiam356/witness-backend/internal/service"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("witness_super_secret_key_2026")

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string, role string) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute)

	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

type SignUpRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req SignUpRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
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

	hashedPassword, err := HashPassword(req.Password)
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

	rawToken, err := GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Помилка генерації токена верифікації", http.StatusInternalServerError)
		return
	}

	tokenHash := HashToken(rawToken)
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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Некоректний JSON запит", http.StatusBadRequest)
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	var dbUserID string
	var dbPasswordHash string
	var dbRole string

	query := "SELECT id, password_hash, role FROM users WHERE email = $1 AND deleted_at IS NULL"
	err = database.DB.QueryRow(context.Background(), query, req.Email).Scan(&dbUserID, &dbPasswordHash, &dbRole)
	if err != nil {
		http.Error(w, "Неправильний email або password", http.StatusUnauthorized)
		return
	}

	if !CheckPasswordHash(req.Password, dbPasswordHash) {
		http.Error(w, "Неправильний email або password", http.StatusUnauthorized)
		return
	}

	accessToken, err := GenerateToken(dbUserID, dbRole)
	if err != nil {
		http.Error(w, "Помилка генерації токена безпеки", http.StatusInternalServerError)
		return
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Помилка генерації токена оновлення", http.StatusInternalServerError)
		return
	}

	rfHash := HashToken(refreshToken)
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

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некоректний JSON", http.StatusBadRequest)
		return
	}

	rfHash := HashToken(req.RefreshToken)

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

	newAccessToken, err := GenerateToken(userID, userRole)
	if err != nil {
		http.Error(w, "Помилка генерації токена", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})
}

func IsTokenValid(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неочікуваний метод підпису: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("невдалий токен")
	}

	return claims, nil
}

func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	rawToken := r.URL.Query().Get("token")
	if rawToken == "" {
		http.Error(w, "Відсутній токен верифікації", http.StatusBadRequest)
		return
	}

	tokenHash := HashToken(rawToken)

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

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

func ResendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req ResendVerificationRequest
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

	rawToken, err := GenerateRefreshToken()
	if err != nil {
		http.Error(w, "Помилка генерації нового токена", http.StatusInternalServerError)
		return
	}

	tokenHash := HashToken(rawToken)
	expiresAt := time.Now().Add(24 * time.Hour)

	tokenQuery := `
	INSERT INTO  email_verification_tokens (user_id, token_hash, expires_at)
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

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req ForgotPasswordRequest
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

		rawToken, genErr := GenerateRefreshToken()
		if genErr == nil {
			tokenHash := HashToken(rawToken)
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

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не підтримується", http.StatusMethodNotAllowed)
		return
	}

	var req ResetPasswordRequest
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

	tokenHash := HashToken(req.Token)

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

	hashedPassword, err := HashPassword(req.NewPassword)
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
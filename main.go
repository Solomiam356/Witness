package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Solomiam356/witness-backend/internal/config"
	"github.com/Solomiam356/witness-backend/internal/database"
	"github.com/Solomiam356/witness-backend/internal/handler"
	"github.com/Solomiam356/witness-backend/internal/middleware"
	"github.com/Solomiam356/witness-backend/internal/repository"
	"github.com/Solomiam356/witness-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	// Підключаємося до PostgreSQL
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Не вдалося підключитися до БД: %v", err)
	}

	// Запускаємо схему бази даних Witness V2.1
	if err := database.InitSchema(); err != nil {
		log.Fatalf("Не вдалося створити схему даних: %v", err)
	}

	// 1. Ініціалізуємо шари архітектури (Dependency Injection)
	testimonyRepo := repository.NewTestimonyRepository(database.DB)
	aiSvc := service.NewAIService(cfg.GeminiAPIKey)
	testimonySvc := service.NewTestimonyService(testimonyRepo, aiSvc)
	testimonyHandler := handler.NewTestimonyHandler(testimonySvc)

	sessionRepo := repository.NewSessionRepository(database.DB)
	sessionService := service.NewSessionService(sessionRepo)

	authHandler := handler.NewAuthHandler(sessionService)

	reportRepo := repository.NewReportRepository(database.DB)
	reportHandler := handler.NewReportHandler(reportRepo)

	taskRepo := repository.NewTaskRepository(database.DB)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	// Ініціалізуємо роутер Chi
	r := chi.NewRouter()

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "https://*.fly.dev"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	r.Use(corsMiddleware.Handler)
	r.Use(middleware.Logger)

	limiterManager := middleware.NewIPManager()

	r.Route("/auth", func(r chi.Router) {
		r.Use(middleware.RateLimiter(limiterManager))

		r.Post("/signup", authHandler.SignUp)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)

		r.Get("/verify-email", authHandler.VerifyEmail)
		r.Post("/resend-verification", authHandler.ResendVerification)

		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.Post("/reset-password", authHandler.ResetPassword)

		r.Get("/sessions", authHandler.GetMySessions)
	})

	// Базовий маршрут перевірки працездатності сервера
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// 2. ЗАХИЩЕНІ МАРШРУТИ (Доступні тільки авторизованим користувачам)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.RequireVerifiedEmail)

		r.Post("/testimonies/{id}/report", reportHandler.CreateReport)

		r.Post("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(middleware.UserIDKey).(string)
			if !ok {
				http.Error(w, "Користувач не визначений", http.StatusUnauthorized)
				return
			}

			if err := sessionRepo.RevokeAllByUserID(r.Context(), userID); err != nil {
				http.Error(w, "Не вдалося закрити сесію: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Ви успішно вийшли з системи на всіх пристроях"})
		})

		// Маршрути для тасок (виправлено під Chi)
		r.Post("/tasks", taskHandler.Create)
		r.Get("/tasks", taskHandler.GetAll)
		r.Patch("/tasks/{id}", taskHandler.UpdateStatus)
		r.Delete("/tasks/{id}", taskHandler.Delete)

		// Перевірка поточного юзера
		r.Get("/auth/me", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(middleware.UserIDKey).(string)
			role := r.Context().Value(middleware.UserRoleKey).(string)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Доступ дозволено!",
				"user_id": userID,
				"role":    role,
			})
		})

		// Маршрути для свідчень (виправлено під Chi)
		r.Post("/testimonies", testimonyHandler.Create)
		r.Get("/testimonies", testimonyHandler.GetAll)
		r.Get("/testimonies/feed", testimonyHandler.GetFeed)

		r.With(middleware.RequireRole("admin", "moderator")).Delete("/testimonies/{id}", testimonyHandler.Delete)
	})

	// Налаштування HTTP-сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Створюємо канал для відловлювання сигналів ОС
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Запускаємо сервер в окремому фоновому потоці (Горутині)
	go func() {
		fmt.Println("Сервер Witness запущено на http://localhost:" + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Помилка сервера: %v", err)
		}
	}()

	<-quit
	fmt.Println("\nОтримано сигнал зупинки...")

	// Даємо серверу 5 секунд на завершення активних запитів
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Критична зупинка сервера: %+v", err)
	}

	database.Close()
	fmt.Println("Сервер успішно зупинено. Всі дані збережено.")
}
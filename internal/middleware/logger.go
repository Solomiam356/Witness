package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)
// Logger — кастомний middleware для логування запитів
func Logger(next http.Handler)  http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

// Використовуємо ResponseWriter, який може запам'ятати статус код
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

// Викликаємо наступний handler
		next.ServeHTTP(ww, r)

// Після виконання handler'a логуємо
		duration := time.Since(start)

		log.Printf("[%s] %s %s | Status: %d | Duration: %v | IP: %s | Agent: %s",
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			ww.Status(),
			duration,
			r.RemoteAddr,
			r.UserAgent(),
)
	})
}
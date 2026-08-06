package middleware

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type IPManager struct {
	ips map[string]*rate.Limiter
	mu sync.RWMutex
}

func NewIPManager() *IPManager {
	return &IPManager{ips: make(map[string]*rate.Limiter)}
}

func (i *IPManager) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(1,3)
		i.ips[ip] = limiter
	}
	return limiter
}

func RateLimiter(manager *IPManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			limiter := manager.GetLimiter(ip)

			if !limiter.Allow() {
				http.Error(w, "Забагато запитів. Спробуйте пізніше.", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
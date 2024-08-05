package api_middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/logger"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/repository"
	"golang.org/x/time/rate"
)

type AuthMiddleware struct {
	userRepo repository.UserRepository
}

func NewAuthMiddleware(userRepo repository.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{userRepo: userRepo}
}

var (
	limiter = rate.NewLimiter(rate.Every(time.Second), 10)
	clients = make(map[string]*rate.Limiter)
	mu      sync.Mutex
)

func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			logger.Error("No API key provided")
			http.Error(w, "No API key provided", http.StatusUnauthorized)
			return
		}
		user, err := am.userRepo.GetByAPIKey(r.Context(), apiKey)
		if err != nil {
			logger.Errorf("Invalid API key: %s", apiKey)
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(role model.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value("user").(*model.User)
			if !ok {
				logger.Error("User not found in context")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if user.Role != role {
				logger.Errorf("User %s does not have required role %s", user.Username, role)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			logger.Errorf("Rate limit exceeded for IP: %s", r.RemoteAddr)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

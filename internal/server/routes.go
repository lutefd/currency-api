package server

import (
	"github.com/Lutefd/challenge-bravo/internal/handler"
	api_middleware "github.com/Lutefd/challenge-bravo/internal/middleware"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) registerRoutes(currencyService *service.CurrencyService, userService *service.UserService) {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	authMiddleware := api_middleware.NewAuthMiddleware(s.userRepo)

	router.Get("/healthz", handler.HandlerReadiness)
	currencyHandler := handler.NewCurrencyHandler(currencyService)
	userHandler := handler.NewUserHandler(userService)
	router.Route("/auth", func(r chi.Router) {
		r.With(api_middleware.RateLimitMiddleware).Post("/register", userHandler.Register)
		r.With(api_middleware.RateLimitMiddleware).Post("/login", userHandler.Login)
	})
	router.Route("/currency", func(r chi.Router) {
		r.Get("/convert", currencyHandler.ConvertCurrency)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(api_middleware.RequireRole(model.RoleAdmin))
			r.Post("/", currencyHandler.AddCurrency)
			r.Put("/{code}", currencyHandler.UpdateCurrency)
			r.Delete("/{code}", currencyHandler.RemoveCurrency)
		})
	})
	s.Router = router
}

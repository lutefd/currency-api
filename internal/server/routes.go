package server

import (
	"github.com/Lutefd/challenge-bravo/internal/handler"
	"github.com/Lutefd/challenge-bravo/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) registerRoutes(currencyService *service.CurrencyService) {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/healthz", handler.HandlerReadiness)
	currencyHandler := handler.NewCurrencyHandler(currencyService)
	router.Route("/currency", func(r chi.Router) {
		r.Get("/convert", currencyHandler.ConvertCurrency)
		r.Post("/", currencyHandler.AddCurrency)
		r.Delete("/{code}", currencyHandler.RemoveCurrency)
	})
	s.router = router
}

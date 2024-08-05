package server

import (
	"github.com/Lutefd/challenge-bravo/internal/handler"
	"github.com/Lutefd/challenge-bravo/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) registerRoutes() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/healthz", handler.HandlerReadiness)
	router.Route("/currency", s.loadCurrencyRoutes)

	s.router = router
}

func (s *Server) loadCurrencyRoutes(router chi.Router) {
	currencyService := service.NewCurrencyService()
	currencyHandler := handler.NewCurrencyHandler(currencyService)

	router.Get("/convert", currencyHandler.ConvertCurrency)
	router.Post("/", currencyHandler.AddCurrency)
	router.Delete("/{code}", currencyHandler.RemoveCurrency)
}

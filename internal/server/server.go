package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/cache"
	"github.com/Lutefd/challenge-bravo/internal/commons"
	"github.com/Lutefd/challenge-bravo/internal/logger"
	"github.com/Lutefd/challenge-bravo/internal/repository"
	"github.com/Lutefd/challenge-bravo/internal/service"
	"github.com/Lutefd/challenge-bravo/internal/worker"
)

type Server struct {
	config        Config
	httpServer    *http.Server
	Router        http.Handler
	rateUpdater   *worker.RateUpdater
	currencyRepo  repository.CurrencyRepository
	currencyCache cache.Cache
	externalAPI   worker.ExternalAPIClient
	userRepo      repository.UserRepository
	logRepo       repository.LogRepository
}

func NewServer(config Config) (*Server, error) {
	repo, err := repository.NewPostgresCurrencyRepository(config.PostgresConn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}
	userRepo, err := repository.NewPostgresUserRepository(config.PostgresConn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user repository: %w", err)
	}
	logRepo, err := repository.NewPostgresLogRepository(config.PostgresConn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log repository: %w", err)
	}
	redisCache, err := cache.NewRedisCache(config.RedisAddr, config.RedisPass)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}
	externalAPI := worker.NewOpenExchangeRatesClient(config.APIKey)
	currencyService := service.NewCurrencyService(repo, redisCache)
	userService := service.NewUserService(userRepo)
	rateUpdater := worker.NewRateUpdater(repo, redisCache, externalAPI, 1*time.Hour)
	logger.InitLogger(logRepo)
	partManager := logger.NewPartitionManager(logRepo)
	if err := partManager.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start partition manager: %w", err)
	}
	server := &Server{
		config:        config,
		currencyRepo:  repo,
		currencyCache: redisCache,
		externalAPI:   externalAPI,
		rateUpdater:   rateUpdater,
		userRepo:      userRepo,
		logRepo:       logRepo,
	}

	server.registerRoutes(currencyService, userService)

	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", config.ServerPort),
		Handler:      server.Router,
		IdleTimeout:  commons.ServerIdleTimeout,
		ReadTimeout:  commons.ServerReadTimeout,
		WriteTimeout: commons.ServerWriteTimeout,
	}

	return server, nil
}

func (s *Server) Start(ctx context.Context) error {
	go s.rateUpdater.Start(ctx)
	go func() {
		logger.Infof("Server started on port %d", s.config.ServerPort)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	return s.Shutdown()
}

func (s *Server) Shutdown() error {
	logger.Info("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		logger.Errorf("HTTP server shutdown error: %v", err)
		return err
	}
	if err := s.currencyRepo.Close(); err != nil {
		logger.Errorf("database connection close error: %v", err)
		return err
	}

	if err := s.currencyCache.Close(); err != nil {
		logger.Errorf("cache connection close error: %v", err)
		return err
	}
	if err := logger.Shutdown(ctx); err != nil {
		logger.Errorf("error shutting down logger: %v", err)
	}

	fmt.Println("server shutdown complete")
	return nil
}

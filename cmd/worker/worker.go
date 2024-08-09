package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/cache"
	"github.com/Lutefd/challenge-bravo/internal/commons"
	"github.com/Lutefd/challenge-bravo/internal/logger"
	"github.com/Lutefd/challenge-bravo/internal/repository"
	"github.com/Lutefd/challenge-bravo/internal/worker"
	"github.com/joho/godotenv"
)

type dependencies struct {
	currencyRepo repository.CurrencyRepository
	cache        cache.Cache
	logRepo      repository.LogRepository
	externalAPI  worker.ExternalAPIClient
	rateUpdater  RateUpdater
	partitionMgr PartitionManager
}

type RateUpdater interface {
	Start(ctx context.Context)
}

type PartitionManager interface {
	Start(ctx context.Context) error
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	config, err := commons.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	deps, err := initDependencies(config)
	if err != nil {
		log.Fatalf("Failed to initialize dependencies: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- runWorker(ctx, deps)
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		if err != nil {
			log.Fatalf("Worker failed: %v", err)
		}
	case <-signalChan:
		log.Println("Shutdown signal received, initiating graceful shutdown...")
		cancel()

		select {
		case <-errChan:
			log.Println("Worker shut down gracefully")
		case <-time.After(30 * time.Second):
			log.Println("Shutdown timed out")
		}
	}
}

func initDependencies(config commons.Config) (*dependencies, error) {
	currencyRepo, err := repository.NewPostgresCurrencyRepository(config.PostgresConn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	redisCache, err := cache.NewRedisCache(config.RedisAddr, config.RedisPass)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	logRepo, err := repository.NewPostgresLogRepository(config.PostgresConn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize log repository: %w", err)
	}

	externalAPI := worker.NewOpenExchangeRatesClient(config.APIKey)
	rateUpdater := worker.NewRateUpdater(currencyRepo, redisCache, externalAPI, commons.RateUpdaterInterval)
	partManager := logger.NewPartitionManager(logRepo)

	return &dependencies{
		currencyRepo: currencyRepo,
		cache:        redisCache,
		logRepo:      logRepo,
		externalAPI:  externalAPI,
		rateUpdater:  rateUpdater,
		partitionMgr: partManager,
	}, nil
}

func runWorker(ctx context.Context, deps *dependencies) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.InitLogger(deps.logRepo)

		if err := deps.partitionMgr.Start(ctx); err != nil {
			errChan <- fmt.Errorf("failed to start partition manager: %w", err)
			return
		}

		deps.rateUpdater.Start(ctx)

		<-ctx.Done()
		log.Println("Worker shutting down...")
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	defer func() {
		if err := deps.currencyRepo.Close(); err != nil {
			log.Printf("Error closing currency repository: %v", err)
		}
		if err := deps.cache.Close(); err != nil {
			log.Printf("Error closing cache: %v", err)
		}
		if err := deps.logRepo.Close(); err != nil {
			log.Printf("Error closing log repository: %v", err)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

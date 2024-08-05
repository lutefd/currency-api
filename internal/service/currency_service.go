package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/cache"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/repository"
	"github.com/Lutefd/challenge-bravo/internal/worker"
)

type CurrencyService struct {
	repo        repository.CurrencyRepository
	cache       cache.Cache
	externalAPI worker.ExternalAPIClient
}

func NewCurrencyService(repo repository.CurrencyRepository, cache cache.Cache, externalAPI worker.ExternalAPIClient) *CurrencyService {
	return &CurrencyService{
		repo:        repo,
		cache:       cache,
		externalAPI: externalAPI,
	}
}

func (s *CurrencyService) Convert(ctx context.Context, from, to string, amount float64) (float64, error) {
	fromRate, err := s.getRate(ctx, from)
	if err != nil {
		return 0, err
	}

	toRate, err := s.getRate(ctx, to)
	if err != nil {
		return 0, err
	}
	usdAmount := amount / fromRate
	result := usdAmount * toRate

	return result, nil
}

func (s *CurrencyService) getRate(ctx context.Context, code string) (float64, error) {
	rate, err := s.cache.Get(ctx, code)
	if err == nil {
		return rate, nil
	}
	currency, err := s.repo.GetByCode(ctx, code)
	if err == nil {
		s.cache.Set(ctx, code, currency.Rate, 1*time.Hour)
		return currency.Rate, nil
	}
	rates, err := s.externalAPI.FetchRates(ctx)
	if err != nil {
		return 0, err
	}

	rate, ok := rates.Rates[code]
	if !ok {
		return 0, fmt.Errorf("currency %s not found", code)
	}
	currency = &model.Currency{
		Code:      code,
		Rate:      rate,
		UpdatedAt: time.Now(),
	}
	err = s.repo.Create(ctx, currency)
	if err != nil {
		return 0, err
	}

	s.cache.Set(ctx, code, rate, 1*time.Hour)

	return rate, nil
}

func (s *CurrencyService) AddCurrency(ctx context.Context, code string, rate float64) error {
	return errors.New("not implemented")
}

func (s *CurrencyService) RemoveCurrency(ctx context.Context, code string) error {
	return errors.New("not implemented")
}
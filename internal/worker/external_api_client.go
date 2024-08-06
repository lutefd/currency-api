package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/logger"
	"github.com/Lutefd/challenge-bravo/internal/model"
)

type OpenExchangeRatesClient struct {
	apiKey     string
	client     *http.Client
	baseURL    string
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

type OpenExchangeRatesClientOption func(*OpenExchangeRatesClient)

func WithMaxRetries(maxRetries int) OpenExchangeRatesClientOption {
	return func(c *OpenExchangeRatesClient) {
		c.maxRetries = maxRetries
	}
}

func WithBaseDelay(baseDelay time.Duration) OpenExchangeRatesClientOption {
	return func(c *OpenExchangeRatesClient) {
		c.baseDelay = baseDelay
	}
}

func WithMaxDelay(maxDelay time.Duration) OpenExchangeRatesClientOption {
	return func(c *OpenExchangeRatesClient) {
		c.maxDelay = maxDelay
	}
}

func NewOpenExchangeRatesClient(apiKey string, options ...OpenExchangeRatesClientOption) *OpenExchangeRatesClient {
	client := &OpenExchangeRatesClient{
		apiKey:     apiKey,
		client:     &http.Client{},
		baseURL:    "https://openexchangerates.org/api",
		maxRetries: 3,
		baseDelay:  time.Second,
		maxDelay:   30 * time.Second,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

func (c *OpenExchangeRatesClient) FetchRates(ctx context.Context) (*model.ExchangeRates, error) {
	url := fmt.Sprintf("%s/latest.json?app_id=%s", c.baseURL, c.apiKey)

	var rates *model.ExchangeRates
	var err error
	var resp *http.Response

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		rates, resp, err = c.doRequest(ctx, url)
		if err == nil {
			return rates, nil
		}

		if !c.shouldRetry(err, resp) {
			logger.Errorf("Non-retryable error encountered: %v", err)
			return nil, err
		}

		backoffDuration := c.calculateBackoff(attempt)
		logger.Infof("Retry attempt %d/%d after %v due to error: %v", attempt+1, c.maxRetries, backoffDuration, err)

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoffDuration):
		}
	}

	logger.Errorf("Max retries reached. Last error: %v", err)
	return nil, fmt.Errorf("max retries reached: %w", err)
}

func (c *OpenExchangeRatesClient) doRequest(ctx context.Context, url string) (*model.ExchangeRates, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var rates model.ExchangeRates
	err = json.NewDecoder(resp.Body).Decode(&rates)
	if err != nil {
		return nil, resp, fmt.Errorf("failed to decode response: %w", err)
	}

	return &rates, resp, nil
}

func (c *OpenExchangeRatesClient) shouldRetry(err error, resp *http.Response) bool {
	if err == nil {
		return false
	}

	if netErr, ok := err.(net.Error); ok {
		logger.Infof("Network error occurred: %v. Will retry.", netErr)
		return netErr.Timeout()
	}

	if resp != nil {
		if resp.StatusCode >= 500 {
			logger.Infof("Received 5xx status code: %d. Will retry.", resp.StatusCode)
			return true
		}
		if resp.StatusCode == 429 {
			logger.Infof("Received 429 status code: %d. Will retry.", resp.StatusCode)
			return true
		}

	}

	logger.Infof("Non-retryable error occurred: %v", err)
	return false
}

func (c *OpenExchangeRatesClient) calculateBackoff(attempt int) time.Duration {
	delay := c.baseDelay * time.Duration(1<<uint(attempt))
	if delay > c.maxDelay {
		delay = c.maxDelay
	}
	jitter := time.Duration(rand.Int63n(int64(delay) / 2))
	return delay + jitter
}

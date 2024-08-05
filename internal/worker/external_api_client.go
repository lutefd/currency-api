package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Lutefd/challenge-bravo/internal/model"
)

type OpenExchangeRatesClient struct {
	apiKey string
	client *http.Client
}

func NewOpenExchangeRatesClient(apiKey string) *OpenExchangeRatesClient {
	return &OpenExchangeRatesClient{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (c *OpenExchangeRatesClient) FetchRates(ctx context.Context) (*model.ExchangeRates, error) {
	url := fmt.Sprintf("https://openexchangerates.org/api/latest.json?app_id=%s", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var rates model.ExchangeRates
	err = json.NewDecoder(resp.Body).Decode(&rates)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &rates, nil
}

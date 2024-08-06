package worker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOpenExchangeRatesClient(t *testing.T) {
	apiKey := "test-api-key"
	client := NewOpenExchangeRatesClient(apiKey)

	assert.Equal(t, apiKey, client.apiKey)
	assert.Equal(t, 3, client.maxRetries)
	assert.Equal(t, time.Second, client.baseDelay)
	assert.Equal(t, 30*time.Second, client.maxDelay)

	customClient := NewOpenExchangeRatesClient(
		apiKey,
		WithMaxRetries(5),
		WithBaseDelay(2*time.Second),
		WithMaxDelay(1*time.Minute),
	)

	assert.Equal(t, 5, customClient.maxRetries)
	assert.Equal(t, 2*time.Second, customClient.baseDelay)
	assert.Equal(t, 1*time.Minute, customClient.maxDelay)
}

func TestOpenExchangeRatesClient_FetchRates(t *testing.T) {
	tests := []struct {
		name           string
		responses      []mockResponse
		expectedError  bool
		expectedRetries int
	}{
		{
			name: "Successful response",
			responses: []mockResponse{
				{statusCode: http.StatusOK, body: `{"rates":{"USD":1,"EUR":0.85}}`},
			},
			expectedError:  false,
			expectedRetries: 0,
		},
		{
			name: "Retry on 500 error",
			responses: []mockResponse{
				{statusCode: http.StatusInternalServerError, body: ""},
				{statusCode: http.StatusOK, body: `{"rates":{"USD":1,"EUR":0.85}}`},
			},
			expectedError:  false,
			expectedRetries: 1,
		},
		{
			name: "Retry on network error",
			responses: []mockResponse{
				{err: &networkError{message: "connection refused"}},
				{statusCode: http.StatusOK, body: `{"rates":{"USD":1,"EUR":0.85}}`},
			},
			expectedError:  false,
			expectedRetries: 1,
		},
		{
			name: "Max retries reached",
			responses: []mockResponse{
				{statusCode: http.StatusInternalServerError, body: ""},
				{statusCode: http.StatusInternalServerError, body: ""},
				{statusCode: http.StatusInternalServerError, body: ""},
				{statusCode: http.StatusInternalServerError, body: ""},
			},
			expectedError:  true,
			expectedRetries: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newMockServer(tt.responses)
			defer server.Close()

			client := NewOpenExchangeRatesClient("test-api-key")
			client.baseURL = server.URL
			client.client = server.Client()

			rates, err := client.FetchRates(context.Background())

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, rates)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, rates)
			}

			assert.Equal(t, tt.expectedRetries, server.requestCount-1)
		})
	}
}

func TestOpenExchangeRatesClient_shouldRetry(t *testing.T) {
	client := NewOpenExchangeRatesClient("test-api-key")

	tests := []struct {
		name     string
		err      error
		resp     *http.Response
		expected bool
	}{
		{
			name:     "No error",
			err:      nil,
			resp:     nil,
			expected: false,
		},
		{
			name:     "Network timeout",
			err:      &networkError{message: "timeout", timeout: true},
			resp:     nil,
			expected: true,
		},
		{
			name:     "500 status code",
			err:      fmt.Errorf("server error"),
			resp:     &http.Response{StatusCode: http.StatusInternalServerError},
			expected: true,
		},
		{
			name:     "429 status code",
			err:      fmt.Errorf("rate limit exceeded"),
			resp:     &http.Response{StatusCode: http.StatusTooManyRequests},
			expected: true,
		},
		{
			name:     "400 status code",
			err:      fmt.Errorf("bad request"),
			resp:     &http.Response{StatusCode: http.StatusBadRequest},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.shouldRetry(tt.err, tt.resp)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOpenExchangeRatesClient_calculateBackoff(t *testing.T) {
	client := NewOpenExchangeRatesClient("test-api-key")

	for i := 0; i < 5; i++ {
		backoff := client.calculateBackoff(i)
		assert.GreaterOrEqual(t, backoff, client.baseDelay*(1<<uint(i)))
		assert.LessOrEqual(t, backoff, client.baseDelay*(1<<uint(i))*3/2)
	}

	// Test max delay
	largeBackoff := client.calculateBackoff(10)
	assert.LessOrEqual(t, largeBackoff, client.maxDelay*3/2)
}

type mockResponse struct {
	statusCode int
	body       string
	err        error
}

type mockServer struct {
	*httptest.Server
	responses    []mockResponse
	requestCount int
}

func newMockServer(responses []mockResponse) *mockServer {
	ms := &mockServer{responses: responses}
	ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ms.requestCount >= len(ms.responses) {
			http.Error(w, "Too many requests", http.StatusInternalServerError)
			return
		}

		resp := ms.responses[ms.requestCount]
		ms.requestCount++

		if resp.err != nil {
			http.Error(w, resp.err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(resp.statusCode)
		w.Write([]byte(resp.body))
	}))
	return ms
}

type networkError struct {
	message string
	timeout bool
}

func (e *networkError) Error() string   { return e.message }
func (e *networkError) Timeout() bool   { return e.timeout }
func (e *networkError) Temporary() bool { return true }

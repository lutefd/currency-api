package commons_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Lutefd/challenge-bravo/internal/commons"
	"github.com/Lutefd/challenge-bravo/internal/logger"
)

func TestRespondWithError(t *testing.T) {
	var buf bytes.Buffer
	oldErrorLogger := logger.ErrorLogger
	logger.ErrorLogger = log.New(&buf, "", 0)
	defer func() { logger.ErrorLogger = oldErrorLogger }()

	tests := []struct {
		name           string
		code           int
		msg            string
		expectedLog    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "4xx error",
			code:           400,
			msg:            "Bad Request",
			expectedLog:    "",
			expectedStatus: 400,
			expectedBody:   `{"error":"Bad Request"}`,
		},
		{
			name:           "5xx error",
			code:           500,
			msg:            "Internal Server Error",
			expectedLog:    "responding with 500 error: Internal Server Error",
			expectedStatus: 500,
			expectedBody:   `{"error":"Internal Server Error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			w := httptest.NewRecorder()
			commons.RespondWithError(w, tt.code, tt.msg)

			if tt.expectedLog != "" {
				logOutput := strings.TrimSpace(buf.String())
				if !strings.Contains(logOutput, tt.expectedLog) {
					t.Errorf("Expected log to contain: %s, got: %s", tt.expectedLog, logOutput)
				}
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code: %d, got: %d", tt.expectedStatus, w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("Expected body: %s, got: %s", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		payload        interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Success response",
			code:           200,
			payload:        map[string]string{"message": "Success"},
			expectedStatus: 200,
			expectedBody:   `{"message":"Success"}`,
		},
		{
			name:           "Error response",
			code:           400,
			payload:        map[string]string{"error": "Bad Request"},
			expectedStatus: 400,
			expectedBody:   `{"error":"Bad Request"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			commons.RespondWithJSON(w, tt.code, tt.payload)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code: %d, got: %d", tt.expectedStatus, w.Code)
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type: application/json, got: %s", w.Header().Get("Content-Type"))
			}

			var result map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &result)
			if err != nil {
				t.Errorf("Error unmarshalling response body: %v", err)
			}

			expectedResult := make(map[string]string)
			err = json.Unmarshal([]byte(tt.expectedBody), &expectedResult)
			if err != nil {
				t.Errorf("Error unmarshalling expected body: %v", err)
			}

			if !jsonEqual(result, expectedResult) {
				t.Errorf("Expected body: %s, got: %s", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func jsonEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}
	return true
}

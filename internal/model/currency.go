package model

import (
	"time"

	"github.com/google/uuid"
)

type Currency struct {
	Code      string    `json:"code"`
	Rate      float64   `json:"rate"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by"`
	UpdatedBy uuid.UUID `json:"updated_by"`
	CreatedAt time.Time `json:"created_at"`
}

type ExchangeRates struct {
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
}

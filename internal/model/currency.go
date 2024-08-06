package model

import (
	"time"
)

type Currency struct {
	Code      string    `json:"code"`
	Rate      float64   `json:"rate"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExchangeRates struct {
	Timestamp int64              `json:"timestamp"`
	Base      string             `json:"base"`
	Rates     map[string]float64 `json:"rates"`
}

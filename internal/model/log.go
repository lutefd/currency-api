package model

import (
	"time"

	"github.com/google/uuid"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelError LogLevel = "ERROR"
)

type Log struct {
	ID        uuid.UUID `json:"id"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

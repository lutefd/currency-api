package server

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	PostgresConn string
	RedisAddr    string
	RedisPass    string
	ServerPort   uint16
	APIKey       string
}

func LoadConfig() (Config, error) {
	var config Config
	var errors []string

	config.RedisPass = os.Getenv("REDIS_PASSWORD")
	if config.RedisPass == "" {
		errors = append(errors, "REDIS_PASSWORD is not set")
	}

	config.RedisAddr = os.Getenv("REDIS_ADDR")
	if config.RedisAddr == "" {
		errors = append(errors, "REDIS_ADDR is not set")
	}

	config.PostgresConn = os.Getenv("POSTGRES_CONN")
	if config.PostgresConn == "" {
		errors = append(errors, "POSTGRES_CONN is not set")
	}

	config.APIKey = os.Getenv("API_KEY")
	if config.APIKey == "" {
		errors = append(errors, "API_KEY is not set")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		errors = append(errors, "SERVER_PORT is not set")
	} else {
		parsedServerPort, err := strconv.ParseUint(serverPort, 10, 16)
		if err != nil {
			errors = append(errors, fmt.Sprintf("invalid SERVER_PORT: %s", err))
		} else {
			config.ServerPort = uint16(parsedServerPort)
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Println("Configuration Error:", err)
		}
		return Config{}, fmt.Errorf("configuration errors occurred")
	}

	return config, nil
}

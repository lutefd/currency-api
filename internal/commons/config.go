package commons

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

const (
	decimalBase = 10
	bitSize     = 16
)

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

	pg_user := os.Getenv("POSTGRES_USER")
	if pg_user == "" {
		errors = append(errors, "POSTGRES_USER is not set")
	}

	pg_pass := os.Getenv("POSTGRES_PASSWORD")
	if pg_pass == "" {
		errors = append(errors, "POSTGRES_PASSWORD is not set")
	}

	pg_host := os.Getenv("POSTGRES_HOST")
	if pg_host == "" {
		errors = append(errors, "POSTGRES_HOST is not set")
	}
	pg_port := os.Getenv("POSTGRES_PORT")
	if pg_port == "" {
		errors = append(errors, "POSTGRES_PORT is not set")
	}

	pg_db := os.Getenv("POSTGRES_NAME")
	if pg_db == "" {
		errors = append(errors, "POSTGRES_NAME is not set")
	}

	config.PostgresConn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", pg_user, pg_pass, pg_host, pg_port, pg_db)

	config.APIKey = os.Getenv("API_KEY")
	if config.APIKey == "" {
		errors = append(errors, "API_KEY is not set")
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		errors = append(errors, "SERVER_PORT is not set")
	} else {
		parsedServerPort, err := strconv.ParseUint(serverPort, decimalBase, bitSize)
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

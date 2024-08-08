package server_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Lutefd/challenge-bravo/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, pair := range originalEnv {
			parts := strings.SplitN(pair, "=", 2)
			os.Setenv(parts[0], parts[1])
		}
	}()

	setEnv := func(key, value string) {
		os.Setenv(key, value)
	}

	t.Run("Valid configuration", func(t *testing.T) {
		setEnv("REDIS_PASSWORD", "password")
		setEnv("REDIS_ADDR", "localhost:6379")
		setEnv("POSTGRES_USER", "user")
		setEnv("POSTGRES_PASSWORD", "pass")
		setEnv("POSTGRES_HOST", "localhost")
		setEnv("POSTGRES_PORT", "5432")
		setEnv("POSTGRES_NAME", "db")
		setEnv("API_KEY", "my-api-key")
		setEnv("SERVER_PORT", "8080")

		config, err := server.LoadConfig()

		assert.NoError(t, err)
		assert.Equal(t, "password", config.RedisPass)
		assert.Equal(t, "localhost:6379", config.RedisAddr)
		assert.Equal(t, "postgres://user:pass@localhost:5432/db?sslmode=disable", config.PostgresConn)
		assert.Equal(t, "my-api-key", config.APIKey)
		assert.Equal(t, uint16(8080), config.ServerPort)
	})

	t.Run("Missing environment variables", func(t *testing.T) {
		os.Clearenv()

		_, err := server.LoadConfig()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration errors occurred")
	})

	t.Run("Invalid SERVER_PORT", func(t *testing.T) {
		setEnv("REDIS_PASSWORD", "password")
		setEnv("REDIS_ADDR", "localhost:6379")
		setEnv("POSTGRES_USER", "user")
		setEnv("POSTGRES_PASSWORD", "pass")
		setEnv("POSTGRES_HOST", "localhost")
		setEnv("POSTGRES_PORT", "5432")
		setEnv("POSTGRES_NAME", "db")
		setEnv("API_KEY", "my-api-key")
		setEnv("SERVER_PORT", "invalid-port")

		_, err := server.LoadConfig()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration errors occurred")
	})

	t.Run("Partial configuration", func(t *testing.T) {
		os.Clearenv()
		setEnv("REDIS_PASSWORD", "password")
		setEnv("REDIS_ADDR", "localhost:6379")

		_, err := server.LoadConfig()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "configuration errors occurred")
	})
}

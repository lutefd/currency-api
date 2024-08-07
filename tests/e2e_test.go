package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/server"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var testServer *server.Server
var adminAPIKey string

func TestMain(m *testing.M) {
	// Load the .env file
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Setup
	err = setup()
	if err != nil {
		log.Fatalf("Error setting up test environment: %v", err)
	}

	// Run tests
	code := m.Run()

	// Teardown
	err = teardown()
	if err != nil {
		log.Printf("Error tearing down test environment: %v", err)
	}

	os.Exit(code)
}

func setup() error {
	var err error

	// Create test database
	err = createTestDatabase()
	if err != nil {
		return fmt.Errorf("error creating test database: %w", err)
	}

	// Run migrations
	err = runMigrations()
	if err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	// Seed the database
	err = seedDatabase()
	if err != nil {
		return fmt.Errorf("error seeding database: %w", err)
	}

	// Set up the test environment
	config, err := server.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Use the test database connection string
	config.PostgresConn = os.Getenv("TEST_POSTGRES_CONN")

	// Create a new server instance
	testServer, err = server.NewServer(config)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return nil
}

func createTestDatabase() error {
	connURL, err := url.Parse(os.Getenv("POSTGRES_CONN"))
	if err != nil {
		return fmt.Errorf("error parsing database URL: %w", err)
	}

	connURL.Path = "/postgres"
	db, err := sql.Open("postgres", connURL.String())
	if err != nil {
		return fmt.Errorf("error connecting to postgres: %w", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE test_currency_db")
	if err != nil {
		return fmt.Errorf("error creating test database: %w", err)
	}

	connURL.Path = "/test_currency_db"
	testDBConn := connURL.String()

	os.Setenv("TEST_POSTGRES_CONN", testDBConn)

	return nil
}

func runMigrations() error {
	projectRoot, err := filepath.Abs("../")
	if err != nil {
		return fmt.Errorf("error getting project root path: %w", err)
	}

	migrationDir := filepath.Join(projectRoot, "sql", "schema")

	cmd := exec.Command("goose",
		"-dir", migrationDir,
		"postgres", os.Getenv("TEST_POSTGRES_CONN"),
		"up")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running migrations: %w\nOutput: %s", err, string(output))
	}

	log.Printf("Migration output: %s", string(output))
	return nil
}

func seedDatabase() error {
	db, err := sql.Open("postgres", os.Getenv("TEST_POSTGRES_CONN"))
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(`
		INSERT INTO currencies (code, rate, updated_at) VALUES
		('USD', 1.0, $1),
		('EUR', 0.85, $1),
		('GBP', 0.75, $1)
	`, time.Now())
	createAdminUser(context.Background(), db)
	return err
}

func teardown() error {
	if testServer != nil {
		if err := testServer.Shutdown(); err != nil {
			return fmt.Errorf("error shutting down test server: %w", err)
		}
	}

	connURL, err := url.Parse(os.Getenv("POSTGRES_CONN"))
	if err != nil {
		return fmt.Errorf("error parsing database URL: %w", err)
	}

	connURL.Path = "/postgres"
	db, err := sql.Open("postgres", connURL.String())
	if err != nil {
		return fmt.Errorf("error connecting to postgres: %w", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = 'test_currency_db'
		AND pid <> pg_backend_pid()
	`)
	if err != nil {
		return fmt.Errorf("error terminating connections to test database: %w", err)
	}

	_, err = db.Exec("DROP DATABASE IF EXISTS test_currency_db")
	if err != nil {
		return fmt.Errorf("error dropping test database: %w", err)
	}

	return nil
}

func setupTestServer(t *testing.T) {
	config, err := server.LoadConfig()
	require.NoError(t, err, "Failed to load config")

	config.PostgresConn = os.Getenv("TEST_POSTGRES_CONN")
	config.RedisAddr = os.Getenv("TEST_REDIS_ADDR")

	testServer, err = server.NewServer(config)
	require.NoError(t, err, "Failed to create test server")
}

func TestEndToEnd(t *testing.T) {
	setupTestServer(t)

	var apiKey string

	t.Run("Register User", func(t *testing.T) {
		payload := map[string]string{
			"username": "testuser",
			"password": "testpassword",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		testServer.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201, got %d", rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "testuser", response["username"])
		assert.NotEmpty(t, response["api_key"])
		apiKey = response["api_key"].(string)
	})

	t.Run("Login User", func(t *testing.T) {
		payload := map[string]string{
			"username": "testuser",
			"password": "testpassword",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		testServer.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200, got %d", rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "testuser", response["username"])
		assert.NotEmpty(t, response["api_key"])
	})

	t.Run("Convert Currency", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/currency/convert?from=USD&to=EUR&amount=100", nil)
		req.Header.Set("X-API-Key", apiKey)

		rr := httptest.NewRecorder()
		testServer.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200, got %d", rr.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "USD", response["from"])
		assert.Equal(t, "EUR", response["to"])
		assert.Equal(t, float64(100), response["amount"])
		assert.NotNil(t, response["result"])
	})

	t.Run("Add Currency", func(t *testing.T) {
		payload := map[string]interface{}{
			"code":        "GBPT",
			"rate_to_usd": 0.75,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/currency", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", adminAPIKey)

		rr := httptest.NewRecorder()
		testServer.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201, got %d", rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "currency added successfully", response["message"])
	})

	t.Run("Update Currency", func(t *testing.T) {
		payload := map[string]interface{}{
			"rate_to_usd": 0.78,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("PUT", "/currency/GBPT", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", adminAPIKey)

		rr := httptest.NewRecorder()
		testServer.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200, got %d", rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "currency updated successfully", response["message"])
	})

	t.Run("Remove Currency", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/currency/GBPT", nil)
		req.Header.Set("X-API-Key", adminAPIKey)

		rr := httptest.NewRecorder()
		testServer.Router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200, got %d", rr.Code)

		var response map[string]string
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "currency removed successfully", response["message"])
	})
}
func createAdminUser(ctx context.Context, db *sql.DB) error {
	adminAPIKey = generateAPIKey()
	adminUser := model.UserDB{
		ID:        uuid.New(),
		Username:  "admin",
		Password:  "password",
		Role:      model.RoleAdmin,
		APIKey:    adminAPIKey,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}
	adminUser.Password = string(hashedPassword)

	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, username, password, role, api_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, adminUser.ID, adminUser.Username, adminUser.Password, adminUser.Role, adminUser.APIKey, adminUser.CreatedAt, adminUser.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error inserting admin user: %w", err)
	}

	return nil
}
func generateAPIKey() string {
	return uuid.New().String()
}

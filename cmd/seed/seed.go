package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/server"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type dependencies struct {
	loadConfig func() (server.Config, error)
	openDB     func(driverName, dataSourceName string) (*sql.DB, error)
	newUUID    func() uuid.UUID
	timeNow    func() time.Time
	loadEnv    func(...string) error
}

var defaultDeps = dependencies{
	loadConfig: server.LoadConfig,
	openDB:     sql.Open,
	newUUID:    uuid.New,
	timeNow:    time.Now,
	loadEnv:    godotenv.Load,
}

func main() {
	if err := run(context.Background(), defaultDeps); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, deps dependencies) error {
	if err := deps.loadEnv(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	config, err := deps.loadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	db, err := deps.openDB("postgres", config.PostgresConn)
	if err != nil {
		return fmt.Errorf("error opening database connection: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("error connecting to the database: %w", err)
	}

	if err := createAdminUser(ctx, db, deps); err != nil {
		return fmt.Errorf("error creating admin user: %w", err)
	}

	fmt.Println("Admin user created successfully!")
	return nil
}

func createAdminUser(ctx context.Context, db *sql.DB, deps dependencies) error {
	adminUser := model.UserDB{
		ID:        deps.newUUID(),
		Username:  "admin",
		Password:  "password",
		Role:      model.RoleAdmin,
		APIKey:    deps.newUUID().String(),
		CreatedAt: deps.timeNow(),
		UpdatedAt: deps.timeNow(),
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

package main

import (
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

func main() {
	godotenv.Load()
	config, err := server.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	db, err := sql.Open("postgres", config.PostgresConn)
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}

	adminUser := model.UserDB{
		ID:        uuid.New(),
		Username:  "admin",
		Password:  "password",
		Role:      model.RoleAdmin,
		APIKey:    uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminUser.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("error hashing password: %v", err)
	}
	adminUser.Password = string(hashedPassword)

	_, err = db.Exec(`
		INSERT INTO users (id, username, password, role, api_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, adminUser.ID, adminUser.Username, adminUser.Password, adminUser.Role, adminUser.APIKey, adminUser.CreatedAt, adminUser.UpdatedAt)

	if err != nil {
		log.Fatalf("Error inserting admin user: %v", err)
	}

	fmt.Println("Admin user created successfully!")
}

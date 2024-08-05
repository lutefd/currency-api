package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/commons"
	"github.com/Lutefd/challenge-bravo/internal/logger"
	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.Errorf("Failed to decode user registration request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user.ID = uuid.New()
	user.Password = string(hashedPassword)
	user.Role = model.RoleUser // Default role
	user.APIKey = generateAPIKey()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Create(r.Context(), &user); err != nil {
		logger.Errorf("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	commons.RespondWithJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		logger.Errorf("Failed to decode login request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByUsername(r.Context(), credentials.Username)
	if err != nil {
		logger.Errorf("Failed to get user: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		logger.Errorf("Invalid password for user %s: %v", credentials.Username, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	user.Password = ""
	commons.RespondWithJSON(w, http.StatusOK, user)
}

func generateAPIKey() string {
	return uuid.New().String()
}

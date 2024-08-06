package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Lutefd/challenge-bravo/internal/commons"
	"github.com/Lutefd/challenge-bravo/internal/logger"
	"github.com/Lutefd/challenge-bravo/internal/service"
)

type UserHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		logger.Errorf("Failed to decode user registration request: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if credentials.Username == "" || credentials.Password == "" {
		logger.Errorf("Invalid input: username or password is empty")
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Create(r.Context(), credentials.Username, credentials.Password)
	if err != nil {
		logger.Errorf("Failed to create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

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

	if credentials.Username == "" || credentials.Password == "" {
		logger.Errorf("Invalid input: username or password is empty")
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Authenticate(r.Context(), credentials.Username, credentials.Password)
	if err != nil {
		logger.Errorf("Authentication failed: %v", err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	commons.RespondWithJSON(w, http.StatusOK, user)
}

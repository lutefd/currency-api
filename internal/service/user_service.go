package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Lutefd/challenge-bravo/internal/model"
	"github.com/Lutefd/challenge-bravo/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (model.User, error) {
	userDB, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userDB.ToUser(), nil
}

func (s *UserService) GetByAPIKey(ctx context.Context, apiKey string) (model.User, error) {
	userDB, err := s.userRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return userDB.ToUser(), nil
}

func (s *UserService) Create(ctx context.Context, username, password string) (model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.UserDB{
		ID:        uuid.New(),
		Username:  username,
		Password:  string(hashedPassword),
		Role:      model.RoleUser,
		APIKey:    generateAPIKey(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return model.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ToUser(), nil
}

func (s *UserService) Delete(ctx context.Context, username string) error {
	if err := s.userRepo.Delete(ctx, username); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *UserService) Authenticate(ctx context.Context, username, password string) (model.User, error) {
	userDB, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return model.User{}, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userDB.Password), []byte(password)); err != nil {
		return model.User{}, fmt.Errorf("invalid credentials")
	}

	return userDB.ToUser(), nil
}

func generateAPIKey() string {
	return uuid.New().String()
}

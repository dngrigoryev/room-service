package service

import (
	"context"
	"errors"
	"time"

	"room-service/internal/domain"
	"room-service/pkg/token"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type AuthService struct {
	userRepo     UserRepository
	tokenManager *token.Manager
}

func NewAuthService(userRepo UserRepository, tm *token.Manager) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenManager: tm,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, role string) (*domain.User, error) {
	if role != "admin" && role != "user" {
		return nil, errors.New("invalid role")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		Role:         role,
		PasswordHash: string(hash),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.tokenManager.Generate(user.ID, user.Role, 24*time.Hour)
}

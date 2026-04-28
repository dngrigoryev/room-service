package service

import (
	"context"
	"errors"
	"testing"

	"room-service/internal/domain"
	"room-service/pkg/token"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	createFunc     func(user *domain.User) error
	getByEmailFunc func(email string) (*domain.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.createFunc != nil {
		return m.createFunc(user)
	}
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(email)
	}
	return nil, errors.New("not found")
}

func TestAuthService_Register(t *testing.T) {
	repo := &mockUserRepo{
		createFunc: func(u *domain.User) error {
			if u.Email == "fail@test.com" {
				return errors.New("db error")
			}
			u.ID = "123"
			return nil
		},
	}
	tm := token.NewManager("secret")
	svc := NewAuthService(repo, tm)

	ctx := context.Background()

	t.Run("success user", func(t *testing.T) {
		u, err := svc.Register(ctx, "test@test.com", "pass123", "user")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if u.Email != "test@test.com" || u.Role != "user" {
			t.Errorf("unexpected user data: %+v", u)
		}
		if u.PasswordHash == "pass123" {
			t.Errorf("password was not hashed")
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		_, err := svc.Register(ctx, "test@test.com", "pass123", "hacker")
		if err == nil {
			t.Fatal("expected error for invalid role")
		}
	})

	t.Run("db error", func(t *testing.T) {
		_, err := svc.Register(ctx, "fail@test.com", "pass123", "admin")
		if err == nil {
			t.Fatal("expected db error")
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("valid_pass"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		getByEmailFunc: func(email string) (*domain.User, error) {
			if email == "exist@test.com" {
				return &domain.User{
					ID:           "user_id_1",
					Role:         "user",
					PasswordHash: string(hashedPassword),
				}, nil
			}
			return nil, errors.New("not found")
		},
	}
	tm := token.NewManager("my_test_secret")
	svc := NewAuthService(repo, tm)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		tok, err := svc.Login(ctx, "exist@test.com", "valid_pass")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tok == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		_, err := svc.Login(ctx, "exist@test.com", "wrong_pass")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Login(ctx, "notexist@test.com", "valid_pass")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

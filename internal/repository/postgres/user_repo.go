package postgres

import (
	"context"
	"errors"

	"room-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, role, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, user.Email, user.Role, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, role, password_hash, created_at FROM users WHERE email = $1`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.Email, &user.Role, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

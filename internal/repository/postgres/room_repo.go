package postgres

import (
	"context"
	"errors"

	"room-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoomRepo struct {
	db *pgxpool.Pool
}

func NewRoomRepo(db *pgxpool.Pool) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) Create(ctx context.Context, room *domain.Room) error {
	query := `INSERT INTO rooms (name, description, capacity) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, room.Name, room.Description, room.Capacity).Scan(&room.ID, &room.CreatedAt)
}

func (r *RoomRepo) List(ctx context.Context) ([]domain.Room, error) {
	query := `SELECT id, name, description, capacity, created_at FROM rooms ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var room domain.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.Capacity, &room.CreatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	if rooms == nil {
		rooms = []domain.Room{}
	}

	return rooms, nil
}

func (r *RoomRepo) Exists(ctx context.Context, id string) (bool, error) {
	var temp string
	err := r.db.QueryRow(ctx, "SELECT id FROM rooms WHERE id = $1", id).Scan(&temp)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

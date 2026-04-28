package postgres

import (
	"context"
	"errors"

	"room-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepo struct {
	db *pgxpool.Pool
}

func NewBookingRepo(db *pgxpool.Pool) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	query := `
		INSERT INTO bookings (slot_id, user_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(ctx, query, booking.SlotID, booking.UserID, booking.Status).
		Scan(&booking.ID, &booking.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.ErrSlotAlreadyBooked
			}
			if pgErr.Code == "23503" {
				return domain.ErrSlotNotFound
			}
		}
		return err
	}

	return nil
}

func (r *BookingRepo) List(ctx context.Context, limit, offset int) ([]domain.Booking, int, error) {
	var total int
	err := r.db.QueryRow(ctx, "SELECT count(*) FROM bookings").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, slot_id, user_id, status, created_at
		FROM bookings
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		result = append(result, b)
	}

	if result == nil {
		result = []domain.Booking{}
	}
	return result, total, nil
}

func (r *BookingRepo) ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error) {
	query := `
		SELECT b.id, b.slot_id, b.user_id, b.status, b.created_at
		FROM bookings b
		JOIN slots s ON b.slot_id = s.id
		WHERE b.user_id = $1 AND s.start_time > now()
		ORDER BY s.start_time ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, b)
	}

	if result == nil {
		result = []domain.Booking{}
	}
	return result, nil
}

func (r *BookingRepo) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	query := `SELECT id, slot_id, user_id, status, created_at FROM bookings WHERE id = $1`
	var b domain.Booking
	err := r.db.QueryRow(ctx, query, bookingID).Scan(&b.ID, &b.SlotID, &b.UserID, &b.Status, &b.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, bookingID, status string) error {
	query := `UPDATE bookings SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, bookingID)
	return err
}

package postgres

import (
	"context"
	"errors"

	"room-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleRepo struct {
	db *pgxpool.Pool
}

func NewScheduleRepo(db *pgxpool.Pool) *ScheduleRepo {
	return &ScheduleRepo{db: db}
}

func (r *ScheduleRepo) Create(ctx context.Context, schedule *domain.Schedule) error {
	query := `INSERT INTO schedules (room_id, days_of_week, start_time, end_time) 
			  VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.db.QueryRow(ctx, query, schedule.RoomID, schedule.DaysOfWeek, schedule.StartTime, schedule.EndTime).Scan(&schedule.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return domain.ErrScheduleExists
			}
			if pgErr.Code == "23503" {
				return domain.ErrRoomNotFound
			}
		}
		return err
	}

	return nil
}

func (r *ScheduleRepo) GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error) {
	query := `
		SELECT id, room_id, days_of_week, 
		       to_char(start_time, 'HH24:MI') as start_time, 
		       to_char(end_time, 'HH24:MI') as end_time
		FROM schedules
		WHERE room_id = $1
	`
	var s domain.Schedule
	err := r.db.QueryRow(ctx, query, roomID).Scan(&s.ID, &s.RoomID, &s.DaysOfWeek, &s.StartTime, &s.EndTime)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrScheduleNotFound
		}
		return nil, err
	}

	return &s, nil
}

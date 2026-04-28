package postgres

import (
	"context"
	"errors"
	"time"

	"room-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SlotRepo struct {
	db *pgxpool.Pool
}

func NewSlotRepo(db *pgxpool.Pool) *SlotRepo {
	return &SlotRepo{db: db}
}

func (r *SlotRepo) EnsureAndFetchAvailableSlots(ctx context.Context, roomID string, generatedSlots []domain.Slot, startOfDay, endOfDay time.Time) ([]domain.Slot, error) {
	if len(generatedSlots) > 0 {
		batch := &pgx.Batch{}
		for _, s := range generatedSlots {
			batch.Queue("INSERT INTO slots (room_id, start_time, end_time) VALUES ($1, $2, $3) ON CONFLICT (room_id, start_time) DO NOTHING",
				s.RoomID, s.Start, s.End)
		}
		br := r.db.SendBatch(ctx, batch)
		for i := 0; i < len(generatedSlots); i++ {
			if _, err := br.Exec(); err != nil {
				br.Close()
				return nil, err
			}
		}
		br.Close()
	}

	query := `
		SELECT s.id, s.room_id, s.start_time, s.end_time
		FROM slots s
		WHERE s.room_id = $1 
		  AND s.start_time >= $2 
		  AND s.end_time <= $3
		  AND NOT EXISTS (
			  SELECT 1 FROM bookings b 
			  WHERE b.slot_id = s.id AND b.status = 'active'
		  )
		ORDER BY s.start_time ASC
	`
	rows, err := r.db.Query(ctx, query, roomID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Slot
	for rows.Next() {
		var s domain.Slot
		if err := rows.Scan(&s.ID, &s.RoomID, &s.Start, &s.End); err != nil {
			return nil, err
		}
		s.Start = s.Start.UTC()
		s.End = s.End.UTC()
		result = append(result, s)
	}

	if result == nil {
		result = []domain.Slot{}
	}

	return result, nil
}

func (r *SlotRepo) GetByID(ctx context.Context, slotID string) (*domain.Slot, error) {
	query := `SELECT id, room_id, start_time, end_time FROM slots WHERE id = $1`
	var s domain.Slot
	err := r.db.QueryRow(ctx, query, slotID).Scan(&s.ID, &s.RoomID, &s.Start, &s.End)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSlotNotFound
		}
		return nil, err
	}
	s.Start = s.Start.UTC()
	s.End = s.End.UTC()
	return &s, nil
}

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"room-service/internal/domain"
)

type mockSlotRepo struct {
	ensureFunc func(roomID string, generatedSlots []domain.Slot, startOfDay, endOfDay time.Time) ([]domain.Slot, error)
}

func (m *mockSlotRepo) EnsureAndFetchAvailableSlots(ctx context.Context, roomID string, generatedSlots []domain.Slot, startOfDay, endOfDay time.Time) ([]domain.Slot, error) {
	if m.ensureFunc != nil {
		return m.ensureFunc(roomID, generatedSlots, startOfDay, endOfDay)
	}
	return nil, nil
}

type mockScheduleGetter struct {
	getByRoomIDFunc func(roomID string) (*domain.Schedule, error)
}

func (m *mockScheduleGetter) GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error) {
	if m.getByRoomIDFunc != nil {
		return m.getByRoomIDFunc(roomID)
	}
	return nil, errors.New("schedule not found")
}

type mockRoomChecker struct {
	existsFunc func(id string) (bool, error)
}

func (m *mockRoomChecker) Exists(ctx context.Context, id string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(id)
	}
	return false, nil
}

func TestSlotService_GenerateAvailableSlots(t *testing.T) {
	rc := &mockRoomChecker{
		existsFunc: func(id string) (bool, error) {
			if id == "err_room" {
				return false, errors.New("db error")
			}
			if id == "not_exist" {
				return false, nil
			}
			return true, nil
		},
	}
	sg := &mockScheduleGetter{
		getByRoomIDFunc: func(roomID string) (*domain.Schedule, error) {
			if roomID == "no_sched" {
				return nil, errors.New("no schedule")
			}
			return &domain.Schedule{
				RoomID:     roomID,
				DaysOfWeek: []int{1, 2, 3, 4, 5, 6, 7},
				StartTime:  "09:00",
				EndTime:    "10:00",
			}, nil
		},
	}
	sr := &mockSlotRepo{
		ensureFunc: func(roomID string, generatedSlots []domain.Slot, startOfDay, endOfDay time.Time) ([]domain.Slot, error) {
			return generatedSlots, nil
		},
	}

	svc := NewSlotService(sr, sg, rc)
	ctx := context.Background()

	t.Run("invalid date", func(t *testing.T) {
		_, err := svc.GenerateAvailableSlots(ctx, "room_1", "bad_date")
		if err == nil {
			t.Fatal("expected err")
		}
	})

	t.Run("room not exist", func(t *testing.T) {
		_, err := svc.GenerateAvailableSlots(ctx, "not_exist", "2025-01-01")
		if err != domain.ErrRoomNotFound {
			t.Fatalf("expected ErrRoomNotFound, got %v", err)
		}
	})

	t.Run("room check error", func(t *testing.T) {
		_, err := svc.GenerateAvailableSlots(ctx, "err_room", "2025-01-01")
		if err == nil {
			t.Fatal("expected err")
		}
	})

	t.Run("success valid generation", func(t *testing.T) {
		slots, err := svc.GenerateAvailableSlots(ctx, "room_1", "2025-01-01")
		if err != nil {
			t.Fatalf("expected no err, got %v", err)
		}
		if len(slots) != 2 {
			t.Errorf("expected 2 slots generated, got %d", len(slots))
		}
	})

	t.Run("holiday", func(t *testing.T) {
		sgHoliday := &mockScheduleGetter{
			getByRoomIDFunc: func(roomID string) (*domain.Schedule, error) {
				return &domain.Schedule{
					RoomID:     roomID,
					DaysOfWeek: []int{7},
					StartTime:  "09:00",
					EndTime:    "10:00",
				}, nil
			},
		}
		svcHoliday := NewSlotService(sr, sgHoliday, rc)
		slots, err := svcHoliday.GenerateAvailableSlots(ctx, "room_1", "2025-01-01")
		if err != nil {
			t.Fatalf("expected no err, got %v", err)
		}
		if len(slots) != 0 {
			t.Errorf("expected 0 slots on holiday, got %d", len(slots))
		}
	})
}

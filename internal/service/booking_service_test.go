package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"room-service/internal/domain"
)

type mockSlotGetter struct {
	getByIDFunc func(slotID string) (*domain.Slot, error)
}

func (m *mockSlotGetter) GetByID(ctx context.Context, slotID string) (*domain.Slot, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(slotID)
	}
	return nil, errors.New("slot not found")
}

type mockBookingRepo struct {
	createFunc           func(booking *domain.Booking) error
	listFunc             func(limit, offset int) ([]domain.Booking, int, error)
	listFutureByUserFunc func(userID string) ([]domain.Booking, error)
	getByIDFunc          func(bookingID string) (*domain.Booking, error)
	updateStatusFunc     func(bookingID, status string) error
}

func (m *mockBookingRepo) Create(ctx context.Context, booking *domain.Booking) error {
	if m.createFunc != nil {
		return m.createFunc(booking)
	}
	return nil
}
func (m *mockBookingRepo) List(ctx context.Context, limit, offset int) ([]domain.Booking, int, error) {
	if m.listFunc != nil {
		return m.listFunc(limit, offset)
	}
	return nil, 0, nil
}
func (m *mockBookingRepo) ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error) {
	if m.listFutureByUserFunc != nil {
		return m.listFutureByUserFunc(userID)
	}
	return nil, nil
}
func (m *mockBookingRepo) GetByID(ctx context.Context, bookingID string) (*domain.Booking, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(bookingID)
	}
	return nil, errors.New("booking not found")
}
func (m *mockBookingRepo) UpdateStatus(ctx context.Context, bookingID, status string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(bookingID, status)
	}
	return nil
}

func TestBookingService_Create(t *testing.T) {
	slotGetter := &mockSlotGetter{
		getByIDFunc: func(slotID string) (*domain.Slot, error) {
			if slotID == "past_slot" {
				return &domain.Slot{ID: "past_slot", Start: time.Now().Add(-2 * time.Hour)}, nil
			}
			if slotID == "future_slot" {
				return &domain.Slot{ID: "future_slot", Start: time.Now().Add(2 * time.Hour)}, nil
			}
			return nil, errors.New("slot repo error")
		},
	}

	repo := &mockBookingRepo{
		createFunc: func(b *domain.Booking) error {
			if b.SlotID == "err_slot" {
				return errors.New("db conflict")
			}
			b.ID = "new_booking_id"
			return nil
		},
	}

	svc := NewBookingService(repo, slotGetter)
	ctx := context.Background()

	t.Run("slot in past", func(t *testing.T) {
		_, err := svc.Create(ctx, "user1", "past_slot")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("slot repo error", func(t *testing.T) {
		_, err := svc.Create(ctx, "user1", "not_exist")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success create", func(t *testing.T) {
		b, err := svc.Create(ctx, "user1", "future_slot")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if b.ID != "new_booking_id" || b.Status != domain.BookingStatusActive {
			t.Errorf("unexpected booking data: %+v", b)
		}
	})
}

func TestBookingService_Cancel(t *testing.T) {
	repo := &mockBookingRepo{
		getByIDFunc: func(bookingID string) (*domain.Booking, error) {
			if bookingID == "b_for_user1" {
				return &domain.Booking{ID: "b_for_user1", UserID: "user1", Status: domain.BookingStatusActive}, nil
			}
			return nil, errors.New("not found")
		},
	}
	svc := NewBookingService(repo, &mockSlotGetter{})
	ctx := context.Background()

	t.Run("not owner", func(t *testing.T) {
		_, err := svc.Cancel(ctx, "b_for_user1", "user2")
		if err == nil {
			t.Fatal("expected error for wrong user")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := svc.Cancel(ctx, "other_b", "user1")
		if err == nil {
			t.Fatal("expected error for non-existent booking")
		}
	})

	t.Run("success cancel", func(t *testing.T) {
		b, err := svc.Cancel(ctx, "b_for_user1", "user1")
		if err != nil {
			t.Fatalf("expected no error: %v", err)
		}
		if b.Status != domain.BookingStatusCancelled {
			t.Errorf("status not changed, is %v", b.Status)
		}
	})
}

package service

import (
	"context"
	"time"

	"room-service/internal/domain"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	List(ctx context.Context, limit, offset int) ([]domain.Booking, int, error)
	ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error)
	GetByID(ctx context.Context, bookingID string) (*domain.Booking, error)
	UpdateStatus(ctx context.Context, bookingID, status string) error
}

type SlotGetter interface {
	GetByID(ctx context.Context, slotID string) (*domain.Slot, error)
}

type BookingService struct {
	repo     BookingRepository
	slotRepo SlotGetter
}

func NewBookingService(repo BookingRepository, slotRepo SlotGetter) *BookingService {
	return &BookingService{repo: repo, slotRepo: slotRepo}
}

func (s *BookingService) Create(ctx context.Context, userID, slotID string) (*domain.Booking, error) {
	slot, err := s.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		return nil, err
	}

	if time.Now().UTC().After(slot.Start) {
		return nil, domain.ErrSlotInPast
	}

	booking := &domain.Booking{
		SlotID: slotID,
		UserID: userID,
		Status: domain.BookingStatusActive,
	}

	if err := s.repo.Create(ctx, booking); err != nil {
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) ListAll(ctx context.Context, page, pageSize int) ([]domain.Booking, domain.Pagination, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	limit := pageSize
	offset := (page - 1) * pageSize

	bookings, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, domain.Pagination{}, err
	}

	pagination := domain.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}

	return bookings, pagination, nil
}

func (s *BookingService) ListFutureByUser(ctx context.Context, userID string) ([]domain.Booking, error) {
	return s.repo.ListFutureByUser(ctx, userID)
}

func (s *BookingService) Cancel(ctx context.Context, bookingID, userID string) (*domain.Booking, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}

	if booking.UserID != userID {
		return nil, domain.ErrForbidden
	}

	if err := s.repo.UpdateStatus(ctx, bookingID, domain.BookingStatusCancelled); err != nil {
		return nil, err
	}

	booking.Status = domain.BookingStatusCancelled
	return booking, nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"room-service/internal/domain"
)

type SlotRepository interface {
	EnsureAndFetchAvailableSlots(ctx context.Context, roomID string, generatedSlots []domain.Slot, startOfDay, endOfDay time.Time) ([]domain.Slot, error)
}

type scheduleGetter interface {
	GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error)
}
type roomChecker interface {
	Exists(ctx context.Context, id string) (bool, error)
}

type SlotService struct {
	slotRepo     SlotRepository
	scheduleRepo scheduleGetter
	roomRepo     roomChecker
}

func NewSlotService(slotRepo SlotRepository, scheduleRepo scheduleGetter, roomRepo roomChecker) *SlotService {
	return &SlotService{
		slotRepo:     slotRepo,
		scheduleRepo: scheduleRepo,
		roomRepo:     roomRepo,
	}
}

func (s *SlotService) GenerateAvailableSlots(ctx context.Context, roomID, dateStr string) ([]domain.Slot, error) {
	targetDate, err := time.ParseInLocation("2006-01-02", dateStr, time.UTC)
	if err != nil {
		return nil, errors.New("invalid date structure. Expected YYYY-MM-DD")
	}

	roomExists, err := s.roomRepo.Exists(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("error checking room existence: %w", err)
	}
	if !roomExists {
		return nil, domain.ErrRoomNotFound
	}

	sched, err := s.scheduleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		if errors.Is(err, domain.ErrScheduleNotFound) {
			return []domain.Slot{}, nil
		}
		return nil, fmt.Errorf("failed to fetch schedule: %w", err)
	}

	goWeekday := int(targetDate.Weekday())
	parsedDay := goWeekday
	if goWeekday == 0 {
		parsedDay = 7
	}

	dayAllowed := false
	for _, allowedDay := range sched.DaysOfWeek {
		if allowedDay == parsedDay {
			dayAllowed = true
			break
		}
	}

	startOfDay := targetDate
	endOfDay := targetDate.Add(24 * time.Hour).Add(-time.Nanosecond)

	var generatedSlots []domain.Slot

	if dayAllowed {
		startTimeObj, _ := time.Parse("15:04", sched.StartTime)
		endTimeObj, _ := time.Parse("15:04", sched.EndTime)

		current := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), startTimeObj.Hour(), startTimeObj.Minute(), 0, 0, time.UTC)
		end := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), endTimeObj.Hour(), endTimeObj.Minute(), 0, 0, time.UTC)

		if end.Before(current) {
			end = end.Add(24 * time.Hour)
		}

		for current.Before(end) {
			next := current.Add(30 * time.Minute)
			if next.After(end) {
				break
			}
			generatedSlots = append(generatedSlots, domain.Slot{
				RoomID: roomID,
				Start:  current,
				End:    next,
			})
			current = next
		}
	}

	available, err := s.slotRepo.EnsureAndFetchAvailableSlots(ctx, roomID, generatedSlots, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("EnsureAndFetchAvailableSlots failed: %w", err)
	}

	return available, nil
}

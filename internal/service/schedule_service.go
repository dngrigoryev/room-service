package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"room-service/internal/domain"
)

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.Schedule) error
}

type ScheduleService struct {
	repo ScheduleRepository
}

func NewScheduleService(repo ScheduleRepository) *ScheduleService {
	return &ScheduleService{repo: repo}
}

func (s *ScheduleService) Create(ctx context.Context, roomID string, daysOfWeek []int, startTime, endTime string) (*domain.Schedule, error) {
	if len(daysOfWeek) == 0 {
		return nil, errors.New("daysOfWeek must not be empty")
	}

	for _, day := range daysOfWeek {
		if day < 1 || day > 7 {
			return nil, fmt.Errorf("invalid day %d: must be between 1 and 7", day)
		}
	}

	parsedStart, errStart := time.Parse("15:04", startTime)
	if errStart != nil {
		return nil, errors.New("invalid start time format, expected HH:MM")
	}

	parsedEnd, errEnd := time.Parse("15:04", endTime)
	if errEnd != nil {
		return nil, errors.New("invalid end time format, expected HH:MM")
	}

	if !parsedStart.Before(parsedEnd) {
		return nil, errors.New("startTime must be before endTime")
	}

	schedule := &domain.Schedule{
		RoomID:     roomID,
		DaysOfWeek: daysOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	if err := s.repo.Create(ctx, schedule); err != nil {
		return nil, err
	}

	return schedule, nil
}

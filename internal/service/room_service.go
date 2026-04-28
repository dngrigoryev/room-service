package service

import (
	"context"
	"errors"

	"room-service/internal/domain"
)

type RoomRepository interface {
	Create(ctx context.Context, room *domain.Room) error
	List(ctx context.Context) ([]domain.Room, error)
}

type RoomService struct {
	repo RoomRepository
}

func NewRoomService(repo RoomRepository) *RoomService {
	return &RoomService{repo: repo}
}

func (s *RoomService) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*domain.Room, error) {
	if name == "" {
		return nil, errors.New("room name is required")
	}

	room := &domain.Room{
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}

	if err := s.repo.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *RoomService) ListRooms(ctx context.Context) ([]domain.Room, error) {
	return s.repo.List(ctx)
}

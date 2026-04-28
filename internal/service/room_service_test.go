package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"room-service/internal/domain"
)

type mockRoomRepo struct {
	createFunc func(room *domain.Room) error
	listFunc   func() ([]domain.Room, error)
}

func (m *mockRoomRepo) Create(ctx context.Context, room *domain.Room) error {
	if m.createFunc != nil {
		return m.createFunc(room)
	}
	return nil
}

func (m *mockRoomRepo) List(ctx context.Context) ([]domain.Room, error) {
	if m.listFunc != nil {
		return m.listFunc()
	}
	return nil, nil
}

func TestRoomService_Create(t *testing.T) {
	repo := &mockRoomRepo{
		createFunc: func(r *domain.Room) error {
			if r.Name == "fail" {
				return errors.New("db error")
			}
			r.ID = "room_1"
			now := time.Now()
			r.CreatedAt = &now
			return nil
		},
	}
	svc := NewRoomService(repo)
	ctx := context.Background()

	t.Run("success create", func(t *testing.T) {
		desc := "Desc"
		cap := 10
		r, err := svc.CreateRoom(ctx, "Test Room", &desc, &cap)
		if err != nil {
			t.Fatalf("expected no err, got %v", err)
		}
		if r.ID != "room_1" {
			t.Errorf("unexpected ID %s", r.ID)
		}
	})

	t.Run("fail create", func(t *testing.T) {
		desc := "Desc"
		cap := 10
		_, err := svc.CreateRoom(ctx, "fail", &desc, &cap)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestRoomService_List(t *testing.T) {
	repo := &mockRoomRepo{
		listFunc: func() ([]domain.Room, error) {
			return []domain.Room{
				{ID: "r1", Name: "Room 1"},
				{ID: "r2", Name: "Room 2"},
			}, nil
		},
	}
	svc := NewRoomService(repo)
	ctx := context.Background()

	t.Run("list all", func(t *testing.T) {
		rooms, err := svc.ListRooms(ctx)
		if err != nil {
			t.Fatalf("expected no err, got %v", err)
		}
		if len(rooms) != 2 {
			t.Errorf("expected 2 rooms, got %d", len(rooms))
		}
	})
}

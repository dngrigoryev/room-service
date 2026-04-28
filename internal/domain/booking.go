package domain

import "time"

const (
	BookingStatusActive    = "active"
	BookingStatusCancelled = "cancelled"
)

type Booking struct {
	ID        string     `json:"id"`
	SlotID    string     `json:"slotId"`
	UserID    string     `json:"userId"`
	Status    string     `json:"status"`
	CreatedAt *time.Time `json:"createdAt"`
}

type Pagination struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

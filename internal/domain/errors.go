package domain

import "errors"

var (
	ErrRoomNotFound      = errors.New("room not found")
	ErrScheduleExists    = errors.New("schedule already exists")
	ErrScheduleNotFound  = errors.New("schedule not found")
	ErrSlotNotFound      = errors.New("slot not found")
	ErrSlotAlreadyBooked = errors.New("slot already booked")
	ErrBookingNotFound   = errors.New("booking not found")
	ErrSlotInPast        = errors.New("cannot book a slot in the past")
	ErrForbidden         = errors.New("forbidden")
)

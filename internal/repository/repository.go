package repository

import (
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res Models.Reservation) (int, error)
	InsertRoomRestriction(r Models.RoomRestriction) error
	SearchAvailabilityByDate(start, end time.Time, roomID int) (bool, error)
}
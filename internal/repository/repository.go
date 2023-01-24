package repository

import (
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res Models.Reservation) (int, error)
	InsertRoomRestriction(r Models.RoomRestriction) error
	SearchAvailabilityByDateByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]Models.Room, error)
	GetRoomByID(id int) (Models.Room, error)

	GetUserByID(id int) (Models.User, error)
	UpdateUser(u Models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	AllReservations() ([]Models.Reservation, error)
}

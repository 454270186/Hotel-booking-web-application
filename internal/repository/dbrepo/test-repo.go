package dbrepo

import (
	"github.com/454270186/Hotel-booking-web-application/internal/Models"
	"time"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into database
func (m *testDBRepo) InsertReservation(res Models.Reservation) (int, error) {
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into database
func (m *testDBRepo) InsertRoomRestriction(r Models.RoomRestriction) error {
	return nil
}

// SearchAvailabilityByDate returns true if availability exists for roomID and false if no availability
func (m *testDBRepo) SearchAvailabilityByDateByRoomID(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]Models.Room, error) {
	var rooms []Models.Room

	return rooms, nil
}

// GetRoomByID gets a room by given id
func (m *testDBRepo) GetRoomByID(id int) (Models.Room, error) {
	var room Models.Room

	return room, nil
}

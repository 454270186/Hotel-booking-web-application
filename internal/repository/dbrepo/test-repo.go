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

func (m *testDBRepo) GetUserByID(id int) (Models.User, error) {
	var u Models.User

	return u, nil
}

func (m *testDBRepo) UpdateUser(u Models.User) error {
	return nil
}

func (m *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 0, "", nil
}

// AllReservations returns a slice of all reservations
func (m *testDBRepo) AllReservations() ([]Models.Reservation, error) {
	var reservations []Models.Reservation

	return reservations, nil
}

// AllNewReservations returns all new reservations
func (m *testDBRepo) AllNewReservations() ([]Models.Reservation, error) {

	var reservations []Models.Reservation

	return reservations, nil
}

// GetReservationByID returns reservation by given ID
func (m *testDBRepo) GetReservationByID(id int) (Models.Reservation, error) {

	var res Models.Reservation

	return res, nil
}

// UpdateReservation updates a reservation in database
func (m *testDBRepo) UpdateReservation(res Models.Reservation) error {

	return nil
}

// DeleteReservation deletes a reservation by given ID
func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

// UpdateProcessedForReservation updates processed for a reservation
func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	return nil
}

// AllRooms returns a slice of all rooms
func (m *testDBRepo) AllRooms() ([]Models.Room, error) {
	var rooms []Models.Room

	return rooms, nil
}

// GetRestrictionsForRoomByDate returns restrictions for a room by date range
func (m *testDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]Models.RoomRestriction, error) {
	var roomRestrictions []Models.RoomRestriction

	return roomRestrictions, nil
}

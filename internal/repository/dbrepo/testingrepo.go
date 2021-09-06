package dbrepo

import (
	"errors"
	"time"

	"github.com/tsawler/bookings-app/internal/models"
)

func (m *testingDBRepo) AllUsers() bool {
	return true
}

func (m *testingDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// if the room id is 2, then fail; otherwise, pass
	if res.RoomID == 2 {
		return 0, errors.New("room 2 is never available!)")
	}
	return 1, nil

}

func (m *testingDBRepo) InsertRoomRestriction(res models.RoomRestriction) (int, error) {
	if res.RoomID == 1000 {
		return 0, errors.New("room 1000 can never be restricted")
	}
	return 1, nil
}

func (m *testingDBRepo) SearchAvailabilityByDatesForRoom(start, end time.Time, roomID int) (bool, error) {
	return true, nil
}

func (m *testingDBRepo) SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error) {
	rooms := []models.Room{}
	return rooms, nil
}

// GetRoom returns a room model. If no room is found, the error will be sql.ErrNoRows.
func (m *testingDBRepo) GetRoom(id int) (models.Room, error) {
	room := models.Room{
		ID:       id,
		RoomName: "Place of Mocking",
	}
	return room, nil
}

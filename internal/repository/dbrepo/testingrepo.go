package dbrepo

import (
	"errors"
	"log"
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
	// if the start date is after 2049-12-31, then not available,
	layout := "2006-01-02"
	str := "2049-12-31"
	noRoomDate, err := time.Parse(layout, str)
	if err != nil {
		log.Println(err)
	}

	testDateToFail, err := time.Parse(layout, "2060-01-01")
	if err != nil {
		log.Println(err)
	}

	if start == testDateToFail {
		return false, errors.New("mock fails db lookup")
	}

	if start.After(noRoomDate) {
		return false, nil
	}

	return true, nil
}

func (m *testingDBRepo) SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	// if the start date is after 2049-12-31, then return empty slice,
	// indicating no rooms are available;
	layout := "2006-01-02"
	str := "2049-12-31"
	t, err := time.Parse(layout, str)
	if err != nil {
		log.Println(err)
	}

	testDateToFail, err := time.Parse(layout, "2060-01-01")
	if err != nil {
		log.Println(err)
	}

	if start == testDateToFail {
		return rooms, errors.New("some error")
	}

	if start.After(t) {
		return rooms, nil
	}

	// otherwise, put an entry into the slice, indicating that some room is
	// available for search dates
	room := models.Room{
		ID:       1,
		RoomName: "Mock Room",
	}
	rooms = append(rooms, room)

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

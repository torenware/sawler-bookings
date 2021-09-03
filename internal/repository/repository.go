package repository

import (
	"time"

	"github.com/tsawler/bookings-app/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool

	GetRoom(id int) (models.Room, error)

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(res models.RoomRestriction) (int, error)

	SearchAvailabilityByDatesForRoom(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error)
}

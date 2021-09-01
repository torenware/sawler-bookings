package dbrepo

import (
	"context"
	"time"

	"github.com/tsawler/bookings-app/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	var newID int

	stmt := `
	  insert into reservations(
		  first_name, last_name, email, phone,
		  start_date, end_date, room_id,
		  created_at, updated_at
	  )
	  values (
		  $1, $2, $3, $4,
		  $5, $6, $7,
		  now(), now()
	  )
	  returning id
	`
	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		res.FirstName, res.LastName, res.Email, res.Phone,
		res.StartDate, res.EndDate, res.RoomID,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *postgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()
	var newID int

	stmt := `
	  insert into room_restrictions(
		  start_date, end_date, room_id,
		  reservation_id, restriction_id,
		  created_at, updated_at
	  )
	  values (
		  $1, $2, $3,
		  $4, $5,
		  now(), now()
	  )
	  returning id
	`
	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		res.StartDate, res.EndDate, res.RoomID,
		res.ReservationID, res.RestrictionID,
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

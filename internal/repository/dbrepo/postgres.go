package dbrepo

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/tsawler/bookings-app/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// tmp: write out our input
	output, _ := json.MarshalIndent(res, "", "    ")
	log.Println(string(output))
	// end debug
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

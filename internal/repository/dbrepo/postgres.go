package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
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

func (m *postgresDBRepo) SearchAvailabilityByDatesForRoom(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `
	  select count(*) from room_restrictions
	  where (
		start_date >= $1 and start_date < $2 or
		end_date > $1 and end_date <= $2
	  )
	  and room_id = $3
	`

	err := m.DB.QueryRowContext(
		ctx, query,
		start,
		end,
		roomID,
	).Scan(&numRows)

	if err != nil {
		return false, err
	}

	return numRows == 0, nil

}

func (m *postgresDBRepo) SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	  select r.id, r.room_name from rooms r
	  where r.id not in (
		select rr.room_id from room_restrictions rr
		where
		  rr.start_date >= $1 and rr.start_date < $2 or
          rr.end_date > $1 and rr.end_date <= $2
	  )
	`

	results := []models.Room{}

	rows, err := m.DB.QueryContext(
		ctx, query,
		start,
		end,
	)

	if err != nil {
		return results, err
	}

	for rows.Next() {
		var rowID int
		var rowName string
		if err := rows.Scan(&rowID, &rowName); err != nil {
			return []models.Room{}, err
		}
		row := models.Room{
			ID:       rowID,
			RoomName: rowName,
		}
		results = append(results, row)
	}

	return results, nil
}

// GetRoom returns a room model. If no room is found, the error will be sql.ErrNoRows.
func (m *postgresDBRepo) GetRoom(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select room_name from rooms where id = $1`
	row := m.DB.QueryRowContext(ctx, query, id)
	var roomName string
	var room models.Room
	err := row.Scan(&roomName)
	if err != nil {
		return room, err
	}
	room.ID = id
	room.RoomName = roomName
	return room, nil
}

// GetUserByID returns a user record by ID.
func (m *postgresDBRepo) GetUserByID(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var u models.User
	u.ID = id
	query := `
	select first_name, last_name, 
			email, access_level, 
			created_at, updated_at
	where id = $1
`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
			&u.FirstName, &u.LastName,
			&u.Email, &u.AccessLevel,
			&u.CreatedAt, &u.UpdatedAt,
		)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		return &u, nil
	}
}

func (m *postgresDBRepo) UpdateUser(u *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	update users 
		set 
			first_name = $1, last_name = $2,
			email = $3, access_level = $4,
			updated_at = now()
		where id = $5
`
	_, err := m.DB.ExecContext(
		ctx, stmt,
		u.FirstName, u.LastName,
		u.Email, u.AccessLevel,
		u.ID,
		)
	return err
}

// Authenticate compares a supplied password with the hashed stored pw.
// returns if authenticated, the hashed pw, and an error.
func (m *postgresDBRepo) Authenticate(email, password string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var pw string
	query := `
	select id, password from users where email = $1
`
	row := m.DB.QueryRowContext(
		ctx, query, email,
		)
	err := row.Scan(&id, &pw)
	if err != nil {
		return id, "", err
	}
	err = bcrypt.CompareHashAndPassword([]byte(pw), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return id, "", errors.New("Incorrect password")
	} else if err != nil {
		return 0, "", err
	}
	return id, pw, nil
}
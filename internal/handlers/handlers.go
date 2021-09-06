package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/tsawler/bookings-app/internal/config"
	"github.com/tsawler/bookings-app/internal/driver"
	"github.com/tsawler/bookings-app/internal/forms"
	"github.com/tsawler/bookings-app/internal/helpers"
	"github.com/tsawler/bookings-app/internal/models"
	"github.com/tsawler/bookings-app/internal/render"
	"github.com/tsawler/bookings-app/internal/repository"
	"github.com/tsawler/bookings-app/internal/repository/dbrepo"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// NewTestRepo returns the repo struct used for test mocks.
func NewTestRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewTestingRepo(a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "about.page.tmpl", &models.TemplateData{})
}

// ChooseRoom takes a link to a room and sets up the reservation form.
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		// really this should be a 404...
		helpers.ServerError(w, err)
		return
	}
	data := m.App.Session.Get(r.Context(), "reservation")
	var res models.Reservation
	if _, ok := data.(models.Reservation); !ok {
		helpers.ServerError(w, errors.New("could not get res data from session"))
		return
	} else {
		res, _ = data.(models.Reservation)
	}

	room, err := m.DB.GetRoom(roomID)
	if err == sql.ErrNoRows {
		// 404 error
		helpers.ClientError(w, http.StatusNotFound)
		return
	} else if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.RoomID = roomID
	res.Room = room
	log.Println("Res before make res form:")
	helpers.PrintStruct(res)
	m.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	sessData := m.App.Session.Get(r.Context(), "reservation")
	var res models.Reservation
	if _, ok := sessData.(models.Reservation); !ok {
		m.App.Session.Put(r.Context(), "error", "could not get data out of session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	res = sessData.(models.Reservation)

	data := make(map[string]interface{})
	data["reservation"] = res

	render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "could not parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.Session.Put(r.Context(), "error", "could not get res out of session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	if !form.Valid() {
		data := make(map[string]interface{})
		http.Error(w, "my own error message", http.StatusSeeOther)
		data["reservation"] = reservation
		render.Template(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "could not create reservation")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	reservation.ID = newID

	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: reservation.ID,
		RestrictionID: 1, // temp
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	_, err = m.DB.InsertRoomRestriction(restriction)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "could not restrict these dates")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// PostAvailability handles post
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	log.Println("form:", r.Form)
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	st, err := time.Parse("2006-01-02", start)
	if err != nil {
		m.App.ErrorLog.Println("Problem with your form data")
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "start date was garbled")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	en, err := time.Parse("2006-01-02", end)
	if err != nil {
		m.App.ErrorLog.Println("Problem with your form data")
		m.App.Session.Put(r.Context(), "error", "end date was garbled")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Now get our options
	rooms, err := m.DB.SearchAvailabilityByDates(st, en)
	if err != nil {
		m.App.ErrorLog.Println("Problem with your form data")
		msg, _ := fmt.Printf("Search query returned error: %s", err)
		m.App.Session.Put(r.Context(), "error", msg)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// If no rooms, show an error to the user on the SA page.
	if len(rooms) == 0 {
		log.Println("No room available for these dates")
		m.App.Session.Put(r.Context(), "error", "No rooms available for these dates")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	// Put up a page with search results.
	data := make(map[string]interface{})
	data["rooms"] = rooms
	data["start_date"] = st
	data["end_date"] = en

	// Build a reservation structure and fill in what we know so far.
	reservation := models.Reservation{
		StartDate: st,
		EndDate:   en,
	}
	m.App.Session.Put(r.Context(), "reservation", reservation)

	render.Template(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	room := r.Form.Get("room_id")

	room_id, err := strconv.Atoi(room)
	if err != nil {
		helpers.ServerError(w, errors.New("room_id must be supplied"))
		return
	}

	st, err := time.Parse("2006-01-02", start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	en, err := time.Parse("2006-01-02", end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	available, err := m.DB.SearchAvailabilityByDatesForRoom(st, en, room_id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	var message string
	if available {
		message = "Available"
	} else {
		message = "Not available for these dates"
	}
	resp := jsonResponse{
		OK:      available,
		Message: message,
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the res summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("Can't get reservation from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	helpers.PrintStruct(reservation)

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

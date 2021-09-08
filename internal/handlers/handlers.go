package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// BookRoom bridges the room pages to the Reservation form.
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	rm := r.URL.Query().Get("id")
	start := r.URL.Query().Get("s")
	end := r.URL.Query().Get("e")

	room_id, err := strconv.Atoi(rm)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "room number is invalid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	layout := "2006-01-02"
	start_date, err := time.Parse(layout, start)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "start date is invalid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	end_date, err := time.Parse(layout, end)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "end date is invalid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := m.DB.GetRoom(room_id)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "requested room does not exist")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation := models.Reservation{
	  RoomID: room_id,
	  StartDate: start_date,
	  EndDate: end_date,
	  Room: room,
	}
	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

// ChooseRoom takes a link to a room and sets up the reservation form.
func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.URL.Path, "/")
	roomID, err := strconv.Atoi(exploded[2])
	if err != nil {
		m.App.Session.Put(r.Context(), "error", "missing url parameter")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	data := m.App.Session.Get(r.Context(), "reservation")
	var res models.Reservation
	if _, ok := data.(models.Reservation); !ok {
		m.App.Session.Put(r.Context(), "error", "could not get reservation out of session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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
		m.App.Session.Put(r.Context(), "error", "could not fetch room for this ID")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.RoomID = roomID
	res.Room = room
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
	err := r.ParseForm()
	if err != nil {
		m.App.ErrorLog.Println("Could not parse your form")
		log.Println(err)
		m.App.Session.Put(r.Context(), "error", "problem parsing the form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
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
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func generateJSONError(w http.ResponseWriter, err error, retcode int) {
	info := jsonResponse{
		OK:      false,
		Message: err.Error(),
	}
	json, _ := json.MarshalIndent(info, "", "    ")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(retcode)
	w.Write(json)
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		generateJSONError(w, err, http.StatusBadRequest)
		return
	}
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	room := r.Form.Get("room_id")

	room_id, err := strconv.Atoi(room)
	if err != nil {
		generateJSONError(w, errors.New("room_id must be supplied"), http.StatusBadRequest)
		return
	}

	st, err := time.Parse("2006-01-02", start)
	if err != nil {
		generateJSONError(w, errors.New("start date is invalid"), http.StatusBadRequest)
		return
	}

	en, err := time.Parse("2006-01-02", end)
	if err != nil {
		generateJSONError(w, errors.New("end date is invalid"), http.StatusBadRequest)
		return
	}

	available, err := m.DB.SearchAvailabilityByDatesForRoom(st, en, room_id)
	if err != nil {
		generateJSONError(w, errors.New("database lookup failed"), http.StatusBadRequest)
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
		StartDate: start,
		EndDate: end,
		RoomID: strconv.Itoa(room_id),
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		generateJSONError(w, errors.New("could not encode result"), http.StatusInternalServerError)
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

	// helpers.PrintStruct(reservation)

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Template(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

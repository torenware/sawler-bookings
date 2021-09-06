package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"
	"github.com/tsawler/bookings-app/internal/config"
	"github.com/tsawler/bookings-app/internal/driver"
	"github.com/tsawler/bookings-app/internal/handlers"
	"github.com/tsawler/bookings-app/internal/helpers"
	"github.com/tsawler/bookings-app/internal/models"
	"github.com/tsawler/bookings-app/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

// main is the main function
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Not loading env from dot file:", err)
	}
	err = run()
	if err != nil {
		log.Fatal(err)
	}

	db, err := StartDB()
	if err != nil {
		log.Fatal("Cannot connect to database! Dying...")
	}
	defer db.SQL.Close()
	InitRepo(&app, db)

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})

	// change this to true when in production
	app.InProduction = false

	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache", err.Error())
		return err
	}

	app.TemplateCache = tc
	app.UseCache = false

	return nil
}

// StartDB factors out the database startup code, so we can mock it later.
func StartDB() (*driver.DB, error) {
	// connect to database
	log.Println("Connecting to database...")
	db, err := driver.ConnectSQL(driver.BuildDSN())
	if err != nil {
		return nil, err
	}

	log.Println("Connected to database!")
	return db, nil
}

// InitRepo sets up various dependencies of the repo.
func InitRepo(appPtr *config.AppConfig, db *driver.DB) {
	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app, nil, nil)
}

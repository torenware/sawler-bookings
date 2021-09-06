package dbrepo

import (
	"database/sql"

	"github.com/tsawler/bookings-app/internal/config"
	"github.com/tsawler/bookings-app/internal/repository"
)

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

type testingDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresRepo returns a repo for our handler functions.
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgresDBRepo{
		App: a,
		DB:  conn,
	}
}

// NewTestingRepo returns a mock repo for handlers.
func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
	return &testingDBRepo{
		App: a,
	}
}

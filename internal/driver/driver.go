package driver

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// DB holds the database connection pool
type DB struct {
	SQL *sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxDbLifetime = 5 * time.Minute


// testDB makes sure the DB is actually live
func testDB(db *sql.DB) error {
	err := db.Ping()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// BuildDSN pulls data from the environment to build the DSN
func BuildDSN() string {
	envItems := []string{
		"DB_NAME",
		"DB_USER",
		"DB_PASSWD",
		"DB_HOST",
	}
	params := make(map[string]string)
	for _, key := range envItems {
		val := os.Getenv(key)
		if val == "" {
			panic(fmt.Sprintf("The environment variable %s must be set", key))
		}
		params[key] = val
	}

	return fmt.Sprintf(
		"host=%s port=5432 dbname=%s user=%s password=%s",
		params["DB_HOST"],
		params["DB_NAME"],
		params["DB_USER"],
		params["DB_PASSWD"],
	)
}

// ConnectSQL creates database pool for Postgres
func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxDbLifetime)

	if err = testDB(d); err != nil {
		return nil, err
	}

	dbConn.SQL = d
	return dbConn, nil
}

func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return db, nil
}

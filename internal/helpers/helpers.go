package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"

	"github.com/tsawler/bookings-app/internal/config"
)

var app *config.AppConfig

// NewHelpers links the package with the config object
func NewHelpers(ac *config.AppConfig, infoW io.Writer, errW io.Writer) {
	app = ac
	if infoW == nil {
		infoW = os.Stdout
	}
	app.InfoLog = log.New(infoW, "INFO\t", log.Ltime|log.Lshortfile)

	if errW == nil {
		errW = os.Stderr
	}
	app.ErrorLog = log.New(errW, "ERROR\t", log.Ltime|log.Llongfile)
}

// ClientError logs a client error
func ClientError(w http.ResponseWriter, code int) {
	app.InfoLog.Printf("%s STATUS=%d", http.StatusText(code), code)
	http.Error(w, http.StatusText(code), code)
}

// ServerError logs an internal error
func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// PrintStruct turns a struct into json and prints it.
func PrintStruct(item interface{}) {
	data, _ := json.MarshalIndent(item, "", "    ")
	fmt.Println(string(data))
}

// IsAuthed checks for authentication
func IsAuthed(r *http.Request) bool {
	exists := app.Session.Exists(r.Context(), "user_id")
	return exists
}

package helpers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/tsawler/bookings-app/internal/config"
)

// @see https://stackoverflow.com/questions/10473800/in-go-how-do-i-capture-stdout-of-a-function-into-a-string
func captureStdOut(proc func(), havErr bool) string {
	old := os.Stdout // keep backup of the real stdout
	oldSE := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	if havErr {
		os.Stderr = w
	}
	proc()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	if havErr {
		os.Stderr = oldSE
	}
	out := <-outC
	return out
}

// Test default loggers do not cause an error
func TestDefaultLoggers(t *testing.T) {
	recorder := httptest.NewRecorder()
	runLogFunc := func() {
		appObj := config.AppConfig{}
		NewHelpers(&appObj, nil, nil)
		ClientError(recorder, 400)
	}
	fmt.Println("Entering capture")
	out := captureStdOut(runLogFunc, false)
	fmt.Println("exiting capture", out)
	if !strings.Contains(out, "INFO\t") {
		t.Error("infolog entry should contain INFO prefix")
	}
	if !strings.Contains(out, "Bad Request") {
		t.Error("infolog entry should contain Bad Request since 400")
	}

	// Did we pass the right info back?
	clientReply := strings.TrimSpace(recorder.Body.String())
	if clientReply != "Bad Request" {
		t.Error("Reply body sent to client should match 400 code text Bad Reply")
	}
}

func TestSubstituteBufferForLog(t *testing.T) {
	buf := new(bytes.Buffer)
	recorder := httptest.NewRecorder()
	appObj := config.AppConfig{}
	NewHelpers(&appObj, buf, nil)
	ClientError(recorder, 400)
	rslt := buf.String()
	if !strings.Contains(rslt, "INFO") {
		t.Error("log entry did not have INFO tag")
	}
}

func TestDefaultErrorLog(t *testing.T) {
	custErr := "oops did it again"
	recorder := httptest.NewRecorder()
	runLogFunc := func() {
		appObj := config.AppConfig{}
		NewHelpers(&appObj, nil, nil)
		err := errors.New(custErr)
		ServerError(recorder, err)
	}
	errText := captureStdOut(runLogFunc, true)
	if errText == "" {
		t.Error("we expected an error here")
	}
	if !strings.Contains(errText, "ERROR\t") {
		t.Error("errorlog entry should contain Error prefix")
	}
	if !strings.Contains(errText, custErr) {
		fmt.Println(errText)
		t.Error("errorlog entry should contain custom error text")
	}

	// And see that we do not leak stuff to client.
	if !strings.Contains(recorder.Body.String(), http.StatusText(http.StatusInternalServerError)) {
		t.Error("Expected internal server error message")
	}

}

func TestSubstituteBufferForErrors(t *testing.T) {
	buf := new(bytes.Buffer)
	recorder := httptest.NewRecorder()
	appObj := config.AppConfig{}
	NewHelpers(&appObj, nil, buf)
	errTxt := "we go boom"
	err := errors.New(errTxt)
	ServerError(recorder, err)
	rslt := buf.String()
	if !strings.Contains(rslt, errTxt) {
		t.Error("error entry did not have expected error txt in its buffer")
	}
}

package forms

import "testing"

func TestErrors(t *testing.T) {
	errs := errors{}

	// No error, won't find one
	err := errs.Get("field")
	if err != "" {
		t.Error("Get found an error when there should be none")
	}

	errs.Add("field", "oops")
	err = errs.Get("field")
	if err != "oops" {
		t.Error("Get did not get back oops as expected")
	}
}

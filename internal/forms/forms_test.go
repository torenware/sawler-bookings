package forms

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

// An empty form should be valid
func TestForm_Valid(t *testing.T) {
	request := httptest.NewRequest("POST", "/goobar", nil)
	form := New(request.PostForm)
	if !form.Valid() {
		t.Error("Empty form should be Valid")
	}
}

func TestForm_Required(t *testing.T) {
	request := httptest.NewRequest("POST", "/goobar", nil)
	form := New(request.PostForm)

	form.Required("a", "b", "c")
	// No such fields, so this should fail.
	if form.Valid() {
		t.Error("expected validation to fail since fields are not present")
	}

	// create a form with these values...
	postedData := url.Values{}
	postedData.Add("a", "apple")
	postedData.Add("b", "basket")
	postedData.Add("c", "cinnamon")

	request = httptest.NewRequest("POST", "/goobar", nil)
	request.PostForm = postedData

	form = New(request.PostForm)
	form.Required("a", "b", "c")
	// fields present, so this should succeed.
	if !form.Valid() {
		t.Error("expected validation to succeed since fields are present")
	}

}

// Test the Has method on *form
func TestForm_Has(t *testing.T) {
	values := url.Values{}
	form := New(values)

	// We have not added any items to the form data, so has should not find a field.
	if form.Has("afield") {
		t.Error("Has detected a field when it should not")
	}

	postdata := url.Values{
		"afield":   []string{"exists"},
		"tooshort": []string{},
	}

	form = New(postdata)
	if !form.Has("afield") {
		t.Error("Expected form Has afield, yet it does not")
	}
}

// Test minlength validator
func TestForm_MinLength(t *testing.T) {
	values := url.Values{}
	form := New(values)

	// We have not added any items to the form data, so minlegth should fail.
	if form.MinLength("afield", 3) {
		t.Error("MinLegth detected and passed a field when it should not")
	}

	postdata := url.Values{
		"afield":   []string{"exists"},
		"tooshort": []string{"ts"},
	}

	form = New(postdata)

	// Check afield, which has more than 3 chars. Should pass.
	if !form.MinLength("afield", 3) {
		t.Error("afield has more than 3 chars, should pass but did not")
	}

	// Check tooshort, which has less than 3 chars. Should pass.
	if form.MinLength("tooshort", 3) {
		t.Error("tooshort has less than 3 chars, should not pass but did")
	}

	// Form should now be invalid as well.
	if form.Valid() {
		t.Error("after invalid input, form should be invalid")
	}

}

// Check isEmail validator.
func TestForm_IsEmail(t *testing.T) {
	request := httptest.NewRequest("POST", "/goobar", nil)
	form := New(request.PostForm)

	// We have not added any items to the form data, so minlegth should fail.
	form.IsEmail("email")
	if form.Valid() {
		t.Error("email to isEmail detected and passed a field when it should not")
	}

	postdata := url.Values{
		"email":    []string{"exists@decarte.org"},
		"notemail": []string{"ts"},
	}

	// VAlidating email should leave form valid
	form = New(postdata)
	form.IsEmail("email")
	if !form.Valid() {
		t.Error("email is valid and form should still be valid")
	}

	form.IsEmail("notemail")
	if form.Valid() {
		t.Error("notemail is not valid and form should now be invalid")
	}

}

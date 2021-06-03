package forms

import (
	"encoding/json"
	"fmt"
	"github.com/bogdanrat/web-server/service/core/lib"
	"net/http"
	"net/url"
	"strings"
)

type Form struct {
	// typically used for query parameters and form values
	url.Values `json:"-"`
	Errors     errors `json:"errors"`
}

func New(data url.Values) *Form {
	return &Form{
		Values: data,
		Errors: make(map[string][]string),
	}
}

// Marshal returns the json encoding of form errors
func (f *Form) Marshal() ([]byte, error) {
	result, err := json.Marshal(f.Errors)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Checks if the form submitted by the user has a specified field
func (f *Form) Has(field string, req *http.Request) bool {
	if value := req.Form.Get(field); value == "" {
		//f.Errors.Add(field, "This field cannot be blank")
		return false
	}
	return true
}

// Checks for required fields
func (f *Form) Required(fields ...string) {
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" {
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// Returns true if there are no errors
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// Checks for string minimum length
func (f *Form) MinLength(field string, length int, r *http.Request) bool {
	value := r.Form.Get(field)
	if len(value) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// Checks for valid email address
func (f *Form) ValidEmail(field string) {
	if !lib.IsValidEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}

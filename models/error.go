package models

import (
	"net/http"
	"strings"
)

type JSONError struct {
	StatusCode  int    `json:"status_code,omitempty"`
	Description string `json:"description,omitempty"`
	Field       string `json:"field,omitempty"`
}

func NewBadRequestError(description string, field ...string) *JSONError {
	err := &JSONError{
		StatusCode:  http.StatusBadRequest,
		Description: description,
	}
	if len(field) > 0 {
		err.Field = strings.Join(field, ";")
	}
	return err
}

func NewNotFoundError(description string, field ...string) *JSONError {
	err := &JSONError{
		StatusCode:  http.StatusNotFound,
		Description: description,
	}
	if len(field) > 0 {
		err.Field = strings.Join(field, ";")
	}
	return err
}

func NewUnauthorizedError(description string, field ...string) *JSONError {
	err := &JSONError{
		StatusCode:  http.StatusUnauthorized,
		Description: description,
	}
	if len(field) > 0 {
		err.Field = strings.Join(field, ";")
	}
	return err
}

func NewInternalServerError(description string, field ...string) *JSONError {
	err := &JSONError{
		StatusCode:  http.StatusInternalServerError,
		Description: description,
	}
	if len(field) > 0 {
		err.Field = strings.Join(field, ";")
	}
	return err
}

func NewAlreadyReportedError(description string, field ...string) *JSONError {
	err := &JSONError{
		StatusCode:  http.StatusAlreadyReported,
		Description: description,
	}
	if len(field) > 0 {
		err.Field = strings.Join(field, ";")
	}
	return err
}

func NewUnprocessableEntityError(description string, field ...string) *JSONError {
	err := &JSONError{
		StatusCode:  http.StatusUnprocessableEntity,
		Description: description,
	}
	if len(field) > 0 {
		err.Field = strings.Join(field, ";")
	}
	return err
}

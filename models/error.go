package models

import "net/http"

type JSONError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func NewBadRequestError(message string) *JSONError {
	return &JSONError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

func NewNotFoundError(message string) *JSONError {
	return &JSONError{
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

func NewUnauthorizedError(message string) *JSONError {
	return &JSONError{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

func NewInternalServerError(message string) *JSONError {
	return &JSONError{
		StatusCode: http.StatusInternalServerError,
		Message:    message,
	}
}

func NewAlreadyReportedError(message string) *JSONError {
	return &JSONError{
		StatusCode: http.StatusAlreadyReported,
		Message:    message,
	}
}

func NewUnprocessableEntityError(message string) *JSONError {
	return &JSONError{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    message,
	}
}

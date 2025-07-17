package handlers

import (
	"fmt"
	"net/http"
)

type ApiError struct {
	StatusCode int `json:"statusCode"`
	Msg        any `json:"msg"`
}

func (e ApiError) Error() string {
	return fmt.Sprintf("api error: %d", e.StatusCode)
}

func NewApiError(statusCode int, err error) ApiError {
	return ApiError{
		StatusCode: statusCode,
		Msg:        err.Error(),
	}
}

func InvalidRequestData(errors map[string]string) ApiError {
	return ApiError{
		StatusCode: http.StatusUnprocessableEntity,
		Msg:        errors,
	}
}

func InvalidJSON() ApiError {
	return NewApiError(http.StatusBadRequest, fmt.Errorf("invalid JSON request data"))
}

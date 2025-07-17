package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPassword(password, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

func GenerateJWT(alumnoID int64, secretKey []byte) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	"sub": alumnoID,
	"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	
	return tokenString, nil
}


package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

// func GenerateJWT(alumnoID int64, secretKey string) (string, error) {

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"sub": alumnoID,
// 		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
// 	})

// 	tokenString, err := token.SignedString([]byte(secretKey))
// 	if err != nil {
// 		return "", fmt.Errorf("error firmando token: %w", err)
// 	}
// 	fmt.Printf("GenerateJWT: Token generado: %s\n", tokenString)
// 	return tokenString, nil
// }

type CustomClaims struct {
	Sub int64 `json:"sub"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(alumnoID int64, secretKey string) (string, error) {

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &CustomClaims{
		Sub: alumnoID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error firmando access token: %w", err)
	}
	fmt.Printf("GenerateAccessToken: Token generado: %s\n", tokenString)
	return tokenString, nil
}

func GenerateRefreshToken(alumnoID int64, secretKey string) (string, error) {
	expirationTime := time.Now().Add(30 * 24 * time.Hour)
	claims := &CustomClaims{
		Sub: alumnoID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error firmando refresh token: %w", err)
	}
	fmt.Printf("GenerateRefreshToken: Token generado: %s\n", tokenString)
	return tokenString, nil
}

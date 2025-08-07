package api

import (
	"bytes"
	"chedul-core/internals/domain"
	"chedul-core/internals/handlers"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/suite"
)

var MIGRATIONS_DIR string = "../../migrations"

type AlumnoHandlerSuite struct {
	suite.Suite
}

func (a *AlumnoHandlerSuite) SignUpRequest(address string,
	payload domain.SignUpRequest) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", address, bytes.NewBuffer(body))

	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func createTestToken(secretKey string, alumnoID int64) string {
	claims := handlers.CustomClaims{
		Sub: alumnoID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secretKey))

	return tokenString
}

func (s *AlumnoHandlerSuite) TestGetAlumnoByIdFail() {
	testApp, err := CreateTestApp()
	s.Require().NoError(err)

	defer testApp.Cleanup()

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		testApp.Address+"/api/v1/alumnos/1", nil)

	s.NoError(err)
	token := createTestToken(testApp.server.JwtSecret.Expose(), 1)
	req.Header.Set("Authorization", "Bearer "+token)

	response, err := client.Do(req)
	s.NoError(err)
	defer response.Body.Close()

	s.Assert().Equal(http.StatusInternalServerError, response.StatusCode)
}

func (s *AlumnoHandlerSuite) TestSignUp() {
	testApp, err := CreateTestApp()
	s.Require().NoError(err)
	defer testApp.Cleanup()
	response, err := s.SignUpRequest(testApp.Address+"/api/v1/signup",
		domain.SignUpRequest{
			Nombre:   "Lautaro",
			Email:    "lautaroacosta@gmail.com",
			Carrera:  "ISI",
			Password: "Chedul123",
		},
	)
	s.NoError(err)
	defer response.Body.Close()
	s.Assert().Equal(http.StatusCreated, response.StatusCode)
}

func TestAlumnoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	suite.Run(t, new(AlumnoHandlerSuite))
}

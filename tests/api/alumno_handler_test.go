package api

import (
	"bytes"
	"chedul-core/internals/domain"
	"chedul-core/internals/handlers"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/uptrace/bun"
)

var MIGRATIONS_DIR string = "../../migrations"

type AlumnoHandlerSuite struct {
	suite.Suite
	db        *bun.DB
	container testcontainers.Container
	ctx       context.Context
}

func (s *AlumnoHandlerSuite) SetupSuite() {
	testingDB, err := SetupPgContainer(MIGRATIONS_DIR)
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.container = testingDB.container
	s.db = testingDB.db
}

func (s *AlumnoHandlerSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		s.container.Terminate(s.ctx)
	}
}

func (s *AlumnoHandlerSuite) SetupTest() {
	s.cleanDatabase()
	s.seedDatabase()
}

func (s *AlumnoHandlerSuite) cleanDatabase() {
	if s.db == nil {
		s.T().Fatal("Database connection is nil")
		return
	}

	goose.SetLogger(goose.NopLogger())
	goose.Reset(s.db.DB, MIGRATIONS_DIR)
	goose.Up(s.db.DB, MIGRATIONS_DIR)
}

func (s *AlumnoHandlerSuite) seedDatabase() {
	if s.db == nil {
		s.T().Fatal("Database connection is nil")
		return
	}
	_, err :=
		s.db.Exec("INSERT INTO carrera(nombre) VALUES ('ISI'), ('Sistemas')")

	s.Require().NoError(err)
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

func (s *AlumnoHandlerSuite) TestGetAlumnoById() {
	configPath := os.Getenv("CONFIG_DIR")
	if configPath == "" {
		s.T().FailNow()
	}

	testApp, err := SpawnApp("../"+configPath, WithDB("../../migrations"))
	defer testApp.Cleanup()

	s.NoError(err)

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		testApp.Address+"/api/v1/alumnos/1", nil)

	s.NoError(err)
	req.Header.Set("Authorization",
		"Bearer "+createTestToken(testApp.server.JwtSecret.Expose(), 1))

	response, err := client.Do(req)
	s.NoError(err)
	defer response.Body.Close()

	s.Assert().Equal(http.StatusOK, response.StatusCode)
}

func (s *AlumnoHandlerSuite) TestSignUp() {
	configPath := os.Getenv("CONFIG_DIR")
	if configPath == "" {
		s.T().FailNow()
	}

	testApp, err := SpawnApp("../"+configPath, WithDB("../../migrations"))
	defer testApp.Cleanup()
	s.NoError(err)

	body, err := json.Marshal(
		domain.SignUpRequest{
			Nombre:   "Lautaro",
			Email:    "lautaroacosta@gmail.com",
			Carrera:  "ISI",
			Password: "Chedul123",
		},
	)
	s.NoError(err)

	client := &http.Client{}
	req, err := http.NewRequest("POST",
		testApp.Address+"/api/v1/signup", bytes.NewBuffer(body))

	s.NoError(err)
	req.Header.Set("Content-Type", "application/json")

	response, err := client.Do(req)
	s.NoError(err)
	defer response.Body.Close()

	s.Assert().Equal(http.StatusOK, response.StatusCode)
}

func TestAlumnoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	if os.Getenv("CONFIG_DIR") == "" {
		t.Skip("env var 'CONFIG_DIR' no inicializada")
	}

	suite.Run(t, new(AlumnoHandlerSuite))
}

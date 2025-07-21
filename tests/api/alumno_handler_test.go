package api

import (
	"chedul-core/internals/domain"
	"chedul-core/internals/repositories"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/uptrace/bun"
)

type AlumnoTestSuite struct {
	suite.Suite
	db        *bun.DB
	container testcontainers.Container
	ctx       context.Context
}

func (s *AlumnoTestSuite) SetupSuite() {
	// TODO: Resolver el tema de que se hardcodee
	testingDB, err := SetupPgContainer("../../migrations")
	s.Require().NoError(err)

	s.ctx = context.Background()
	s.container = testingDB.container
	s.db = testingDB.db
}

func (s *AlumnoTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		s.container.Terminate(s.ctx)
	}
}

func (s *AlumnoTestSuite) SetupTest() {
	s.cleanDatabase()
	s.seedDatabase()
}

func (s *AlumnoTestSuite) cleanDatabase() {
	if s.db == nil {
		s.T().Fatal("Database connection is nil")
		return
	}

	s.db.Exec("DELETE FROM alumno")
	s.db.Exec("DELETE FROM carrera")
	s.db.Exec("ALTER SEQUENCE alumno_id_seq RESTART WITH 1")
	s.db.Exec("ALTER SEQUENCE carrera_id_seq RESTART WITH 1")
}

func (s *AlumnoTestSuite) seedDatabase() {
	if s.db == nil {
		s.T().Fatal("Database connection is nil")
		return
	}
	_, err := s.db.Exec("INSERT INTO carrera(nombre) VALUES ('ISI'), ('Sistemas')")
	s.Require().NoError(err)
}

func (s *AlumnoTestSuite) createTestAlumno(name, email string) domain.Alumno {
	username, err := domain.NewUserName(name)
	s.Require().NoError(err)

	emailObj, err := domain.NewEmail(email)
	s.Require().NoError(err)

	return domain.Alumno{
		Nombre:   username,
		Email:    emailObj,
		Carrera:  1,
		Password: "password123",
	}
}

func (s *AlumnoTestSuite) TestGetAlumnoById() {
	repo := repositories.NewAlumnoRepository(s.db)

	alumno := s.createTestAlumno("Lautaro Acosta", "lautaro@acosta.com")
	err := repo.Create(s.ctx, &alumno)
	s.Require().NoError(err)

	result, err := repo.GetByID(s.ctx, 1)
	s.Assert().NoError(err)
	s.assertAlumnoEquals(alumno, *result)
}

func (s *AlumnoTestSuite) TestGetAlumnoByEmail() {
	repo := repositories.NewAlumnoRepository(s.db)
	alumno := s.createTestAlumno("Lautaro Acosta", "lautaro@acosta.com")
	err := repo.Create(s.ctx, &alumno)
	s.Require().NoError(err)

	result, err := repo.GetByEmail(s.ctx, "lautaro@acosta.com")
	s.Assert().NoError(err)
	s.Assert().Equal(alumno.Email, result.Email)
}

func (s *AlumnoTestSuite) assertAlumnoEquals(expected, actual domain.Alumno) {
	s.Assert().Equal(expected.Password, actual.Password)
	s.Assert().Equal(expected.Carrera, actual.Carrera)
	s.Assert().Equal(expected.Nombre, actual.Nombre)
	s.Assert().Equal(expected.Email, actual.Email)
}

func TestAlumnoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	suite.Run(t, new(AlumnoTestSuite))
}

package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AlumnoHandlerSuite struct {
	suite.Suite
	app *TestApp
}

func (s *AlumnoHandlerSuite) SetupSuite() {
	app, err := CreateTestApp()
	s.Require().NoError(err)
	s.app = app
}

func (s *AlumnoHandlerSuite) TearDownSuite() {
	s.app.Cleanup()
}

func (s *AlumnoHandlerSuite) TestRutasProtegidasExigenToken() {
	c := NewClient(s.T(), s.app)

	for _, path := range []string{"/alumnos/me", "/alumnos/me/progreso", "/condicion_alumno", "/aportes"} {
		resp := c.JSON("GET", path, nil)
		s.Equal(http.StatusUnauthorized, resp.Status, path)
	}

	c.AccessToken = "token-trucho"
	resp := c.JSON("GET", "/alumnos/me", nil)
	s.Equal(http.StatusUnauthorized, resp.Status)
	s.Contains(string(resp.Body), "Access Token inválido o expirado")
}

func (s *AlumnoHandlerSuite) TestSignUpLoginYPerfil() {
	c := NewClient(s.T(), s.app)
	id := c.Registrar("Lautaro Acosta Quintana", "Lautaro@Gmail.com")

	resp := c.JSON("GET", "/alumnos/me", nil)
	s.Require().Equal(http.StatusOK, resp.Status)

	var me struct {
		ID      int64  `json:"id"`
		Nombre  string `json:"nombre"`
		Email   string `json:"email"`
		Carrera int64  `json:"carrera"`
	}
	resp.JSON(s.T(), &me)
	s.Equal(id, me.ID)
	s.Equal("lautaro@gmail.com", me.Email, "el correo se guarda en minusculas")
	s.NotContains(string(resp.Body), "password")

	resp = c.JSON("PUT", "/alumnos/me", map[string]any{"nombre": "Lauti", "carrera_id": 1})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	resp.JSON(s.T(), &me)
	s.Equal("Lauti", me.Nombre)
}

func (s *AlumnoHandlerSuite) TestSignUpConCorreoRepetido() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Ana", "ana@chedul.com")

	resp := c.JSON("POST", "/signup", map[string]any{
		"nombre": "Ana 2", "email": "ANA@chedul.com", "carrera_id": 1, "password": "Chedul-123",
	})
	s.Equal(http.StatusUnprocessableEntity, resp.Status)
}

func (s *AlumnoHandlerSuite) TestSignUpValidaCarrera() {
	c := NewClient(s.T(), s.app)
	resp := c.JSON("POST", "/signup", map[string]any{
		"nombre": "Juan", "email": "juan@chedul.com", "carrera_id": 999, "password": "Chedul-123",
	})
	s.Equal(http.StatusUnprocessableEntity, resp.Status)
}

func (s *AlumnoHandlerSuite) TestLoginInvalido() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Sofi", "sofi@chedul.com")

	resp := c.JSON("POST", "/login", map[string]any{"email": "sofi@chedul.com", "password": "otra-Clave1"})
	s.Equal(http.StatusUnauthorized, resp.Status)
}

func (s *AlumnoHandlerSuite) TestRefreshYLogout() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Mati", "mati@chedul.com")

	c.AccessToken = ""
	resp := c.JSON("POST", "/refresh-token", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))

	var data struct {
		AccessToken string `json:"accessToken"`
	}
	resp.JSON(s.T(), &data)
	s.NotEmpty(data.AccessToken)

	resp = c.JSON("POST", "/logout", nil)
	s.Equal(http.StatusNoContent, resp.Status)

	resp = c.JSON("POST", "/refresh-token", nil)
	s.Equal(http.StatusUnauthorized, resp.Status)
}

func (s *AlumnoHandlerSuite) TestCondicionYProgreso() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Caro", "caro@chedul.com")

	am1 := c.MateriaID("isi-am1")
	ayga := c.MateriaID("isi-ayga")
	am2 := c.MateriaID("isi-am2")
	an := c.MateriaID("isi-an")

	aprobada := c.CondicionID("Aprobada")
	regularizada := c.CondicionID("Regularizada")

	progreso := func() progresoTest {
		resp := c.JSON("GET", "/alumnos/me/progreso", nil)
		s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
		var p progresoTest
		resp.JSON(s.T(), &p)
		return p
	}

	contiene := func(lista []materiaProgresoTest, id int64) bool {
		for _, m := range lista {
			if m.ID == id {
				return true
			}
		}
		return false
	}

	p := progreso()
	s.True(contiene(p.NoDisponibles, am2), "AM2 requiere AM1 y AyGA regularizadas")

	nota := 8
	resp := c.JSON("PUT", ruta("/condicion_alumno/%d", am1), map[string]any{"condicion_id": aprobada, "nota": nota})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	resp = c.JSON("PUT", ruta("/condicion_alumno/%d", ayga), map[string]any{"condicion_id": regularizada})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))

	p = progreso()
	s.True(contiene(p.Aprobadas, am1))
	s.True(contiene(p.Regularizadas, ayga))
	s.True(contiene(p.Disponibles, am2), "con AM1 aprobada y AyGA regular ya puede cursar AM2")
	s.True(contiene(p.NoDisponibles, an), "Analisis Numerico requiere AyGA aprobada")
	s.Require().NotNil(p.Promedio)
	s.Equal(8.0, *p.Promedio)

	resp = c.JSON("GET", "/condicion_alumno", nil)
	s.Require().Equal(http.StatusOK, resp.Status)
	var condiciones []map[string]any
	resp.JSON(s.T(), &condiciones)
	s.Len(condiciones, 2)

	resp = c.JSON("DELETE", ruta("/condicion_alumno/%d", ayga), nil)
	s.Equal(http.StatusNoContent, resp.Status)

	resp = c.JSON("PUT", ruta("/condicion_alumno/%d", am1), map[string]any{"condicion_id": aprobada, "nota": 11})
	s.Equal(http.StatusUnprocessableEntity, resp.Status)

	resp = c.JSON("PUT", "/condicion_alumno/99999", map[string]any{"condicion_id": aprobada})
	s.Equal(http.StatusNotFound, resp.Status)
}

type materiaProgresoTest struct {
	ID      int64  `json:"id"`
	Mensaje string `json:"mensaje"`
}

type progresoTest struct {
	Aprobadas     []materiaProgresoTest `json:"materias_aprobadas"`
	Regularizadas []materiaProgresoTest `json:"materias_regularizadas"`
	Disponibles   []materiaProgresoTest `json:"materias_pendientes_disponibles"`
	NoDisponibles []materiaProgresoTest `json:"materias_pendientes_no_disponibles"`
	Promedio      *float64              `json:"promedio"`
}

func TestAlumnoSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	suite.Run(t, new(AlumnoHandlerSuite))
}

func (s *AlumnoHandlerSuite) TestLoginGoogleRechazaTokenTrucho() {
	c := NewClient(s.T(), s.app)

	resp := c.JSON("GET", "/auth/google", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	s.Contains(string(resp.Body), ".apps.googleusercontent.com")

	resp = c.JSON("POST", "/auth/google", map[string]string{"credential": "token-trucho"})
	s.Equal(http.StatusUnauthorized, resp.Status, string(resp.Body))
	s.Contains(string(resp.Body), "No se pudo verificar tu cuenta de Google")

	resp = c.JSON("POST", "/auth/google", map[string]string{})
	s.Equal(http.StatusBadRequest, resp.Status, string(resp.Body))
}

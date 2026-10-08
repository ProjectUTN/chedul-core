package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type eventoTest struct {
	ID          int64   `json:"id"`
	Titulo      string  `json:"titulo"`
	Tipo        string  `json:"tipo"`
	Fecha       string  `json:"fecha"`
	Hora        *string `json:"hora"`
	Descripcion string  `json:"descripcion"`
	Materia     *struct {
		ID     int64  `json:"id"`
		Nombre string `json:"nombre"`
	} `json:"materia"`
}

type claseTest struct {
	ID         int64  `json:"id"`
	Titulo     string `json:"titulo"`
	Dia        int    `json:"dia"`
	HoraInicio string `json:"hora_inicio"`
	HoraFin    string `json:"hora_fin"`
	Aula       string `json:"aula"`
	Materia    *struct {
		ID int64 `json:"id"`
	} `json:"materia"`
	ComisionID   *int64  `json:"comision_id"`
	Cuatrimestre *string `json:"cuatrimestre"`
}

type CalendarioHandlerSuite struct {
	suite.Suite
	app *TestApp
}

func (s *CalendarioHandlerSuite) SetupSuite() {
	app, err := CreateTestApp()
	s.Require().NoError(err)
	s.app = app
}

func (s *CalendarioHandlerSuite) TearDownSuite() {
	s.app.Cleanup()
}

func (s *CalendarioHandlerSuite) eventos(c *Client, desde, hasta string) []eventoTest {
	resp := c.JSON("GET", "/eventos?desde="+desde+"&hasta="+hasta, nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var eventos []eventoTest
	resp.JSON(s.T(), &eventos)
	return eventos
}

func (s *CalendarioHandlerSuite) TestEventos() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Facu", "facu.academico@chedul.com")
	am1 := c.MateriaID("isi-am1")

	resp := c.JSON("POST", "/eventos", map[string]any{
		"titulo": "Primer parcial", "tipo": "parcial", "fecha": "2026-05-12",
		"hora": "18:30", "materia_id": am1, "descripcion": "Límites y derivadas",
	})
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	var parcial eventoTest
	resp.JSON(s.T(), &parcial)
	s.Equal("2026-05-12", parcial.Fecha)
	s.Require().NotNil(parcial.Hora)
	s.Equal("18:30", *parcial.Hora)
	s.Require().NotNil(parcial.Materia)
	s.Equal(am1, parcial.Materia.ID)

	// Sin hora, sin materia y sin tipo: queda como "otro"
	resp = c.JSON("POST", "/eventos", map[string]any{"titulo": "Pagar matrícula", "fecha": "2026-05-01", "hora": ""})
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	var recordatorio eventoTest
	resp.JSON(s.T(), &recordatorio)
	s.Equal("otro", recordatorio.Tipo)
	s.Nil(recordatorio.Hora)
	s.Nil(recordatorio.Materia)

	resp = c.JSON("POST", "/eventos", map[string]any{"titulo": "Final", "tipo": "final", "fecha": "2026-07-20"})
	s.Require().Equal(http.StatusCreated, resp.Status)

	mayo := s.eventos(c, "2026-05-01", "2026-05-31")
	s.Require().Len(mayo, 2)
	s.Equal("Pagar matrícula", mayo[0].Titulo, "ordenados por fecha")

	resp = c.JSON("PUT", ruta("/eventos/%d", parcial.ID), map[string]any{
		"titulo": "Primer parcial (recuperatorio)", "tipo": "parcial", "fecha": "2026-06-02",
	})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var editado eventoTest
	resp.JSON(s.T(), &editado)
	s.Equal("2026-06-02", editado.Fecha)
	s.Nil(editado.Hora)
	s.Nil(editado.Materia)
	s.Len(s.eventos(c, "2026-05-01", "2026-05-31"), 1)

	// Otro alumno no ve ni toca los eventos ajenos
	otro := NewClient(s.T(), s.app)
	otro.Registrar("Beto", "beto.calendario@chedul.com")
	s.Empty(s.eventos(otro, "2026-01-01", "2026-12-31"))
	resp = otro.JSON("PUT", ruta("/eventos/%d", parcial.ID), map[string]any{"titulo": "x", "fecha": "2026-01-01"})
	s.Equal(http.StatusNotFound, resp.Status)
	resp = otro.JSON("DELETE", ruta("/eventos/%d", parcial.ID), nil)
	s.Equal(http.StatusNotFound, resp.Status)

	resp = c.JSON("DELETE", ruta("/eventos/%d", parcial.ID), nil)
	s.Equal(http.StatusNoContent, resp.Status)
	s.Empty(s.eventos(c, "2026-06-01", "2026-06-30"))
}

func (s *CalendarioHandlerSuite) TestValidacionesEventos() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Caro", "caro.calendario@chedul.com")

	casos := map[string]map[string]any{
		"sin titulo":          {"fecha": "2026-05-01"},
		"fecha invalida":      {"titulo": "x", "fecha": "01/05/2026"},
		"hora invalida":       {"titulo": "x", "fecha": "2026-05-01", "hora": "25:00"},
		"tipo invalido":       {"titulo": "x", "fecha": "2026-05-01", "tipo": "fiesta"},
		"materia inexistente": {"titulo": "x", "fecha": "2026-05-01", "materia_id": 99999},
	}
	for nombre, payload := range casos {
		resp := c.JSON("POST", "/eventos", payload)
		s.Equal(http.StatusUnprocessableEntity, resp.Status, nombre+": "+string(resp.Body))
	}

	for _, query := range []string{"", "?desde=2026-05-01", "?desde=2026-05-10&hasta=2026-05-01", "?desde=2020-01-01&hasta=2026-01-01"} {
		resp := c.JSON("GET", "/eventos"+query, nil)
		s.Equal(http.StatusUnprocessableEntity, resp.Status, query)
	}
}

func (s *CalendarioHandlerSuite) TestHorarioSemanal() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Dani", "dani.calendario@chedul.com")
	aed := c.MateriaID("isi-aed")

	nueva := func(payload map[string]any) claseTest {
		resp := c.JSON("POST", "/clases", payload)
		s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
		var clase claseTest
		resp.JSON(s.T(), &clase)
		return clase
	}

	jueves := nueva(map[string]any{"titulo": "AED práctica", "dia": 4, "hora_inicio": "08:30", "hora_fin": "12:45", "materia_id": aed, "aula": "Lab 3"})
	lunes := nueva(map[string]any{"titulo": "AED teoría", "dia": 1, "hora_inicio": "19:00", "hora_fin": "23:00", "materia_id": aed})
	s.Equal("08:30", jueves.HoraInicio)
	s.Equal("Lab 3", jueves.Aula)
	s.Nil(jueves.ComisionID)
	s.Nil(jueves.Cuatrimestre)

	// Clase cargada desde una comision: guarda de cual viene
	resp0 := c.JSON("GET", ruta("/materias/%d/comisiones", aed), nil)
	var comisiones []struct {
		ID int64 `json:"id"`
	}
	resp0.JSON(s.T(), &comisiones)
	s.Require().NotEmpty(comisiones)
	desdeComision := nueva(map[string]any{"titulo": "AED", "dia": 2, "hora_inicio": "10:55", "hora_fin": "12:25", "materia_id": aed, "comision_id": comisiones[0].ID})
	s.Require().NotNil(desdeComision.ComisionID)
	s.Equal(comisiones[0].ID, *desdeComision.ComisionID)
	// y el cuatrimestre de esa comision, para mostrarla solo cuando se cursa
	s.Require().NotNil(desdeComision.Cuatrimestre)
	s.Equal("1C", *desdeComision.Cuatrimestre)
	resp0 = c.JSON("POST", "/clases", map[string]any{"titulo": "x", "dia": 2, "hora_inicio": "10:55", "hora_fin": "12:25", "materia_id": c.MateriaID("isi-am1"), "comision_id": comisiones[0].ID})
	s.Equal(http.StatusUnprocessableEntity, resp0.Status, string(resp0.Body))
	resp0 = c.JSON("DELETE", ruta("/clases/%d", desdeComision.ID), nil)
	s.Require().Equal(http.StatusNoContent, resp0.Status, string(resp0.Body))

	resp := c.JSON("GET", "/clases", nil)
	s.Require().Equal(http.StatusOK, resp.Status)
	var clases []claseTest
	resp.JSON(s.T(), &clases)
	s.Require().Len(clases, 2)
	s.Equal(lunes.ID, clases[0].ID, "ordenadas por día")

	casos := map[string]map[string]any{
		"dia fuera de rango":  {"titulo": "x", "dia": 8, "hora_inicio": "08:00", "hora_fin": "09:00"},
		"termina antes":       {"titulo": "x", "dia": 2, "hora_inicio": "10:00", "hora_fin": "09:00"},
		"hora sin formato":    {"titulo": "x", "dia": 2, "hora_inicio": "8", "hora_fin": "09:00"},
		"materia inexistente": {"titulo": "x", "dia": 2, "hora_inicio": "08:00", "hora_fin": "09:00", "materia_id": 99999},
	}
	for nombre, payload := range casos {
		resp := c.JSON("POST", "/clases", payload)
		s.Equal(http.StatusUnprocessableEntity, resp.Status, nombre+": "+string(resp.Body))
	}

	resp = c.JSON("PUT", ruta("/clases/%d", jueves.ID), map[string]any{"titulo": "AED práctica", "dia": 5, "hora_inicio": "08:30", "hora_fin": "12:45"})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var editada claseTest
	resp.JSON(s.T(), &editada)
	s.Equal(5, editada.Dia)
	s.Nil(editada.Materia)

	otro := NewClient(s.T(), s.app)
	otro.Registrar("Eze", "eze.calendario@chedul.com")
	resp = otro.JSON("DELETE", ruta("/clases/%d", jueves.ID), nil)
	s.Equal(http.StatusNotFound, resp.Status)

	resp = c.JSON("DELETE", ruta("/clases/%d", jueves.ID), nil)
	s.Equal(http.StatusNoContent, resp.Status)
}

func (s *CalendarioHandlerSuite) TestCalendarioAcademico() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Facu", "facu.academico@chedul.com")

	resp := c.JSON("GET", "/calendario-academico?desde=2026-07-15&hasta=2026-08-12", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var fechas []struct {
		Titulo string `json:"titulo"`
		Tipo   string `json:"tipo"`
		Desde  string `json:"desde"`
		Hasta  string `json:"hasta"`
	}
	resp.JSON(s.T(), &fechas)
	s.Require().Len(fechas, 4)
	// Empezo antes del rango pero sigue adentro, asi que viene
	s.Equal("Recuperatorios y Fin Primer Cuatrimestre", fechas[0].Titulo)
	s.Equal("2026-07-13", fechas[0].Desde)
	s.Equal("2026-07-18", fechas[0].Hasta)
	s.Equal("receso", fechas[1].Tipo)
	s.Equal("Exámenes Finales 3° Llamado", fechas[2].Titulo)
	s.Equal("examen", fechas[2].Tipo)
	s.Equal("Inicio Segundo Cuatrimestre", fechas[3].Titulo)

	resp = c.JSON("GET", "/calendario-academico?desde=2026-08-12&hasta=2026-08-01", nil)
	s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))
}

func (s *CalendarioHandlerSuite) TestRequiereSesion() {
	c := NewClient(s.T(), s.app)
	s.Equal(http.StatusUnauthorized, c.JSON("GET", "/clases", nil).Status)
	s.Equal(http.StatusUnauthorized, c.JSON("GET", "/eventos?desde=2026-01-01&hasta=2026-01-31", nil).Status)
	s.Equal(http.StatusUnauthorized, c.JSON("GET", "/calendario-academico?desde=2026-01-01&hasta=2026-01-31", nil).Status)
}

func TestCalendarioSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	suite.Run(t, new(CalendarioHandlerSuite))
}

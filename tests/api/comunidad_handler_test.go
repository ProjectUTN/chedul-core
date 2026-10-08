package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type comunidadTest struct {
	ID         int64  `json:"id"`
	Nombre     string `json:"nombre"`
	Plataforma string `json:"plataforma"`
	Link       string `json:"link"`
	EsMia      bool   `json:"es_mia"`
	Reportada  bool   `json:"reportada"`
	Materia    *struct {
		ID int64 `json:"id"`
	} `json:"materia"`
}

type ComunidadHandlerSuite struct {
	suite.Suite
	app *TestApp
}

func (s *ComunidadHandlerSuite) SetupSuite() {
	app, err := CreateTestApp()
	s.Require().NoError(err)
	s.app = app
}

func (s *ComunidadHandlerSuite) TearDownSuite() {
	s.app.Cleanup()
}

func (s *ComunidadHandlerSuite) listar(c *Client) []comunidadTest {
	resp := c.JSON("GET", "/comunidades", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var lista []comunidadTest
	resp.JSON(s.T(), &lista)
	return lista
}

func (s *ComunidadHandlerSuite) TestCargarBorrarYReportar() {
	ana := NewClient(s.T(), s.app)
	ana.Registrar("Ana Comunidad", "ana.comunidad@chedul.com")
	aed := ana.MateriaID("isi-aed")

	resp := ana.JSON("POST", "/comunidades", map[string]any{"nombre": "AED 2026", "link": "https://chat.whatsapp.com/aed2026", "materia_id": aed})
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	var grupo comunidadTest
	resp.JSON(s.T(), &grupo)
	s.Equal("whatsapp", grupo.Plataforma)
	s.True(grupo.EsMia)
	s.Require().NotNil(grupo.Materia)

	for _, payload := range []map[string]any{
		{"nombre": "Repetida", "link": "https://chat.whatsapp.com/aed2026"},
		{"nombre": "Sin https", "link": "http://discord.gg/x"},
		{"nombre": "", "link": "https://discord.gg/x"},
		{"nombre": "Materia rara", "link": "https://discord.gg/y", "materia_id": 999999},
	} {
		resp = ana.JSON("POST", "/comunidades", payload)
		s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))
	}

	// Otros la ven pero no la pueden borrar
	otros := make([]*Client, 3)
	for i, email := range []string{"beto.comunidad@chedul.com", "caro.comunidad@chedul.com", "dani.comunidad@chedul.com"} {
		otros[i] = NewClient(s.T(), s.app)
		otros[i].Registrar("Otro", email)
	}
	lista := s.listar(otros[0])
	s.Require().Len(lista, 1)
	s.False(lista[0].EsMia)
	s.Equal(http.StatusNotFound, otros[0].JSON("DELETE", ruta("/comunidades/%d", grupo.ID), nil).Status)

	// Reportar dos veces cuenta una; con tres alumnos distintos se oculta
	s.Equal(http.StatusNoContent, otros[0].JSON("POST", ruta("/comunidades/%d/reportar", grupo.ID), nil).Status)
	s.Equal(http.StatusNoContent, otros[0].JSON("POST", ruta("/comunidades/%d/reportar", grupo.ID), nil).Status)
	s.True(s.listar(otros[0])[0].Reportada)
	s.Equal(http.StatusNoContent, otros[1].JSON("POST", ruta("/comunidades/%d/reportar", grupo.ID), nil).Status)
	s.Len(s.listar(ana), 1)
	s.Equal(http.StatusNoContent, otros[2].JSON("POST", ruta("/comunidades/%d/reportar", grupo.ID), nil).Status)
	s.Empty(s.listar(ana))
	s.Equal(http.StatusNotFound, otros[2].JSON("POST", "/comunidades/999999/reportar", nil).Status)

	// La autora la puede borrar igual
	resp = ana.JSON("POST", "/comunidades", map[string]any{"nombre": "ISI FRRe", "link": "https://discord.gg/isifrre"})
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	resp.JSON(s.T(), &grupo)
	s.Equal("discord", grupo.Plataforma)
	s.Equal(http.StatusNoContent, ana.JSON("DELETE", ruta("/comunidades/%d", grupo.ID), nil).Status)
	s.Empty(s.listar(ana))
}

func TestComunidadHandler(t *testing.T) {
	suite.Run(t, new(ComunidadHandlerSuite))
}

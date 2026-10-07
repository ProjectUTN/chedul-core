package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type materiaTest struct {
	ID           int64  `json:"id"`
	Codigo       string `json:"codigo"`
	Tipo         string `json:"tipo"`
	Programa     string `json:"programa"`
	Correlativas []struct {
		MateriaID int64  `json:"materia_id"`
		Tipo      string `json:"tipo"`
	} `json:"correlativas"`
}

type comisionTest struct {
	Codigo       string `json:"codigo"`
	Cuatrimestre string `json:"cuatrimestre"`
	Horarios     []struct {
		Dia        int    `json:"dia"`
		HoraInicio string `json:"hora_inicio"`
		HoraFin    string `json:"hora_fin"`
	} `json:"horarios"`
}

type MateriaHandlerSuite struct {
	suite.Suite
	app *TestApp
}

func (s *MateriaHandlerSuite) SetupSuite() {
	app, err := CreateTestApp()
	s.Require().NoError(err)
	s.app = app
}

func (s *MateriaHandlerSuite) TearDownSuite() {
	s.app.Cleanup()
}

func (s *MateriaHandlerSuite) TestPlanCompleto() {
	c := NewClient(s.T(), s.app)

	resp := c.JSON("GET", "/materias", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var materias []materiaTest
	resp.JSON(s.T(), &materias)

	porCodigo := map[string]materiaTest{}
	electivas := 0
	for _, m := range materias {
		porCodigo[m.Codigo] = m
		if m.Tipo == "Electiva" {
			electivas++
		}
	}
	s.Len(materias, 51)
	s.Equal(14, electivas)
	s.Contains(porCodigo, "isi-ps")

	// Las electivas traen sus correlativas
	s.Len(porCodigo["isi-devops"].Correlativas, 4)
	// y las materias el link al programa analitico
	s.Contains(porCodigo["isi-aed"].Programa, "frre.utn.edu.ar")
}

func (s *MateriaHandlerSuite) TestComisiones() {
	c := NewClient(s.T(), s.app)
	aed := c.MateriaID("isi-aed")

	resp := c.JSON("GET", fmt.Sprintf("/materias/%d/comisiones", aed), nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var comisiones []comisionTest
	resp.JSON(s.T(), &comisiones)

	s.Require().Len(comisiones, 5)
	k11 := comisiones[0]
	s.Equal("K1.1", k11.Codigo)
	s.Equal("Anual", k11.Cuatrimestre)
	s.Require().Len(k11.Horarios, 2)
	s.Equal(2, k11.Horarios[0].Dia) // martes
	s.Equal("10:55", k11.Horarios[0].HoraInicio)
	s.Equal("12:25", k11.Horarios[0].HoraFin)
	s.Equal(5, k11.Horarios[1].Dia) // viernes

	// Materia sin comisiones cargadas: lista vacia
	resp = c.JSON("GET", fmt.Sprintf("/materias/%d/comisiones", c.MateriaID("isi-devops")), nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	s.JSONEq("[]", string(resp.Body))

	resp = c.JSON("GET", "/materias/999999/comisiones", nil)
	s.Equal(http.StatusNotFound, resp.Status)
}

func TestMateriaHandler(t *testing.T) {
	suite.Run(t, new(MateriaHandlerSuite))
}

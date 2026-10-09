package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type sesionTest struct {
	ID      int64  `json:"id"`
	Modo    string `json:"modo"`
	Minutos int    `json:"minutos"`
	Inicio  string `json:"inicio"`
	Fin     string `json:"fin"`
	Materia *struct {
		ID     int64  `json:"id"`
		Nombre string `json:"nombre"`
	} `json:"materia"`
}

type rankingTest struct {
	Participo     bool `json:"participo"`
	Participantes int  `json:"participantes"`
	Puestos       []struct {
		Posicion int    `json:"posicion"`
		Nombre   string `json:"nombre"`
		Minutos  int    `json:"minutos"`
		SoyYo    bool   `json:"soy_yo"`
	} `json:"puestos"`
}

type EstudioHandlerSuite struct {
	suite.Suite
	app *TestApp
}

func (s *EstudioHandlerSuite) SetupSuite() {
	app, err := CreateTestApp()
	s.Require().NoError(err)
	s.app = app
}

func (s *EstudioHandlerSuite) TearDownSuite() {
	s.app.Cleanup()
}

func (s *EstudioHandlerSuite) TestSesionesYResumen() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Gabi Sosa", "gabi.estudio@chedul.com")
	aed := c.MateriaID("isi-aed")

	nueva := func(payload map[string]any) sesionTest {
		resp := c.JSON("POST", "/estudio/sesiones", payload)
		s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
		var sesion sesionTest
		resp.JSON(s.T(), &sesion)
		return sesion
	}

	pomodoro := nueva(map[string]any{"modo": "pomodoro", "minutos": 25, "materia_id": aed})
	s.Equal(25, pomodoro.Minutos)
	s.Require().NotNil(pomodoro.Materia)
	s.Equal(aed, pomodoro.Materia.ID)
	s.NotEmpty(pomodoro.Fin)
	libre := nueva(map[string]any{"modo": "libre", "minutos": 50})
	s.Nil(libre.Materia)

	for _, payload := range []map[string]any{
		{"modo": "siesta", "minutos": 25},
		{"modo": "libre", "minutos": 0},
		{"modo": "libre", "minutos": 721},
		{"modo": "libre", "minutos": 10, "materia_id": 999999},
	} {
		resp := c.JSON("POST", "/estudio/sesiones", payload)
		s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))
	}

	resp := c.JSON("GET", "/estudio/resumen", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var resumen struct {
		HoyMinutos    int `json:"hoy_minutos"`
		SemanaMinutos int `json:"semana_minutos"`
		MesMinutos    int `json:"mes_minutos"`
		MetaDiaria    int `json:"meta_diaria"`
		RachaDias     int `json:"racha_dias"`
		PorDia        []struct {
			Fecha   string `json:"fecha"`
			Minutos int    `json:"minutos"`
		} `json:"por_dia"`
		PorMateria []struct {
			MateriaID *int64 `json:"materia_id"`
			Minutos   int    `json:"minutos"`
		} `json:"por_materia"`
	}
	resp.JSON(s.T(), &resumen)
	s.Equal(75, resumen.HoyMinutos)
	s.Equal(75, resumen.SemanaMinutos)
	s.Equal(1, resumen.RachaDias)
	s.Len(resumen.PorDia, 182)
	s.Equal(75, resumen.PorDia[181].Minutos)
	s.Equal(75, resumen.MesMinutos)
	s.Equal(120, resumen.MetaDiaria)
	s.Require().Len(resumen.PorMateria, 2)
	s.Nil(resumen.PorMateria[0].MateriaID)
	s.Equal(50, resumen.PorMateria[0].Minutos)

	// Otro alumno no ve ni borra las sesiones ajenas
	otro := NewClient(s.T(), s.app)
	otro.Registrar("Hugo", "hugo.estudio@chedul.com")
	s.Equal(http.StatusNotFound, otro.JSON("DELETE", ruta("/estudio/sesiones/%d", libre.ID), nil).Status)
	resp = otro.JSON("GET", "/estudio/sesiones", nil)
	var ajenas []sesionTest
	resp.JSON(s.T(), &ajenas)
	s.Empty(ajenas)

	s.Equal(http.StatusNoContent, c.JSON("DELETE", ruta("/estudio/sesiones/%d", libre.ID), nil).Status)
	resp = c.JSON("GET", "/estudio/sesiones", nil)
	var mias []sesionTest
	resp.JSON(s.T(), &mias)
	s.Require().Len(mias, 1)
	s.Equal(pomodoro.ID, mias[0].ID)
}

func (s *EstudioHandlerSuite) TestRankingSoloConQuienParticipa() {
	ivan := NewClient(s.T(), s.app)
	ivan.Registrar("Iván Pérez", "ivan.estudio@chedul.com")
	juli := NewClient(s.T(), s.app)
	juli.Registrar("Juli Gómez", "juli.estudio@chedul.com")
	s.Require().Equal(http.StatusCreated, ivan.JSON("POST", "/estudio/sesiones", map[string]any{"modo": "libre", "minutos": 90}).Status)
	s.Require().Equal(http.StatusCreated, juli.JSON("POST", "/estudio/sesiones", map[string]any{"modo": "libre", "minutos": 40}).Status)

	ranking := func(c *Client, metodo string, payload any) rankingTest {
		resp := c.JSON(metodo, "/estudio/ranking", payload)
		s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
		var r rankingTest
		resp.JSON(s.T(), &r)
		return r
	}

	// Nadie se sumo todavia
	r := ranking(ivan, "GET", nil)
	s.False(r.Participo)
	s.Empty(r.Puestos)

	r = ranking(ivan, "PUT", map[string]any{"participar": true})
	s.True(r.Participo)
	s.Require().Len(r.Puestos, 1)
	s.Equal("Iván P.", r.Puestos[0].Nombre)
	s.Equal(90, r.Puestos[0].Minutos)
	s.True(r.Puestos[0].SoyYo)

	// Juli ve a Iván, pero ella no aparece hasta que se suma
	r = ranking(juli, "GET", nil)
	s.Require().Len(r.Puestos, 1)
	s.False(r.Puestos[0].SoyYo)

	r = ranking(juli, "PUT", map[string]any{"participar": true})
	s.Equal(2, r.Participantes)
	s.Equal("Juli G.", r.Puestos[1].Nombre)
	s.Equal(2, r.Puestos[1].Posicion)

	r = ranking(ivan, "PUT", map[string]any{"participar": false})
	s.False(r.Participo)
	s.Require().Len(r.Puestos, 1)
	s.Equal("Juli G.", r.Puestos[0].Nombre)

	resp := ivan.JSON("PUT", "/estudio/ranking", map[string]any{})
	s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))
}

func (s *EstudioHandlerSuite) TestRequiereSesion() {
	c := NewClient(s.T(), s.app)
	s.Equal(http.StatusUnauthorized, c.JSON("GET", "/estudio/resumen", nil).Status)
	s.Equal(http.StatusUnauthorized, c.JSON("GET", "/estudio/ranking", nil).Status)
}

func (s *EstudioHandlerSuite) TestEditarSoloLaMateriaDeUnaSesion() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Lola Paz", "lola.sesion@chedul.com")
	aed := c.MateriaID("isi-aed")
	so := c.MateriaID("isi-so")

	resp := c.JSON("POST", "/estudio/sesiones", map[string]any{"modo": "pomodoro", "minutos": 25, "materia_id": aed})
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	var original sesionTest
	resp.JSON(s.T(), &original)

	fin, err := time.Parse(time.RFC3339, original.Fin)
	s.Require().NoError(err)
	inicio, err := time.Parse(time.RFC3339, original.Inicio)
	s.Require().NoError(err)
	s.Equal(25*time.Minute, fin.Sub(inicio), "el inicio es el fin menos la duracion")

	// Cambia la materia y nada mas, aunque se mande otra cosa
	resp = c.JSON("PUT", ruta("/estudio/sesiones/%d", original.ID), map[string]any{
		"materia_id": so, "minutos": 1, "fin": "2020-01-01T00:00:00Z", "modo": "libre",
	})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var editada sesionTest
	resp.JSON(s.T(), &editada)
	s.Require().NotNil(editada.Materia)
	s.Equal(so, editada.Materia.ID)
	s.Equal(25, editada.Minutos)
	s.Equal("pomodoro", editada.Modo)
	s.Equal(original.Fin, editada.Fin)
	s.Equal(original.Inicio, editada.Inicio)

	// Sin materia
	resp = c.JSON("PUT", ruta("/estudio/sesiones/%d", original.ID), map[string]any{"materia_id": nil})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	resp.JSON(s.T(), &editada)
	s.Nil(editada.Materia)

	resp = c.JSON("PUT", ruta("/estudio/sesiones/%d", original.ID), map[string]any{"materia_id": 999999})
	s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))

	// No se puede tocar la sesion de otro alumno
	otro := NewClient(s.T(), s.app)
	otro.Registrar("Otro Alumno", "otro.sesion@chedul.com")
	resp = otro.JSON("PUT", ruta("/estudio/sesiones/%d", original.ID), map[string]any{"materia_id": aed})
	s.Equal(http.StatusNotFound, resp.Status, string(resp.Body))
}

func (s *EstudioHandlerSuite) TestTemporizadorSigueEntreDispositivos() {
	celu := NewClient(s.T(), s.app)
	celu.Registrar("Tomi Paz", "tomi.timer@chedul.com")
	// Mismo alumno desde otro dispositivo
	compu := NewClient(s.T(), s.app)
	compu.AccessToken = celu.AccessToken

	type respuesta struct {
		Estado *struct {
			Fase      string `json:"fase"`
			Acumulado int64  `json:"acumulado"`
			Desde     *int64 `json:"desde"`
		} `json:"estado"`
		Rev   int   `json:"rev"`
		Ahora int64 `json:"ahora"`
	}
	leer := func(c *Client) respuesta {
		resp := c.JSON("GET", "/estudio/temporizador", nil)
		s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
		var r respuesta
		resp.JSON(s.T(), &r)
		return r
	}
	guardar := func(c *Client, rev int, fase string, acumulado int64) (int, respuesta) {
		resp := c.JSON("PUT", "/estudio/temporizador", map[string]any{
			"rev": rev,
			"estado": map[string]any{
				"modo": "pomodoro", "fase": fase, "foco": 25, "descanso": 5,
				"materia_id": 0, "acumulado": acumulado, "desde": nil,
			},
		})
		var r respuesta
		if resp.Status == http.StatusOK || resp.Status == http.StatusConflict {
			resp.JSON(s.T(), &r)
		}
		return resp.Status, r
	}

	vacio := leer(celu)
	s.Nil(vacio.Estado)
	s.NotZero(vacio.Ahora)

	status, r := guardar(celu, 0, "foco", 60_000)
	s.Equal(http.StatusOK, status)
	s.Equal(1, r.Rev)

	// El otro dispositivo ve lo mismo
	visto := leer(compu)
	s.Require().NotNil(visto.Estado)
	s.Equal(int64(60_000), visto.Estado.Acumulado)

	// Los dos intentan cerrar el mismo foco: gana el primero, el otro recibe
	// lo guardado y no escribe
	status, r = guardar(celu, 1, "descanso", 0)
	s.Equal(http.StatusOK, status)
	s.Equal(2, r.Rev)
	status, r = guardar(compu, 1, "descanso", 0)
	s.Equal(http.StatusConflict, status)
	s.Equal(2, r.Rev)
	s.Require().NotNil(r.Estado)
	s.Equal("descanso", r.Estado.Fase)

	// Un estado invalido se rechaza
	resp := celu.JSON("PUT", "/estudio/temporizador", map[string]any{
		"rev":    2,
		"estado": map[string]any{"modo": "raro", "fase": "foco", "foco": 25, "descanso": 5, "acumulado": 0},
	})
	s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))

	// Otro alumno no ve el cronometro ajeno
	otro := NewClient(s.T(), s.app)
	otro.Registrar("Otro", "otro.timer@chedul.com")
	s.Nil(leer(otro).Estado)
}

func TestEstudioSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	suite.Run(t, new(EstudioHandlerSuite))
}

func (s *EstudioHandlerSuite) TestMetaDiaria() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Kari Luna", "kari.estudio@chedul.com")

	resp := c.JSON("PUT", "/estudio/meta", map[string]any{"minutos": 90})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	for _, minutos := range []int{0, 14, 721} {
		resp = c.JSON("PUT", "/estudio/meta", map[string]any{"minutos": minutos})
		s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))
	}

	resp = c.JSON("GET", "/estudio/resumen", nil)
	var resumen struct {
		MetaDiaria int `json:"meta_diaria"`
	}
	resp.JSON(s.T(), &resumen)
	s.Equal(90, resumen.MetaDiaria)
}

func (s *EstudioHandlerSuite) TestTareas() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Lara Paz", "lara.estudio@chedul.com")
	aed := c.MateriaID("isi-aed")

	type tareaTest struct {
		ID      int64  `json:"id"`
		Titulo  string `json:"titulo"`
		Hecha   bool   `json:"hecha"`
		Materia *struct {
			ID int64 `json:"id"`
		} `json:"materia"`
	}

	resp := c.JSON("POST", "/estudio/tareas", map[string]any{"titulo": "  Repasar listas  ", "materia_id": aed})
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	var repasar tareaTest
	resp.JSON(s.T(), &repasar)
	s.Equal("Repasar listas", repasar.Titulo)
	s.False(repasar.Hecha)
	s.Require().NotNil(repasar.Materia)
	s.Equal(aed, repasar.Materia.ID)

	resp = c.JSON("POST", "/estudio/tareas", map[string]any{"titulo": "Leer el apunte"})
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))

	for _, payload := range []map[string]any{
		{"titulo": "   "},
		{"titulo": "Algo", "materia_id": 999999},
	} {
		resp = c.JSON("POST", "/estudio/tareas", payload)
		s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))
	}

	resp = c.JSON("PUT", ruta("/estudio/tareas/%d", repasar.ID), map[string]any{"titulo": "Repasar listas", "materia_id": aed, "hecha": true})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	resp.JSON(s.T(), &repasar)
	s.True(repasar.Hecha)

	// Las pendientes van primero
	resp = c.JSON("GET", "/estudio/tareas", nil)
	var tareas []tareaTest
	resp.JSON(s.T(), &tareas)
	s.Require().Len(tareas, 2)
	s.Equal("Leer el apunte", tareas[0].Titulo)
	s.True(tareas[1].Hecha)

	// Otro alumno no las ve ni las toca
	otro := NewClient(s.T(), s.app)
	otro.Registrar("Mati", "mati.estudio@chedul.com")
	s.Equal(http.StatusNotFound, otro.JSON("PUT", ruta("/estudio/tareas/%d", repasar.ID), map[string]any{"titulo": "x"}).Status)
	s.Equal(http.StatusNotFound, otro.JSON("DELETE", ruta("/estudio/tareas/%d", repasar.ID), nil).Status)

	s.Equal(http.StatusNoContent, c.JSON("DELETE", ruta("/estudio/tareas/%d", repasar.ID), nil).Status)
	resp = c.JSON("GET", "/estudio/tareas", nil)
	resp.JSON(s.T(), &tareas)
	s.Len(tareas, 1)
}

func (s *EstudioHandlerSuite) TestNoMeInteresaSoloElectivas() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Nico Ruiz", "nico.electivas@chedul.com")
	noMeInteresa := c.CondicionID("No me interesa")
	electiva := c.MateriaID("isi-qca")
	obligatoria := c.MateriaID("isi-aed")

	resp := c.JSON("PUT", ruta("/condicion_alumno/%d", obligatoria), map[string]any{"condicion_id": noMeInteresa})
	s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))

	resp = c.JSON("PUT", ruta("/condicion_alumno/%d", electiva), map[string]any{"condicion_id": noMeInteresa})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))

	// La electiva descartada no aparece en ninguna lista del progreso
	resp = c.JSON("GET", "/alumnos/me/progreso", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var progreso map[string]any
	resp.JSON(s.T(), &progreso)
	for clave, valor := range progreso {
		lista, ok := valor.([]any)
		if !ok {
			continue
		}
		for _, m := range lista {
			s.NotEqual(float64(electiva), m.(map[string]any)["id"], "la electiva aparece en %s", clave)
		}
	}
}

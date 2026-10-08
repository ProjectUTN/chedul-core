package api

import (
	"bytes"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"
)

type aporteTest struct {
	ID          int64   `json:"id"`
	Titulo      string  `json:"titulo"`
	Descripcion string  `json:"descripcion"`
	Link        *string `json:"link"`
	Materia     struct {
		ID     int64  `json:"id"`
		Nombre string `json:"nombre"`
	} `json:"materia"`
	Tag struct {
		ID     int64  `json:"id"`
		Nombre string `json:"nombre"`
	} `json:"tag"`
	Autor struct {
		ID     int64  `json:"id"`
		Nombre string `json:"nombre"`
	} `json:"autor"`
	Archivo *struct {
		Nombre string `json:"nombre"`
		Tipo   string `json:"tipo"`
		Tamano int64  `json:"tamano"`
	} `json:"archivo"`
	Favoritos  int  `json:"favoritos"`
	EsFavorito bool `json:"es_favorito"`
	EsMio      bool `json:"es_mio"`
}

type listaAportesTest struct {
	Items []aporteTest `json:"items"`
	Total int          `json:"total"`
}

type AporteHandlerSuite struct {
	suite.Suite
	app *TestApp
}

func (s *AporteHandlerSuite) SetupSuite() {
	app, err := CreateTestApp()
	s.Require().NoError(err)
	s.app = app
}

func (s *AporteHandlerSuite) TearDownSuite() {
	s.app.Cleanup()
}

func (s *AporteHandlerSuite) tagID(c *Client, nombre string) int64 {
	resp := c.JSON("GET", "/aportes/tags", nil)
	s.Require().Equal(http.StatusOK, resp.Status)

	var tags []struct {
		ID     int64  `json:"id"`
		Nombre string `json:"nombre"`
	}
	resp.JSON(s.T(), &tags)
	for _, t := range tags {
		if t.Nombre == nombre {
			return t.ID
		}
	}
	s.FailNow("no existe el tag " + nombre)
	return 0
}

func (s *AporteHandlerSuite) listar(c *Client, query string) listaAportesTest {
	resp := c.JSON("GET", "/aportes"+query, nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var lista listaAportesTest
	resp.JSON(s.T(), &lista)
	return lista
}

func (s *AporteHandlerSuite) TestSubirArchivoListarYDescargar() {
	ana := NewClient(s.T(), s.app)
	ana.Registrar("Ana", "ana.aportes@chedul.com")

	am1 := ana.MateriaID("isi-am1")
	resumen := s.tagID(ana, "Resumen")

	pdf := []byte("%PDF-1.4 resumen de limites y derivadas")
	resp := ana.Multipart("POST", "/aportes", map[string]string{
		"titulo":      "Resumen primer parcial",
		"descripcion": "Límites, derivadas y estudio de funciones",
		"materia_id":  strconv.FormatInt(am1, 10),
		"tag_id":      strconv.FormatInt(resumen, 10),
	}, &Archivo{Nombre: "../../resumen AM1.pdf", Contenido: pdf})
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))

	var creado aporteTest
	resp.JSON(s.T(), &creado)
	s.Equal("Resumen primer parcial", creado.Titulo)
	s.Equal(am1, creado.Materia.ID)
	s.Equal("Resumen", creado.Tag.Nombre)
	s.Equal("Ana", creado.Autor.Nombre)

	// Con apellido, los demas ven nombre e inicial
	resp = ana.JSON("PUT", "/alumnos/me", map[string]any{"nombre": "Ana María", "apellido": "Pérez Gómez", "carrera_id": 1})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	for _, item := range s.listar(ana, "").Items {
		if item.ID == creado.ID {
			s.Equal("Ana P.", item.Autor.Nombre)
		}
	}
	s.True(creado.EsMio)
	s.Require().NotNil(creado.Archivo)
	s.Equal("resumen AM1.pdf", creado.Archivo.Nombre, "el nombre no conserva carpetas")
	s.Equal("application/pdf", creado.Archivo.Tipo)
	s.Equal(int64(len(pdf)), creado.Archivo.Tamano)

	// Otro alumno lo ve, lo descarga y lo marca como favorito
	beto := NewClient(s.T(), s.app)
	beto.Registrar("Beto", "beto.aportes@chedul.com")

	lista := s.listar(beto, "?materia_id="+strconv.FormatInt(am1, 10))
	s.Require().Equal(1, lista.Total)
	s.False(lista.Items[0].EsMio)

	resp = beto.Do("GET", ruta("/aportes/%d/archivo", creado.ID), nil, "")
	s.Require().Equal(http.StatusOK, resp.Status)
	s.True(bytes.Equal(pdf, resp.Body))
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))
	s.Contains(resp.Header.Get("Content-Disposition"), "attachment")

	resp = beto.JSON("POST", ruta("/aportes/%d/favorito", creado.ID), nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	resp = beto.JSON("POST", ruta("/aportes/%d/favorito", creado.ID), nil)
	s.Require().Equal(http.StatusOK, resp.Status, "marcar dos veces no falla")

	var fav struct {
		Favoritos  int  `json:"favoritos"`
		EsFavorito bool `json:"es_favorito"`
	}
	resp.JSON(s.T(), &fav)
	s.Equal(1, fav.Favoritos)
	s.True(fav.EsFavorito)

	s.Equal(1, s.listar(beto, "?favoritos=1").Total)
	s.Equal(0, s.listar(beto, "?mios=1").Total)
	s.Equal(1, s.listar(ana, "?mios=1").Total)

	resp = beto.JSON("DELETE", ruta("/aportes/%d/favorito", creado.ID), nil)
	s.Require().Equal(http.StatusOK, resp.Status)
	s.Equal(0, s.listar(beto, "?favoritos=1").Total)
}

func (s *AporteHandlerSuite) TestLinkSinArchivo() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Caro", "caro.aportes@chedul.com")

	materia := strconv.FormatInt(c.MateriaID("isi-aed"), 10)
	video := strconv.FormatInt(s.tagID(c, "Video"), 10)

	resp := c.Multipart("POST", "/aportes", map[string]string{
		"titulo": "Clase de punteros", "materia_id": materia, "tag_id": video,
		"link": "https://www.youtube.com/watch?v=abc",
	}, nil)
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))

	var creado aporteTest
	resp.JSON(s.T(), &creado)
	s.Nil(creado.Archivo)
	s.Require().NotNil(creado.Link)

	resp = c.Do("GET", ruta("/aportes/%d/archivo", creado.ID), nil, "")
	s.Equal(http.StatusNotFound, resp.Status)
}

func (s *AporteHandlerSuite) TestValidaciones() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Dani", "dani.aportes@chedul.com")

	materia := strconv.FormatInt(c.MateriaID("isi-aed"), 10)
	tag := strconv.FormatInt(s.tagID(c, "Otro"), 10)

	casos := []struct {
		nombre  string
		campos  map[string]string
		archivo *Archivo
	}{
		{"sin link ni archivo", map[string]string{"titulo": "x", "materia_id": materia, "tag_id": tag}, nil},
		{"link que no es http", map[string]string{"titulo": "x", "materia_id": materia, "tag_id": tag, "link": "javascript:alert(1)"}, nil},
		{"sin titulo", map[string]string{"materia_id": materia, "tag_id": tag, "link": "https://a.com"}, nil},
		{"materia inexistente", map[string]string{"titulo": "x", "materia_id": "99999", "tag_id": tag, "link": "https://a.com"}, nil},
		{"tag inexistente", map[string]string{"titulo": "x", "materia_id": materia, "tag_id": "99999", "link": "https://a.com"}, nil},
		{"extension no permitida", map[string]string{"titulo": "x", "materia_id": materia, "tag_id": tag}, &Archivo{Nombre: "virus.exe", Contenido: []byte("MZ")}},
		{"archivo muy grande", map[string]string{"titulo": "x", "materia_id": materia, "tag_id": tag}, &Archivo{Nombre: "grande.pdf", Contenido: bytes.Repeat([]byte("a"), 2*1024*1024)}},
	}

	for _, caso := range casos {
		resp := c.Multipart("POST", "/aportes", caso.campos, caso.archivo)
		s.Equal(http.StatusUnprocessableEntity, resp.Status, caso.nombre+": "+string(resp.Body))
	}
}

func (s *AporteHandlerSuite) TestSoloElAutorEditaYBorra() {
	autor := NewClient(s.T(), s.app)
	autor.Registrar("Eze", "eze.aportes@chedul.com")
	otro := NewClient(s.T(), s.app)
	otro.Registrar("Fede", "fede.aportes@chedul.com")

	materia := autor.MateriaID("isi-so")
	tag := s.tagID(autor, "Parcial")

	resp := autor.Multipart("POST", "/aportes", map[string]string{
		"titulo": "Parcial 2024", "materia_id": strconv.FormatInt(materia, 10),
		"tag_id": strconv.FormatInt(tag, 10), "link": "https://drive.google.com/x",
	}, nil)
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	var creado aporteTest
	resp.JSON(s.T(), &creado)

	cambios := map[string]any{
		"titulo": "Parcial 2024 resuelto", "descripcion": "Con soluciones",
		"materia_id": materia, "tag_id": tag, "link": "https://drive.google.com/y",
	}

	resp = otro.JSON("PUT", ruta("/aportes/%d", creado.ID), cambios)
	s.Equal(http.StatusForbidden, resp.Status)
	resp = otro.JSON("DELETE", ruta("/aportes/%d", creado.ID), nil)
	s.Equal(http.StatusForbidden, resp.Status)

	resp = autor.JSON("PUT", ruta("/aportes/%d", creado.ID), cambios)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var editado aporteTest
	resp.JSON(s.T(), &editado)
	s.Equal("Parcial 2024 resuelto", editado.Titulo)
	s.Equal("Con soluciones", editado.Descripcion)

	resp = autor.JSON("DELETE", ruta("/aportes/%d", creado.ID), nil)
	s.Equal(http.StatusNoContent, resp.Status)
	resp = autor.JSON("GET", ruta("/aportes/%d", creado.ID), nil)
	s.Equal(http.StatusNotFound, resp.Status)
}

func (s *AporteHandlerSuite) TestBusquedaYOrden() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Gabi", "gabi.aportes@chedul.com")

	materia := strconv.FormatInt(c.MateriaID("isi-bdd"), 10)
	tag := strconv.FormatInt(s.tagID(c, "Ejercicios"), 10)

	for _, titulo := range []string{"Guia de SQL", "Normalizacion 100% resuelta", "Ejercicios de joins"} {
		resp := c.Multipart("POST", "/aportes", map[string]string{
			"titulo": titulo, "materia_id": materia, "tag_id": tag, "link": "https://a.com",
		}, nil)
		s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	}

	s.Equal(1, s.listar(c, "?q=sql").Total, "busca sin distinguir mayusculas")
	s.Equal(1, s.listar(c, "?q=100%25").Total, "el % se busca literal")
	s.Equal(3, s.listar(c, "?q=base+de+datos").Total, "busca por nombre de materia")

	pagina := s.listar(c, "?materia_id="+materia+"&limite=2&pagina=2")
	s.Equal(3, pagina.Total)
	s.Len(pagina.Items, 1)

	recientes := s.listar(c, "?materia_id="+materia+"&orden=recientes")
	s.Equal("Ejercicios de joins", recientes.Items[0].Titulo)
}

func TestAporteSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	suite.Run(t, new(AporteHandlerSuite))
}

func TestAportesSoloLinks(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	cfg := createTestConfig()
	cfg.Uploads.Disabled = true
	app, err := CreateTestAppWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	c := NewClient(t, app)
	c.Registrar("Hana", "hana.aportes@chedul.com")

	resp := c.JSON("GET", "/aportes/config", nil)
	if resp.Status != http.StatusOK || !bytes.Contains(resp.Body, []byte(`"subida_archivos":false`)) {
		t.Fatalf("config inesperada: %d %s", resp.Status, resp.Body)
	}

	materia := strconv.FormatInt(c.MateriaID("isi-aed"), 10)
	campos := map[string]string{"titulo": "Apunte", "materia_id": materia, "tag_id": "1"}

	resp = c.Multipart("POST", "/aportes", campos, &Archivo{Nombre: "apunte.pdf", Contenido: []byte("%PDF")})
	if resp.Status != http.StatusUnprocessableEntity {
		t.Fatalf("con la subida desactivada un archivo tiene que dar 422, dio %d %s", resp.Status, resp.Body)
	}

	campos["link"] = "https://drive.google.com/file/d/abc"
	resp = c.Multipart("POST", "/aportes", campos, nil)
	if resp.Status != http.StatusCreated {
		t.Fatalf("un link tiene que poder subirse: %d %s", resp.Status, resp.Body)
	}
}

func (s *AporteHandlerSuite) TestAporteDeTodaLaCarrera() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Iara", "iara.aportes@chedul.com")
	otro := strconv.FormatInt(s.tagID(c, "Otro"), 10)

	resp := c.Multipart("POST", "/aportes", map[string]string{
		"titulo": "Plan de estudios ISI", "tag_id": otro, "link": "https://www.frre.utn.edu.ar/plan",
	}, nil)
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))
	var raw map[string]any
	resp.JSON(s.T(), &raw)
	s.Nil(raw["materia"], "sin materia es de toda la carrera")

	materia := strconv.FormatInt(c.MateriaID("isi-aed"), 10)
	resp = c.Multipart("POST", "/aportes", map[string]string{
		"titulo": "Resumen AED", "materia_id": materia, "tag_id": otro, "link": "https://a.com/aed",
	}, nil)
	s.Require().Equal(http.StatusCreated, resp.Status, string(resp.Body))

	s.Equal(2, s.listar(c, "?mios=1").Total)
	carrera := s.listar(c, "?mios=1&carrera=1")
	s.Require().Equal(1, carrera.Total)
	s.Equal("Plan de estudios ISI", carrera.Items[0].Titulo)

	// Se puede pasar a una materia y volver a toda la carrera
	id := carrera.Items[0].ID
	resp = c.JSON("PUT", ruta("/aportes/%d", id), map[string]any{
		"titulo": "Plan de estudios ISI", "tag_id": s.tagID(c, "Otro"), "link": "https://www.frre.utn.edu.ar/plan", "materia_id": c.MateriaID("isi-aed"),
	})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	s.Equal(0, s.listar(c, "?mios=1&carrera=1").Total)
	resp = c.JSON("PUT", ruta("/aportes/%d", id), map[string]any{
		"titulo": "Plan de estudios ISI", "tag_id": s.tagID(c, "Otro"), "link": "https://www.frre.utn.edu.ar/plan", "materia_id": 0,
	})
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	s.Equal(1, s.listar(c, "?mios=1&carrera=1").Total)
}

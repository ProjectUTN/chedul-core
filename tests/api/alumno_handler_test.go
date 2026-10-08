package api

import (
	"fmt"
	"net/http"
	"strings"
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

	for _, nota := range []int{5, 11} {
		resp = c.JSON("PUT", ruta("/condicion_alumno/%d", am1), map[string]any{"condicion_id": aprobada, "nota": nota})
		s.Equal(http.StatusUnprocessableEntity, resp.Status, "nota %d", nota)
	}

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

func (s *AlumnoHandlerSuite) TestCambiarPassword() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Pato", "pato.clave@chedul.com")

	resp := c.JSON("PUT", "/alumnos/me/password", map[string]any{"actual": "Otra-123", "nueva": "Nueva-Clave-9"})
	s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))
	resp = c.JSON("PUT", "/alumnos/me/password", map[string]any{"actual": "Chedul-123", "nueva": "corta"})
	s.Equal(http.StatusUnprocessableEntity, resp.Status, string(resp.Body))

	resp = c.JSON("PUT", "/alumnos/me/password", map[string]any{"actual": "Chedul-123", "nueva": "Nueva-Clave-9"})
	s.Require().Equal(http.StatusNoContent, resp.Status, string(resp.Body))

	otro := NewClient(s.T(), s.app)
	s.Equal(http.StatusUnauthorized, otro.JSON("POST", "/login", map[string]any{"email": "pato.clave@chedul.com", "password": "Chedul-123"}).Status)
	s.Equal(http.StatusOK, otro.JSON("POST", "/login", map[string]any{"email": "pato.clave@chedul.com", "password": "Nueva-Clave-9"}).Status)
}

func (s *AlumnoHandlerSuite) TestCambiarPasswordCierraLasOtrasSesiones() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Juli", "juli.sesiones@chedul.com")
	otra := NewClient(s.T(), s.app)
	s.Require().Equal(http.StatusOK, otra.JSON("POST", "/login", map[string]any{"email": "juli.sesiones@chedul.com", "password": "Chedul-123"}).Status)

	resp := c.JSON("PUT", "/alumnos/me/password", map[string]any{"actual": "Chedul-123", "nueva": "otra clave"})
	s.Require().Equal(http.StatusNoContent, resp.Status, string(resp.Body))

	s.Equal(http.StatusUnauthorized, otra.JSON("POST", "/refresh-token", nil).Status, "la otra sesion se cierra")
	s.Equal(http.StatusOK, c.JSON("POST", "/refresh-token", nil).Status, "quien cambio la clave sigue adentro")
}

func (s *AlumnoHandlerSuite) TestPanelAdmin() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Admin", "admin.panel@chedul.com")

	// Para los demas el panel no existe
	s.Equal(http.StatusNotFound, c.JSON("GET", "/admin/resumen", nil).Status)
	s.NotContains(string(c.JSON("GET", "/alumnos/me", nil).Body), "es_admin")

	s.app.HacerAdmin("admin.panel@chedul.com")
	s.Contains(string(c.JSON("GET", "/alumnos/me", nil).Body), `"es_admin":true`)

	resp := c.JSON("GET", "/admin/resumen", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var resumen struct {
		Alumnos         int `json:"alumnos"`
		Nuevos7         int `json:"nuevos_7_dias"`
		AccesosHoy      int `json:"accesos_hoy"`
		RegistrosPorDia []struct {
			Dia      string `json:"dia"`
			Cantidad int    `json:"cantidad"`
		} `json:"registros_por_dia"`
	}
	resp.JSON(s.T(), &resumen)
	s.GreaterOrEqual(resumen.Alumnos, 1)
	s.GreaterOrEqual(resumen.Nuevos7, 1)
	s.GreaterOrEqual(resumen.AccesosHoy, 1)
	s.Len(resumen.RegistrosPorDia, 30)
	s.GreaterOrEqual(resumen.RegistrosPorDia[29].Cantidad, 1)

	resp = c.JSON("GET", "/admin/alumnos?buscar=admin.panel", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	var lista struct {
		Alumnos []struct {
			Email        string  `json:"email"`
			UltimoAcceso *string `json:"ultimo_acceso"`
			EsAdmin      bool    `json:"es_admin"`
		} `json:"alumnos"`
		Total int `json:"total"`
	}
	resp.JSON(s.T(), &lista)
	s.Require().Equal(1, lista.Total)
	s.Equal("admin.panel@chedul.com", lista.Alumnos[0].Email)
	s.NotNil(lista.Alumnos[0].UltimoAcceso)
	s.True(lista.Alumnos[0].EsAdmin)

	resp = c.JSON("GET", "/admin/accesos", nil)
	s.Require().Equal(http.StatusOK, resp.Status, string(resp.Body))
	s.Contains(string(resp.Body), "admin.panel@chedul.com")

	// El contador es publico
	publico := NewClient(s.T(), s.app)
	resp = publico.JSON("GET", "/estadisticas", nil)
	s.Require().Equal(http.StatusOK, resp.Status)
	s.Contains(string(resp.Body), `"alumnos":`)
}

// pedidoDesde manda un POST JSON como si viniera de otra IP (por el proxy)
func (s *AlumnoHandlerSuite) pedidoDesde(ip, path, contentType, body string) int {
	req, err := http.NewRequest("POST", s.app.Address+"/api/v1"+path, strings.NewReader(body))
	s.Require().NoError(err)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Forwarded-For", ip+", 10.0.0.1")
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)
	resp.Body.Close()
	return resp.StatusCode
}

func (s *AlumnoHandlerSuite) TestLoginLimitaIntentosPorCorreo() {
	c := NewClient(s.T(), s.app)
	c.Registrar("Tobi", "tobi@chedul.com")

	// Registrar ya gasto un intento (el login que hace despues del registro)
	body := `{"email":"tobi@chedul.com","password":"incorrecta"}`
	for i := range 9 {
		s.Equal(http.StatusUnauthorized, s.pedidoDesde(fmt.Sprintf("198.51.100.%d", i), "/login", "application/json", body))
	}
	// Aunque cambie de IP, el correo ya no admite mas intentos por un rato
	s.Equal(http.StatusTooManyRequests, s.pedidoDesde("198.51.100.99", "/login", "application/json", body))
}

func (s *AlumnoHandlerSuite) TestLoginLimitaPedidosPorIP() {
	ip := "203.0.113.7"
	codigos := map[int]int{}
	for i := range 25 {
		body := fmt.Sprintf(`{"email":"nadie%d@chedul.com","password":"incorrecta"}`, i)
		codigos[s.pedidoDesde(ip, "/login", "application/json", body)]++
	}
	s.Equal(20, codigos[http.StatusUnauthorized])
	s.Equal(5, codigos[http.StatusTooManyRequests])
}

func (s *AlumnoHandlerSuite) TestLoginSoloAceptaJSON() {
	s.Equal(http.StatusUnsupportedMediaType,
		s.pedidoDesde("192.0.2.1", "/login", "application/x-www-form-urlencoded", "email=a@b.com&password=123456"))
}

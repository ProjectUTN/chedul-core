package handlers

import (
	"chedul-core/internals/domain"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CalendarioHandler struct {
	repo        domain.CalendarioRepository
	materiaRepo domain.MateriaRepository
	logger      *zap.Logger
}

func NewCalendarioHandler(repo domain.CalendarioRepository, materiaRepo domain.MateriaRepository, logger *zap.Logger) *CalendarioHandler {
	return &CalendarioHandler{repo: repo, materiaRepo: materiaRepo, logger: logger}
}

// validarMateria comprueba que la materia elegida (si hay) exista.
func (h *CalendarioHandler) validarMateria(c echo.Context, materiaID *int64, errs map[string]string) error {
	if materiaID == nil {
		return nil
	}
	if _, err := h.materiaRepo.GetByID(c.Request().Context(), *materiaID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		errs["materia_id"] = "La materia no existe"
	}
	return nil
}

func (h *CalendarioHandler) leerEvento(c echo.Context) (domain.DatosEvento, error) {
	var datos domain.DatosEvento
	if err := c.Bind(&datos); err != nil {
		return datos, InvalidJSON()
	}

	datos.Normalizar()
	errs := datos.Validate()
	if err := h.validarMateria(c, datos.MateriaID, errs); err != nil {
		return datos, err
	}
	if len(errs) > 0 {
		return datos, InvalidRequestData(errs)
	}
	return datos, nil
}

// La comision tiene que ser de la materia de la clase
func (h *CalendarioHandler) validarComision(c echo.Context, materiaID, comisionID *int64, errs map[string]string) error {
	if comisionID == nil {
		return nil
	}
	if materiaID == nil {
		errs["comision_id"] = "La comisión necesita una materia"
		return nil
	}
	comisiones, err := h.materiaRepo.GetComisiones(c.Request().Context(), *materiaID)
	if err != nil {
		return err
	}
	for _, co := range comisiones {
		if co.ID == *comisionID {
			return nil
		}
	}
	errs["comision_id"] = "La comisión no es de esa materia"
	return nil
}

func (h *CalendarioHandler) leerClase(c echo.Context) (domain.DatosClase, error) {
	var datos domain.DatosClase
	if err := c.Bind(&datos); err != nil {
		return datos, InvalidJSON()
	}

	datos.Normalizar()
	errs := datos.Validate()
	if err := h.validarMateria(c, datos.MateriaID, errs); err != nil {
		return datos, err
	}
	if err := h.validarComision(c, datos.MateriaID, datos.ComisionID, errs); err != nil {
		return datos, err
	}
	if len(errs) > 0 {
		return datos, InvalidRequestData(errs)
	}
	return datos, nil
}

// leerRango valida los parametros desde y hasta (AAAA-MM-DD) de un listado.
func leerRango(c echo.Context) (string, string, error) {
	desdeStr, hastaStr := c.QueryParam("desde"), c.QueryParam("hasta")
	errs := make(map[string]string)
	desde, err := time.Parse(domain.FormatoFecha, desdeStr)
	if err != nil {
		errs["desde"] = "La fecha tiene que tener el formato AAAA-MM-DD"
	}
	hasta, err := time.Parse(domain.FormatoFecha, hastaStr)
	if err != nil {
		errs["hasta"] = "La fecha tiene que tener el formato AAAA-MM-DD"
	}
	if len(errs) == 0 {
		dias := hasta.Sub(desde).Hours() / 24
		if dias < 0 {
			errs["hasta"] = "Tiene que ser igual o posterior a desde"
		} else if dias > domain.EventosRangoMaximoEnDias {
			errs["hasta"] = "El rango no puede ser de más de 400 días"
		}
	}
	if len(errs) > 0 {
		return "", "", InvalidRequestData(errs)
	}
	return desdeStr, hastaStr, nil
}

// ListFechasAcademicas devuelve el calendario de la facultad en el rango pedido.
func (h *CalendarioHandler) ListFechasAcademicas(c echo.Context) error {
	desde, hasta, err := leerRango(c)
	if err != nil {
		return err
	}

	fechas, err := h.repo.ListFechasAcademicas(c.Request().Context(), desde, hasta)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, fechas)
}

// ListEventos devuelve los eventos del alumno entre desde y hasta (inclusive),
// con formato AAAA-MM-DD.
func (h *CalendarioHandler) ListEventos(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	desdeStr, hastaStr, err := leerRango(c)
	if err != nil {
		return err
	}

	eventos, err := h.repo.ListEventos(c.Request().Context(), alumnoID, desdeStr, hastaStr)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, eventos)
}

func (h *CalendarioHandler) responderEvento(c echo.Context, status int, alumnoID, id int64) error {
	evento, err := h.repo.GetEvento(c.Request().Context(), alumnoID, id)
	if errors.Is(err, domain.ErrEventoNoEncontrado) {
		return NotFound("Evento")
	}
	if err != nil {
		return err
	}
	return c.JSON(status, evento)
}

func (h *CalendarioHandler) CreateEvento(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	datos, err := h.leerEvento(c)
	if err != nil {
		return err
	}

	id, err := h.repo.CreateEvento(c.Request().Context(), alumnoID, datos)
	if err != nil {
		return err
	}
	return h.responderEvento(c, http.StatusCreated, alumnoID, id)
}

func (h *CalendarioHandler) UpdateEvento(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}

	datos, err := h.leerEvento(c)
	if err != nil {
		return err
	}

	err = h.repo.UpdateEvento(c.Request().Context(), alumnoID, id, datos)
	if errors.Is(err, domain.ErrEventoNoEncontrado) {
		return NotFound("Evento")
	}
	if err != nil {
		return err
	}
	return h.responderEvento(c, http.StatusOK, alumnoID, id)
}

func (h *CalendarioHandler) DeleteEvento(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}

	err = h.repo.DeleteEvento(c.Request().Context(), alumnoID, id)
	if errors.Is(err, domain.ErrEventoNoEncontrado) {
		return NotFound("Evento")
	}
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *CalendarioHandler) ListClases(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	clases, err := h.repo.ListClases(c.Request().Context(), alumnoID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, clases)
}

func (h *CalendarioHandler) responderClase(c echo.Context, status int, alumnoID, id int64) error {
	clase, err := h.repo.GetClase(c.Request().Context(), alumnoID, id)
	if errors.Is(err, domain.ErrClaseNoEncontrada) {
		return NotFound("Clase")
	}
	if err != nil {
		return err
	}
	return c.JSON(status, clase)
}

func (h *CalendarioHandler) CreateClase(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	datos, err := h.leerClase(c)
	if err != nil {
		return err
	}

	id, err := h.repo.CreateClase(c.Request().Context(), alumnoID, datos)
	if err != nil {
		return err
	}
	return h.responderClase(c, http.StatusCreated, alumnoID, id)
}

func (h *CalendarioHandler) UpdateClase(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}

	datos, err := h.leerClase(c)
	if err != nil {
		return err
	}

	err = h.repo.UpdateClase(c.Request().Context(), alumnoID, id, datos)
	if errors.Is(err, domain.ErrClaseNoEncontrada) {
		return NotFound("Clase")
	}
	if err != nil {
		return err
	}
	return h.responderClase(c, http.StatusOK, alumnoID, id)
}

func (h *CalendarioHandler) DeleteClase(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}

	err = h.repo.DeleteClase(c.Request().Context(), alumnoID, id)
	if errors.Is(err, domain.ErrClaseNoEncontrada) {
		return NotFound("Clase")
	}
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type suscripcionCalendario struct {
	Token string `json:"token"`
}

func nuevoTokenCalendario() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetSuscripcion devuelve el token del link de calendario del alumno y lo
// crea si todavia no tiene. El frontend arma el link con el.
func (h *CalendarioHandler) GetSuscripcion(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	token, err := h.repo.CalendarioToken(c.Request().Context(), alumnoID)
	if err != nil {
		return err
	}
	if token == "" {
		return h.RenovarSuscripcion(c)
	}
	return c.JSON(http.StatusOK, suscripcionCalendario{Token: token})
}

// RenovarSuscripcion cambia el token: el link anterior deja de andar.
func (h *CalendarioHandler) RenovarSuscripcion(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	token, err := nuevoTokenCalendario()
	if err != nil {
		return err
	}
	if err := h.repo.SetCalendarioToken(c.Request().Context(), alumnoID, token); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, suscripcionCalendario{Token: token})
}

// ExportarICS es la ruta publica que leen Google Calendar y las demas apps.
// No lleva sesion: el token del link es lo que identifica al alumno.
func (h *CalendarioHandler) ExportarICS(c echo.Context) error {
	token := strings.TrimSuffix(c.Param("archivo"), ".ics")
	if token == "" {
		return NotFound("Calendario")
	}
	ctx := c.Request().Context()
	alumnoID, err := h.repo.AlumnoDelCalendario(ctx, token)
	if errors.Is(err, domain.ErrCalendarioNoEncontrado) {
		return NotFound("Calendario")
	}
	if err != nil {
		return err
	}

	ahora := time.Now().In(zonaArgentina)
	formato := domain.FormatoFecha
	eventos, err := h.repo.ListEventos(ctx, alumnoID, ahora.AddDate(0, -6, 0).Format(formato), ahora.AddDate(1, 0, 0).Format(formato))
	if err != nil {
		return err
	}
	clases, err := h.repo.ListClases(ctx, alumnoID)
	if err != nil {
		return err
	}
	inicioAnio := time.Date(ahora.Year(), time.January, 1, 0, 0, 0, 0, zonaArgentina)
	academicas, err := h.repo.ListFechasAcademicas(ctx, inicioAnio.Format(formato), inicioAnio.AddDate(2, 0, -1).Format(formato))
	if err != nil {
		return err
	}

	c.Response().Header().Set("Content-Disposition", `inline; filename="chedul.ics"`)
	c.Response().Header().Set("Cache-Control", "private, max-age=300")
	return c.Blob(http.StatusOK, "text/calendar; charset=utf-8", []byte(ArmarICS(eventos, clases, academicas, ahora)))
}

// ListEventosConfirmados devuelve los parciales, entregas y finales que
// cargaron varios compañeros de la comision y el alumno todavia no tiene.
func (h *CalendarioHandler) ListEventosConfirmados(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	hoy := time.Now().In(zonaArgentina).Format(domain.FormatoFecha)
	eventos, err := h.repo.ListEventosConfirmados(c.Request().Context(), alumnoID, hoy, domain.MinimoConfirmaciones)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, eventos)
}

// DesmentirEvento registra que el alumno dice que un evento confirmado no es
// asi. Deja de vercelo y, con tantos "no" como confirmaciones, nadie lo ve.
func (h *CalendarioHandler) DesmentirEvento(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	var datos domain.DatosDesmentido
	if err := c.Bind(&datos); err != nil {
		return InvalidJSON()
	}
	datos.Tipo = strings.ToLower(strings.TrimSpace(datos.Tipo))
	datos.Fecha = strings.TrimSpace(datos.Fecha)
	errs := datos.Validate()
	if err := h.validarMateria(c, &datos.MateriaID, errs); err != nil {
		return err
	}
	if len(errs) > 0 {
		return InvalidRequestData(errs)
	}
	if err := h.repo.DesmentirEvento(c.Request().Context(), alumnoID, datos); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

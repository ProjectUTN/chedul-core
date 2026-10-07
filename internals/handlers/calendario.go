package handlers

import (
	"chedul-core/internals/domain"
	"database/sql"
	"errors"
	"net/http"
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

// ListEventos devuelve los eventos del alumno entre desde y hasta (inclusive),
// con formato AAAA-MM-DD.
func (h *CalendarioHandler) ListEventos(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

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
		return InvalidRequestData(errs)
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

package handlers

import (
	"chedul-core/internals/domain"
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ComunidadHandler struct {
	repo        domain.ComunidadRepository
	materiaRepo domain.MateriaRepository
}

func NewComunidadHandler(repo domain.ComunidadRepository, materiaRepo domain.MateriaRepository) *ComunidadHandler {
	return &ComunidadHandler{repo: repo, materiaRepo: materiaRepo}
}

func comunidadNoEncontrada() ApiError {
	return NewApiError(http.StatusNotFound, errors.New("Comunidad no encontrada"))
}

func (h *ComunidadHandler) List(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	comunidades, err := h.repo.List(c.Request().Context(), alumnoID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, comunidades)
}

func (h *ComunidadHandler) Create(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	var datos domain.DatosComunidad
	if err := c.Bind(&datos); err != nil {
		return InvalidJSON()
	}
	datos.Normalizar()
	errs := datos.Validate()
	ctx := c.Request().Context()
	if datos.MateriaID != nil {
		if _, err := h.materiaRepo.GetByID(ctx, *datos.MateriaID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			errs["materia_id"] = "La materia no existe"
		}
	}
	if len(errs) > 0 {
		return InvalidRequestData(errs)
	}

	id, err := h.repo.Create(ctx, alumnoID, datos)
	if errors.Is(err, domain.ErrComunidadRepetida) {
		return InvalidRequestData(map[string]string{"link": "Esa comunidad ya está cargada"})
	}
	if err != nil {
		return err
	}
	comunidad, err := h.repo.Get(ctx, id, alumnoID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, comunidad)
}

func (h *ComunidadHandler) Delete(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}
	err = h.repo.Delete(c.Request().Context(), id, alumnoID)
	if errors.Is(err, domain.ErrComunidadNoEncontrada) {
		return comunidadNoEncontrada()
	}
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// Reportar marca un link roto o que no es de la carrera; con varios reportes se oculta
func (h *ComunidadHandler) Reportar(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	if _, err := h.repo.Get(ctx, id, alumnoID); err != nil {
		if errors.Is(err, domain.ErrComunidadNoEncontrada) {
			return comunidadNoEncontrada()
		}
		return err
	}
	if err := h.repo.Reportar(ctx, id, alumnoID); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

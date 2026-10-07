package handlers

import (
	"chedul-core/internals/domain"
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CondicionAlumnoHandler struct {
	repo          domain.CondicionAlumnoRepository
	condicionRepo domain.CondicionRepository
	materiaRepo   domain.MateriaRepository
	logger        *zap.Logger
}

func NewCondicionAlumnoHandler(repo domain.CondicionAlumnoRepository, condicionRepo domain.CondicionRepository, materiaRepo domain.MateriaRepository, logger *zap.Logger) *CondicionAlumnoHandler {
	return &CondicionAlumnoHandler{
		repo:          repo,
		condicionRepo: condicionRepo,
		materiaRepo:   materiaRepo,
		logger:        logger,
	}
}

// GetMisCondiciones devuelve la condicion del alumno autenticado en cada
// materia. Las materias sin condicion estan pendientes.
func (h *CondicionAlumnoHandler) GetMisCondiciones(c echo.Context) error {
	ctx := c.Request().Context()

	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	condiciones, err := h.repo.GetCondicionPorAlumno(ctx, alumnoID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, condiciones)
}

// SetCondicion guarda la condicion del alumno autenticado en una materia.
func (h *CondicionAlumnoHandler) SetCondicion(c echo.Context) error {
	ctx := c.Request().Context()

	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	materiaID, err := ParamID(c, "materia_id")
	if err != nil {
		return err
	}

	var req domain.SetCondicionRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	if errs := req.Validate(); len(errs) > 0 {
		return InvalidRequestData(errs)
	}

	if _, err := h.materiaRepo.GetByID(ctx, materiaID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NotFound("Materia")
		}
		return err
	}

	condiciones, err := h.condicionRepo.GetAll(ctx)
	if err != nil {
		return err
	}
	existe := false
	for _, cond := range condiciones {
		if cond.ID == req.CondicionID {
			existe = true
		}
	}
	if !existe {
		return InvalidRequestData(map[string]string{"condicion_id": "La condición no existe"})
	}

	condicion := domain.CondicionAlumno{
		CondicionID: req.CondicionID,
		Nota:        req.Nota,
		AlumnoID:    alumnoID,
		MateriaID:   materiaID,
	}

	if err := h.repo.Set(ctx, &condicion); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, condicion)
}

// DeleteCondicion vuelve la materia a "pendiente" para el alumno autenticado.
func (h *CondicionAlumnoHandler) DeleteCondicion(c echo.Context) error {
	ctx := c.Request().Context()

	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	materiaID, err := ParamID(c, "materia_id")
	if err != nil {
		return err
	}

	if err := h.repo.Delete(ctx, alumnoID, materiaID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

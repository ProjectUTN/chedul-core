package handlers

import (
	"chedul-core/internals/domain"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CondicionAlumnoHandler struct {
	repo   domain.CondicionAlumnoRepository
	logger *zap.Logger
}

func NewCondicionAlumnoHandle(repo domain.CondicionAlumnoRepository, logger *zap.Logger) *CondicionAlumnoHandler {
	return &CondicionAlumnoHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *CondicionAlumnoHandler) GetCondicionPorAlumno(c echo.Context) error {
	ctx := c.Request().Context()

	alumnoId, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de alumno inválido"))
	}

	condiciones, err := h.repo.GetCondicionPorAlumno(ctx, alumnoId)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, condiciones)

}

func (h *CondicionAlumnoHandler) SetCondicionAlumno(c echo.Context) error {
	ctx := c.Request().Context()

	// TODO: Aca habria que validar esto mejor y devolver todos los errores de una con InvalidRequestData
	alumnoID, err := strconv.ParseInt(c.Param("alumno_id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de alumno inválido"))
	}
	materiaID, err := strconv.ParseInt(c.Param("materia_id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de materia inválido"))
	}

	var req domain.SetCondicionRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	condicion := domain.CondicionAlumno{
		CondicionID: req.CondicionID,
		Nota:        req.Nota,
		AlumnoID:    alumnoID,
		MateriaID:   materiaID,
	}

	if err := h.repo.Create(ctx, &condicion); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, condicion)

}

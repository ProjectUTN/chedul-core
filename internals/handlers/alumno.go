package handlers

import (
	"chedul-core/internals/domain"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AlumnoHandler struct {
	service domain.AlumnoService
	logger  *zap.Logger
}

func NewAlumnoHandler(service domain.AlumnoService, logger *zap.Logger) *AlumnoHandler {
	return &AlumnoHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AlumnoHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	alumnos, err := h.service.GetAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, alumnos)
}

func (h *AlumnoHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}

	alumno, err := h.service.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, alumno)
}

func (h *AlumnoHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var req domain.AlumnoRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	alumno, err := h.service.Create(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, alumno)
}

func (h *AlumnoHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}

	var req domain.AlumnoRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	alumno, err := h.service.Update(ctx, id, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, alumno)
}

func (h *AlumnoHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return InvalidJSON()
	}

	if err := h.service.Delete(ctx, id); err != nil {
		h.logger.Error("failed to delete alumno", zap.Error(err), zap.Int64("id", id))
		return err
	}

	return c.JSON(http.StatusOK, "Alumno deleted successfully")
}

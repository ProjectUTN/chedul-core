package handlers

import (
	"chedul-core/internals/domain"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AlumnoHandler struct {
	alumnoRepo  domain.AlumnoRepository
	carreraRepo domain.CarreraRepository
	logger      *zap.Logger
}

func NewAlumnoHandler(alumnoRepo domain.AlumnoRepository, carreraRepo domain.CarreraRepository, logger *zap.Logger) *AlumnoHandler {
	return &AlumnoHandler{
		alumnoRepo:  alumnoRepo,
		logger:      logger,
		carreraRepo: carreraRepo,
	}
}

func (h *AlumnoHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	alumnos, err := h.alumnoRepo.GetAll(ctx)
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

	alumno, err := h.alumnoRepo.GetByID(ctx, id)
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

	if errors := req.Validate(); len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	if existing, _ := h.alumnoRepo.GetByEmail(ctx, req.Email); existing != nil {
		// TODO: Cambiar el error
		return InvalidJSON()
	}

	carrera, err := h.carreraRepo.GetByName(ctx, req.Carrera)
	if err != nil {
		// TODO: Cambiar el error
		return InvalidJSON()
	}

	alumno := &domain.Alumno{
		Nombre:  strings.TrimSpace(req.Nombre),
		Email:   strings.ToLower(strings.TrimSpace(req.Email)),
		Carrera: carrera.ID,
	}

	if err := h.alumnoRepo.Create(ctx, alumno); err != nil {
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

	if errors := req.Validate(); len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	alumno, err := h.alumnoRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	existing, _ := h.alumnoRepo.GetByEmail(ctx, req.Email)
	if existing != nil && existing.ID != id {
		// TODO: Mejorar el error
		return InvalidJSON()
	}
	alumno.Email = req.Email

	carrera, err := h.carreraRepo.GetByName(ctx, req.Carrera)
	if err != nil {
		// TODO: mejorar el error
		return InvalidJSON()
	}
	alumno.Carrera = carrera.ID

	if err := h.alumnoRepo.Update(ctx, alumno); err != nil {
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

	if err := h.alumnoRepo.Delete(ctx, id); err != nil {
		h.logger.Error("failed to delete alumno", zap.Error(err), zap.Int64("id", id))
		return err
	}

	return c.JSON(http.StatusOK, "Alumno deleted successfully")
}

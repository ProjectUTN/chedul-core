package handlers

import (
	"chedul-core/internals/domain"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type MateriaHandler struct {
	repo   domain.MateriaRepository
	logger *zap.Logger
}

func NewMateriaHandler(repo domain.MateriaRepository, logger *zap.Logger) *MateriaHandler {
	return &MateriaHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *MateriaHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	materias, err := h.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, materias)
}

func (h *MateriaHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return InvalidJSON()
	}

	materia, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, materia)
}

package handlers

import (
	"chedul-core/internals/domain"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CarreraHandler struct {
	repo   domain.CarreraRepository
	logger *zap.Logger
}

func NewCarreraHandler(repo domain.CarreraRepository, logger *zap.Logger) *CarreraHandler {
	return &CarreraHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *CarreraHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	carreras, err := h.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, carreras)
}

func (h *CarreraHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}

	carrera, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, carrera)
}

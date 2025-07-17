package handlers

import (
	"chedul-core/internals/domain"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CarreraHandler struct {
	service domain.CarreraService
	logger  *zap.Logger
}

func NewCarreraHandler(service domain.CarreraService, logger *zap.Logger) *CarreraHandler {
	return &CarreraHandler{
		service: service,
		logger:  logger,
	}
}

func (h *CarreraHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	carreras, err := h.service.GetAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, carreras)
}

func (h *CarreraHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return InvalidJSON()
	}

	carrera, err := h.service.GetByID(ctx, id)
	if err != nil {

		h.logger.Error("failed to get carrera", zap.Error(err), zap.Int64("id", id))
		return err
	}

	return c.JSON(http.StatusOK, carrera)
}

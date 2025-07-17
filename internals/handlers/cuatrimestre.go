package handlers

import (
	"chedul-core/internals/domain"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CuatrimestreHandler struct {
	service domain.CuatrimestreService
	logger  *zap.Logger
}

func NewCuatrimestreHandler(service domain.CuatrimestreService, logger *zap.Logger) *CuatrimestreHandler {
	return &CuatrimestreHandler{
		service: service,
		logger:  logger,
	}
}

func (h *CuatrimestreHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	cuatrimestres, err := h.service.GetAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, cuatrimestres)
}

func (h *CuatrimestreHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}

	cuatrimestre, err := h.service.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, cuatrimestre)
}

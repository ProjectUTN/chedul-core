package handlers

import (
	"chedul-core/internals/domain"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CuatrimestreHandler struct {
	repo   domain.CuatrimestreRepository
	logger *zap.Logger
}

func NewCuatrimestreHandler(repo domain.CuatrimestreRepository, logger *zap.Logger) *CuatrimestreHandler {
	return &CuatrimestreHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *CuatrimestreHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	cuatrimestres, err := h.repo.GetAll(ctx)
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

	cuatrimestre, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, cuatrimestre)
}

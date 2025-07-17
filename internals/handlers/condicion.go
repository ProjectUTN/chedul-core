package handlers

import (
	"chedul-core/internals/domain"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CondicionHandler struct {
	repo   domain.CondicionRepository
	logger *zap.Logger
}

func NewCondicionHandler(repo domain.CondicionService, logger *zap.Logger) *CondicionHandler {
	return &CondicionHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *CondicionHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	carreras, err := h.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, carreras)
}

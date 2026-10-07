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

// GetAll lista las materias con sus correlativas. Con ?carrera_id=N se
// filtran las materias de esa carrera.
func (h *MateriaHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	var (
		materias []domain.Materia
		err      error
	)

	if carreraParam := c.QueryParam("carrera_id"); carreraParam != "" {
		carreraID, parseErr := strconv.ParseInt(carreraParam, 10, 64)
		if parseErr != nil {
			return InvalidRequestData(map[string]string{"carrera_id": "Debe ser un número"})
		}
		materias, err = h.repo.GetByCarrera(ctx, carreraID)
	} else {
		materias, err = h.repo.GetAll(ctx)
	}

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, materias)
}

func (h *MateriaHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}

	materia, err := h.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, materia)
}

package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type MateriasPorCarrera struct {
	bun.BaseModel `bun:"table:materiasporcarrera"`
	ID            int64 `json:"id" bun:"id,pk,autoincrement"`
	CarreraID     int64 `json:"carrera_id" bun:"carrera_id,notnull"`
	MateriaID     int64 `json:"materia_id" bun:"materia_id,notnull"`
}

func HandleAsociarMateriaACarrera(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	carreraID, err := strconv.ParseInt(c.Param("carrera_id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de carrera inválido"))
	}

	materiaID, err := strconv.ParseInt(c.Param("materia_id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de materia inválido"))
	}

	asociacion := MateriasPorCarrera{
		CarreraID: carreraID,
		MateriaID: materiaID,
	}

	_, err = conn.NewInsert().Model(&asociacion).Exec(ctx.Request().Context())
	if err != nil {

		return err
	}

	return c.JSON(http.StatusCreated, asociacion)
}

func HandleDesasociarMateriaDeCarrera(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	carreraID, err := strconv.ParseInt(c.Param("carrera_id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de carrera inválido"))
	}

	materiaID, err := strconv.ParseInt(c.Param("materia_id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de materia inválido"))
	}

	res, err := conn.NewDelete().
		Model((*MateriasPorCarrera)(nil)).
		Where("carrera_id = ? AND materia_id = ?", carreraID, materiaID).
		Exec(ctx.Request().Context())

	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return NewApiError(http.StatusNotFound, fmt.Errorf("asociación no encontrada para eliminar"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Asociación eliminada exitosamente"})
}
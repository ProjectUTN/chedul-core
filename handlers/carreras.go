package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)


type Carrera struct {
	bun.BaseModel `bun:"table:carrera"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull,unique"`
}


type CarreraRequest struct {
	Nombre string `json:"nombre"`
}


func HandleGetCarreras(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn
	var carreras []Carrera
	err := conn.NewSelect().Model(&carreras).Scan(ctx.Request().Context())
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, []Carrera{})
		}
		return err
	}
	return c.JSON(http.StatusOK, carreras)
}

func HandlePostCarrera(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	var req CarreraRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	if req.Nombre == "" {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("el nombre no puede estar vacío"))
	}

	carrera := Carrera{
		Nombre: req.Nombre,
	}

	_, err := conn.NewInsert().Model(&carrera).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, carrera)
}

func HandleDeleteCarrera(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID inválido"))
	}

	res, err := conn.NewDelete().Model((*Carrera)(nil)).Where("id = ?", id).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return NewApiError(http.StatusNotFound, fmt.Errorf("Carrera no encontrada para eliminar"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Carrera eliminada exitosamente"})
}

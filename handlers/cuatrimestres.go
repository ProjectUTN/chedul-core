package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type Cuatrimestre struct {
	bun.BaseModel `bun:"table:cuatrimestre"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull,unique"`
}


type CuatrimestreRequest struct {
	Nombre string `json:"nombre"`
}



func HandleGetCuatrimestres(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn
	var cuatrimestres []Cuatrimestre
	err := conn.NewSelect().Model(&cuatrimestres).Scan(ctx.Request().Context())
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, []Cuatrimestre{})
		}
		return err
	}
	return c.JSON(http.StatusOK, cuatrimestres)
}

func HandlePostCuatrimestre(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	var req CuatrimestreRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	if req.Nombre == "" {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("el nombre no puede estar vacío"))
	}

	cuatrimestre := Cuatrimestre{
		Nombre: req.Nombre,
	}

	_, err := conn.NewInsert().Model(&cuatrimestre).Exec(ctx.Request().Context())
	if err != nil {
		return err 
	}

	return c.JSON(http.StatusCreated, cuatrimestre)
}

func HandleDeleteCuatrimestre(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID inválido"))
	}

	res, err := conn.NewDelete().Model((*Cuatrimestre)(nil)).Where("id = ?", id).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return NewApiError(http.StatusNotFound, fmt.Errorf("Cuatrimestre no encontrado para eliminar"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Cuatrimestre eliminado exitosamente"})
}

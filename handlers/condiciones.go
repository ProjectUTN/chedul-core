package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)


type Condicion struct {
	bun.BaseModel `bun:"table:condicion"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Condicion     string `json:"condicion" bun:"condicion,notnull,unique"`
}


type CondicionRequest struct {
	Condicion string `json:"condicion"`
}

func HandleGetCondiciones(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn
	var condiciones []Condicion
	err := conn.NewSelect().Model(&condiciones).Scan(ctx.Request().Context())
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, []Condicion{})
		}
		return err
	}
	return c.JSON(http.StatusOK, condiciones)
}

func HandlePostCondicion(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	var req CondicionRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	if req.Condicion == "" {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("el nombre de la condición no puede estar vacío"))
	}

	condicion := Condicion{
		Condicion: req.Condicion,
	}

	_, err := conn.NewInsert().Model(&condicion).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, condicion)
}

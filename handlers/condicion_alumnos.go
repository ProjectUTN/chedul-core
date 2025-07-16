package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)


type CondicionAlumno struct {
	bun.BaseModel `bun:"table:condicion_alumno"`
	ID            int64 `json:"id" bun:"id,pk,autoincrement"`
	CondicionID   int64 `json:"condicion_id" bun:"condicion_id,notnull"`
	Nota          *int  `json:"nota,omitempty" bun:"nota"`
	AlumnoID      int64 `json:"alumno_id" bun:"alumno_id,notnull"`
	MateriaID     int64 `json:"materia_id" bun:"materia_id,notnull"`
}

type SetCondicionRequest struct {
	CondicionID int64 `json:"condicion_id"`
	Nota        *int  `json:"nota"`
}


type CondicionAlumnoResponse struct {
	bun.BaseModel `bun:"table:condicion_alumno"`
	Nota          *int      `json:"nota,omitempty"`
	Materia       *Materia  `json:"materia" bun:"rel:belongs-to,join:materia_id=id"`
	Condicion     *Condicion `json:"condicion" bun:"rel:belongs-to,join:condicion_id=id"`
}


func HandleSetCondicionAlumno(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	alumnoID, err := strconv.ParseInt(c.Param("alumno_id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de alumno inválido"))
	}
	materiaID, err := strconv.ParseInt(c.Param("materia_id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de materia inválido"))
	}

	var req SetCondicionRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	condicion := CondicionAlumno{
		AlumnoID:    alumnoID,
		MateriaID:   materiaID,
		CondicionID: req.CondicionID,
		Nota:        req.Nota,
	}

	_, err = conn.NewInsert().
		Model(&condicion).
		On("CONFLICT (alumno_id, materia_id) DO UPDATE").
		Set("condicion_id = EXCLUDED.condicion_id, nota = EXCLUDED.nota").
		Exec(ctx.Request().Context())

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, condicion)
}

func HandleGetCondicionesPorAlumno(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	alumnoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de alumno inválido"))
	}

	var condiciones []CondicionAlumnoResponse
	err = conn.NewSelect().
		Model(&condiciones).
		Relation("Materia").
		Relation("Condicion").
		Where("alumno_id = ?", alumnoID).
		Scan(ctx.Request().Context())

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return c.JSON(http.StatusOK, condiciones)
}

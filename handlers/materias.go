package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)


type Materia struct {
	bun.BaseModel  `bun:"table:materia"`
	ID             int64   `json:"id" bun:"id,pk,autoincrement"`
	Nombre         string  `json:"nombre" bun:"nombre,notnull"`
	CargaHoraria   int     `json:"carga_horaria" bun:"carga_horaria,notnull"`
	CorrelativaID  *int64  `json:"correlativa_id,omitempty" bun:"correlativa_id"`
	Nivel          int     `json:"nivel" bun:"nivel,notnull"`
	Area           string  `json:"area" bun:"area,notnull"`
	Tipo           string  `json:"tipo" bun:"tipo,notnull"`
	CuatrimestreID int64   `json:"cuatrimestre_id" bun:"cuatrimestre_id,notnull"`
	Horas          float64 `json:"horas" bun:"horas"`
	Bloque         string  `json:"bloque" bun:"bloque,notnull"`
	Programa       string  `json:"programa" bun:"programa,notnull"`
}


type MateriaRequest struct {
	Nombre         string  `json:"nombre"`
	CargaHoraria   int     `json:"carga_horaria"`
	CorrelativaID  *int64  `json:"correlativa_id"`
	Nivel          int     `json:"nivel"`
	Area           string  `json:"area"`
	Tipo           string  `json:"tipo"`
	CuatrimestreID int64   `json:"cuatrimestre_id"`
	Horas          float64 `json:"horas"`
	Bloque         string  `json:"bloque"`
	Programa       string  `json:"programa"`
}



func HandleGetMaterias(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn
	materias := []Materia{}
	err := conn.NewSelect().Model(&materias).Scan(ctx.Request().Context())
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, []Materia{})
		}
		return err
	}
	return c.JSON(http.StatusOK, materias)
}

func HandleGetMateriaByID(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID inválido"))
	}
	var materia Materia
	err = conn.NewSelect().Model(&materia).Where("id=?", id).Scan(ctx.Request().Context())
	if err != nil {
		if err == sql.ErrNoRows {
			return NewApiError(http.StatusNotFound, fmt.Errorf("Materia no encontrada"))
		}
		return err
	}
	return c.JSON(http.StatusOK, materia)
}


func HandlePostMateria(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	var req MateriaRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	materia := Materia{
		Nombre:         req.Nombre,
		CargaHoraria:   req.CargaHoraria,
		CorrelativaID:  req.CorrelativaID,
		Nivel:          req.Nivel,
		Area:           req.Area,
		Tipo:           req.Tipo,
		CuatrimestreID: req.CuatrimestreID,
		Horas:          req.Horas,
		Bloque:         req.Bloque,
		Programa:       req.Programa,
	}

	_, err := conn.NewInsert().Model(&materia).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, materia)
}



func HandlePutMateria(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID inválido"))
	}

	var req MateriaRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	var materia Materia
	if err := conn.NewSelect().Model(&materia).Where("id = ?", id).Scan(ctx.Request().Context()); err != nil {
		if err == sql.ErrNoRows {
			return NewApiError(http.StatusNotFound, fmt.Errorf("Materia no encontrada para actualizar"))
		}
		return err
	}

	materia.Nombre = req.Nombre
	materia.CargaHoraria = req.CargaHoraria
	materia.CorrelativaID = req.CorrelativaID // Se actualiza el puntero
	materia.Nivel = req.Nivel
	materia.Area = req.Area
	materia.Tipo = req.Tipo
	materia.CuatrimestreID = req.CuatrimestreID
	materia.Horas = req.Horas
	materia.Bloque = req.Bloque
	materia.Programa = req.Programa

	_, err = conn.NewUpdate().Model(&materia).Where("id = ?", id).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, materia)
}



func HandleDeleteMateria(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID inválido"))
	}

	res, err := conn.NewDelete().Model((*Materia)(nil)).Where("id = ?", id).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return NewApiError(http.StatusNotFound, fmt.Errorf("Materia no encontrada para eliminar"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Materia eliminada exitosamente"})
}

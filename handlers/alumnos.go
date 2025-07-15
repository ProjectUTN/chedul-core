package handlers

import (
	"chedul-core/db"
	"chedul-core/domain"
	"chedul-core/logger"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type Alumno struct {
	bun.BaseModel `bun:"table:alumno"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull"`
	Email         string `json:"email" bun:"email,unique"`
	Carrera       int64  `json:"carrera" bun:"carrera_id,notnull"`
}

type UserRequest struct {
	Nombre  string `json:"nombre"`
	Email   string `json:"email"`
	Carrera string `json:"carrera"`
}

func (self *UserRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if _, err := domain.NewAlumnoName(self.Nombre); err != nil {
		errors["nombre"] = err.Error()
	}

	if _, err := domain.NewAlumnoEmail(self.Email); err != nil {
		errors["email"] = err.Error()
	}

	return errors
}

func HandleGetAlumnos(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()
	logger := logger.GetLoggerFromCtx(c)

	logger.Info("Buscando todos los alumnos")

	users := []Alumno{}

	err := conn.NewSelect().Model(&users).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, []Alumno{})
		}

		return err
	}

	return c.JSON(http.StatusOK, users)
}

func HandleGetAlumno(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return err
	}

	var user Alumno
	err = conn.NewSelect().Model(&user).Where("id=?", id).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, "Usuario no encontrado")
		}
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func HandlePostAlumno(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	var req UserRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	if errors := req.Validate(); len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	var carrera_id int64
	if err := conn.NewRaw("select id from ? where nombre = ?", bun.Ident("carrera"), req.Carrera).Scan(ctx, &carrera_id); err != nil {
		return err
	}

	user := Alumno{
		Nombre:  req.Nombre,
		Email:   req.Email,
		Carrera: carrera_id,
	}

	_, err := conn.NewInsert().
		Model(&user).
		Exec(ctx)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, user)
}

func HandlePutAlumno(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return err
	}

	var req UserRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	if errors := req.Validate(); len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	var user Alumno
	err = conn.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, "Usuario no encontrado")
		}
		return err
	}

	updateQuery := conn.NewUpdate().
		Model(&user).
		Where("id = ?", id)

	updated := false
	if req.Nombre != "" {
		updateQuery = updateQuery.Set("name = ?", req.Nombre)
		user.Nombre = req.Nombre
		updated = true
	}

	if req.Email != "" {
		updateQuery = updateQuery.Set("email = ?", req.Email)
		user.Email = req.Email
		updated = true
	}

	if !updated {
		return InvalidJSON()
	}

	_, err = updateQuery.Exec(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)

}

func HandleDeleteAlumno(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return err
	}

	var user Alumno
	err = conn.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, "Usuario no encontrado")
		}
		return err
	}

	_, err = conn.NewDelete().
		Model((*Alumno)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK,
		"Usuario eliminado exitosamente",
	)
}

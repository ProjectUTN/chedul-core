package handlers

import (
	"chedul-core/db"
	"chedul-core/domain"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type Alumno struct {
	bun.BaseModel `bun:"table:alumno"`
	ID            int64              `json:"id" bun:"id,pk,autoincrement"`
	Nombre        domain.AlumnoName  `json:"nombre" bun:"nombre,notnull"`
	Email         domain.AlumnoEmail `json:"email" bun:"email,unique"`
	Carrera       int64              `json:"carrera" bun:"carrera_id,notnull"`
}

type CreateUserRequest struct {
	Nombre  domain.AlumnoName  `json:"nombre"`
	Email   domain.AlumnoEmail `json:"email"`
	Carrera string             `json:"carrera"`
}

type UpdateUserRequest struct {
	Nombre  domain.AlumnoName  `json:"nombre"`
	Email   domain.AlumnoEmail `json:"email"`
	Carrera string             `json:"carrera"`
}

func HandleGetAlumnos(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	users := []Alumno{}

	err := conn.NewSelect().Model(&users).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, []Alumno{})
		}

		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, users)

}

func HandleGetAlumno(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "ID de usuario inválido")
	}

	var user Alumno
	err = conn.NewSelect().Model(&user).Where("id=?", id).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, "Usuario no encontrado")
		}
		return c.JSON(http.StatusInternalServerError, "Fallo al obtener usuario")
	}

	return c.JSON(http.StatusOK, user)
}

func HandlePostAlumno(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		// TODO: Tendria que loggear el error en INFO o DEBUG
		return c.JSON(http.StatusBadRequest, "Datos de solicitud inválidos")
	}

	var carrera_id int64
	if err := conn.NewRaw("select id from ? where nombre = ?", bun.Ident("carrera"), req.Carrera).Scan(ctx, &carrera_id); err != nil {
		return c.JSON(http.StatusInternalServerError, "Fallo al crear usuario")
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
		return c.JSON(http.StatusInternalServerError, "Fallo al crear usuario")
	}

	return c.JSON(http.StatusCreated, user)
}

func HandlePutAlumno(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "ID de usuario inválido")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "Datos de solicitud inválidos")
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
		return c.JSON(http.StatusInternalServerError, "Fallo al verificar usuario")
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
		return c.JSON(http.StatusBadRequest, "No se proporcionaron campos para actualizar")
	}

	_, err = updateQuery.Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Fallo al actualizar usuario")
	}

	return c.JSON(http.StatusOK, user)

}

func HandleDeleteAlumno(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "ID de usuario inválido")
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
		return c.JSON(http.StatusInternalServerError, "Fallo al verificar usuario")
	}

	_, err = conn.NewDelete().
		Model((*Alumno)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Fallo al eliminar usuario")
	}

	return c.JSON(http.StatusOK,
		"Usuario eliminado exitosamente",
	)
}

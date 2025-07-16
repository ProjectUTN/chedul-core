package handlers

import (
	"chedul-core/domain"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

// Alumno es el modelo de la base de datos.
type Alumno struct {
	bun.BaseModel `bun:"table:alumno"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull"`
	Email         string `json:"email" bun:"email,unique"`
	Carrera       int64  `json:"carrera" bun:"carrera_id,notnull"`
}

// UserRequest ahora espera un carrera_id numérico.
type UserRequest struct {
	Nombre    string `json:"nombre"`
	Email     string `json:"email"`
	CarreraID int64  `json:"carrera_id"`
}

// Validate ahora también comprueba que el carrera_id sea válido.
func (self *UserRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if _, err := domain.NewAlumnoName(self.Nombre); err != nil {
		errors["nombre"] = err.Error()
	}
	if _, err := domain.NewAlumnoEmail(self.Email); err != nil {
		errors["email"] = err.Error()
	}
	// --- AÑADIDO: Validación para carrera_id ---
	if self.CarreraID <= 0 {
		errors["carrera_id"] = "el ID de la carrera es requerido y debe ser un número positivo"
	}
	return errors
}

func HandleGetAlumnos(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn
	users := []Alumno{}
	err := conn.NewSelect().Model(&users).Scan(ctx.Request().Context())
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, []Alumno{})
		}
		return err
	}
	return c.JSON(http.StatusOK, users)
}

func HandleGetAlumno(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return err
	}
	var user Alumno
	err = conn.NewSelect().Model(&user).Where("id=?", id).Scan(ctx.Request().Context())
	if err != nil {
		if err == sql.ErrNoRows {
			return NewApiError(http.StatusNotFound, fmt.Errorf("Usuario no encontrado"))
		}
		return err
	}
	return c.JSON(http.StatusOK, user)
}

// HandlePostAlumno (versión corregida)
func HandlePostAlumno(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	var req UserRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	// La validación ahora se encarga de comprobar el carrera_id
	if errors := req.Validate(); len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	// Ya no necesitamos buscar el ID, lo recibimos directamente.
	user := Alumno{
		Nombre:  req.Nombre,
		Email:   req.Email,
		Carrera: req.CarreraID,
	}

	_, err := conn.NewInsert().Model(&user).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, user)
}

// HandlePutAlumno y HandleDeleteAlumno (sin cambios, pero incluidos por completitud)
func HandlePutAlumno(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

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
	err = conn.NewSelect().Model(&user).Where("id = ?", id).Scan(ctx.Request().Context())
	if err != nil {
		if err == sql.ErrNoRows {
			return NewApiError(http.StatusNotFound, fmt.Errorf("Usuario no encontrado"))
		}
		return err
	}

	updateQuery := conn.NewUpdate().Model(&user).Where("id = ?", id)

	updated := false
	if req.Nombre != "" {
		updateQuery = updateQuery.Set("nombre = ?", req.Nombre)
		user.Nombre = req.Nombre
		updated = true
	}
	if req.Email != "" {
		updateQuery = updateQuery.Set("email = ?", req.Email)
		user.Email = req.Email
		updated = true
	}
	if req.CarreraID > 0 {
		updateQuery = updateQuery.Set("carrera_id = ?", req.CarreraID)
		user.Carrera = req.CarreraID
		updated = true
	}

	if !updated {
		return InvalidJSON()
	}

	_, err = updateQuery.Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user)
}

func HandleDeleteAlumno(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return err
	}

	res, err := conn.NewDelete().Model((*Alumno)(nil)).Where("id = ?", id).Exec(ctx.Request().Context())
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return NewApiError(http.StatusNotFound, fmt.Errorf("Usuario no encontrado para eliminar"))
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Usuario eliminado exitosamente"})
}

package handlers

import (
	"database/sql"
	"fmt"
	"microser/db"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"name,notnull"`
	Email         string `json:"email" bun:"email,unique"`
}

type CreateUserRequest struct {
	Nombre string `json:"nombre" validate:"required, min=2, max=100"`
	Email  string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
	Nombre string `json:"nombre" validate:"min=2,max=100"`
	Email  string `json:"email" validate:"email"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func HandleGetUsers(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	users := []User{}

	err := conn.NewSelect().Model(&users).Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, []User{})
		}

		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "database_error", Message: "Fallo al retirar usuarios"})
	}

	return c.JSON(http.StatusOK, users)

}

func HandleGetUser(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "ID de usuario inválido",
		})
	}

	var user User
	err = conn.NewSelect().Model(&user).Where("id=?", id).Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "user_not_found",
				Message: "Usuario no encontrado",
			})
		}

		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Fallo al obtener usuario",
		})
	}

	return c.JSON(http.StatusOK, user)
}

func HandlePostUser(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Datos de solicitud inválidos",
		})
	}

	if req.Nombre == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "El nombre es requerido",
		})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: "El email es requerido",
		})
	}

	user := User{
		Nombre: req.Nombre,
		Email:  req.Email,
	}

	_, err := conn.NewInsert().
		Model(&user).
		Exec(ctx)

	if err != nil {
		err_msg := fmt.Sprintf("Fallo al crear usuario: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: err_msg,
		})
	}

	return c.JSON(http.StatusCreated, user)
}

func HandlePutUser(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "ID de usuario inválido",
		})
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Datos de solicitud inválidos",
		})
	}

	var user User
	err = conn.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "user_not_found",
				Message: "Usuario no encontrado",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Fallo al verificar usuario",
		})
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
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "no_fields_to_update",
			Message: "No se proporcionaron campos para actualizar",
		})
	}

	_, err = updateQuery.Exec(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Fallo al actualizar usuario",
		})
	}

	return c.JSON(http.StatusOK, user)

}

func HandleDeleteUser(c echo.Context) error {
	conn := db.GetDB()
	ctx := c.Request().Context()

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_id",
			Message: "ID de usuario inválido",
		})
	}

	var user User
	err = conn.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "user_not_found",
				Message: "Usuario no encontrado",
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Fallo al verificar usuario",
		})
	}

	_, err = conn.NewDelete().
		Model((*User)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "database_error",
			Message: "Fallo al eliminar usuario",
		})
	}

	return c.JSON(http.StatusOK,
		"Usuario eliminado exitosamente",
	)
}

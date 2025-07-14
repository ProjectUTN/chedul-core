package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type User struct {
	Nombre string `json:"nombre"`
}

func HandleGetUsers(c echo.Context) error {

	resp := []User{{
		Nombre: "Lautaro",
	},
	}

	return c.JSON(http.StatusOK, resp)

}

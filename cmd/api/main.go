package main

import (
	"chedul-core/db"
	"chedul-core/handlers"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	_, err := db.InitDB()

	if err != nil {
		fmt.Println("Error al conectar con la DB", err)
	}

	api := echo.New()
	api.HideBanner = true

	api.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"*"},
	}))

	api.Use(middleware.Recover())
	api.Use(handlers.TracingMiddleware("chedul-service"))

	api.GET("/alumnos", handlers.HandleGetAlumnos)
	api.GET("/alumnos/:id", handlers.HandleGetAlumno)
	api.POST("/alumnos", handlers.HandlePostAlumno)
	api.PUT("/alumnos", handlers.HandlePutAlumno)
	api.DELETE("/alumnos/:id", handlers.HandleDeleteAlumno)

	api.Logger.Fatal(api.Start(os.Getenv("LISTEN_ADDR")))
}

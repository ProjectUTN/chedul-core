package main

import (
	"chedul-core/db"
	"chedul-core/handlers"
	"chedul-core/logger"
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	if err := logger.InitLogger(); err != nil {
		fmt.Println("Error iniciando logger:", err)
		os.Exit(1)
	}

	if _, err := db.InitDB(); err != nil {
		logger.GetLogger().Error("Error al conectar con la DB", zap.Error(err))
		os.Exit(1)
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

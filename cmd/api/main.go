package main

import (
	"chedul-core/db"
	"chedul-core/handlers"
	"chedul-core/logger"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func main() {
	logger, err := logger.InitLogger()

	if err != nil {
		log.Fatal("Error iniciando logger:", err)
	}

	pg_pool, err := db.InitDB()

	if err != nil {
		logger.Fatal("Error al conectar con la DB", zap.Error(err))
	}

	api := echo.New()
	api.Use(handlers.AppContextMiddleware(pg_pool, logger))
	api.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"*"},
	}))

	api.Use(middleware.Recover())
	api.Use(handlers.TracingMiddleware("chedul-service"))

	api.HideBanner = true
	api.HTTPErrorHandler = handlers.HttpErrorHandler

	api.GET("/alumnos", handlers.HandleGetAlumnos)
	api.GET("/alumnos/:id", handlers.HandleGetAlumno)
	api.POST("/alumnos", handlers.HandlePostAlumno)
	api.PUT("/alumnos", handlers.HandlePutAlumno)
	api.DELETE("/alumnos/:id", handlers.HandleDeleteAlumno)

	api.GET("/materias", handlers.HandleGetMaterias)
	api.GET("/materias/:id", handlers.HandleGetMateriaByID)
	api.POST("/materias", handlers.HandlePostMateria)

	api.Logger.Fatal(api.Start(os.Getenv("LISTEN_ADDR")))
}

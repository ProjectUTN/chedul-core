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
	api.GET("/alumnos/:id/condiciones", handlers.HandleGetCondicionesPorAlumno)
	api.PUT("/alumnos/:alumno_id/materias/:materia_id", handlers.HandleSetCondicionAlumno) 
	api.GET("/alumnos/:id/progreso", handlers.HandleGetProgresoAlumno)

	api.GET("/materias", handlers.HandleGetMaterias)
	api.GET("/materias/:id", handlers.HandleGetMateriaByID)
	api.GET("/carreras/:id/materias", handlers.HandleGetMateriasPorCarrera)
	api.POST("/materias", handlers.HandlePostMateria)
	api.PUT("/materias/:id", handlers.HandlePutMateria)
	api.DELETE("/materias/:id", handlers.HandleDeleteMateria)

	api.GET("/cuatrimestres", handlers.HandleGetCuatrimestres)
	api.POST("/cuatrimestres", handlers.HandlePostCuatrimestre)
	api.DELETE("/cuatrimestres/:id", handlers.HandleDeleteCuatrimestre)

	api.GET("/carreras", handlers.HandleGetCarreras)
	api.POST("/carreras", handlers.HandlePostCarrera)
	api.DELETE("/carreras/:id", handlers.HandleDeleteCarrera)
	api.GET("/carreras/:id/materias", handlers.HandleGetMateriasPorCarrera)
	api.POST("/carreras/:carrera_id/materias/:materia_id", handlers.HandleAsociarMateriaACarrera)
	api.DELETE("/carreras/:carrera_id/materias/:materia_id", handlers.HandleDesasociarMateriaDeCarrera)

	api.GET("/condiciones", handlers.HandleGetCondiciones)
	api.POST("/condiciones", handlers.HandlePostCondicion)

	api.Logger.Fatal(api.Start(os.Getenv("LISTEN_ADDR")))
}

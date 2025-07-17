package server

import (
	"chedul-core/internals/handlers"
	"chedul-core/internals/repositories"
	"chedul-core/pkg/config"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type Server struct {
	echo   *echo.Echo
	config *config.AppConfig
	db     *bun.DB
	logger *zap.Logger
}

func New(cfg *config.AppConfig, db *bun.DB, logger *zap.Logger) *Server {
	api := echo.New()
	api.HideBanner = true
	api.HTTPErrorHandler = HttpErrorHandler

	return &Server{
		echo:   api,
		config: cfg,
		db:     db,
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.setupMiddleware()
	s.setupRoutes()

	go func() {
		addr := fmt.Sprintf(":%s", s.config.Server.Port)
		s.logger.Info("starting server", zap.String("addr", addr))

		if err := s.echo.Start(addr); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	s.logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Error("failed to shutdown server", zap.Error(err))
		return err
	}

	s.logger.Info("server shutdown complete")
	return nil
}

func (s *Server) setupMiddleware() {
	s.echo.Use(TracingMiddleware(s.logger, "chedul-service"))
	s.echo.Use(CORSMiddleware())
	s.echo.Use(RecoverMiddleware(s.logger))
}

func (s *Server) setupRoutes() {
	s.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"statusCode": http.StatusOK,
			"msg":        "Servicio activo",
		})
	})

	alumnoRepo := repositories.NewAlumnoRepository(s.db)
	carreraRepo := repositories.NewCarreraRepository(s.db)
	materiaRepo := repositories.NewMateriaRepository(s.db)
	condicionRepo := repositories.NewCondicionRepository(s.db)
	condicionAlumnoRepo := repositories.NewCondicionAlumnoRepository(s.db)
	cuatrimestreRepo := repositories.NewCuatrimestreRepository(s.db)

	alumnoHandler := handlers.NewAlumnoHandler(alumnoRepo, carreraRepo, s.logger)
	carreraHandler := handlers.NewCarreraHandler(carreraRepo, s.logger)
	materiaHandler := handlers.NewMateriaHandler(materiaRepo, s.logger)
	condicionHandler := handlers.NewCondicionHandler(condicionRepo, s.logger)
	condicionAlumnoHandler := handlers.NewCondicionAlumnoHandle(condicionAlumnoRepo, s.logger)
	progresoHandler := handlers.NewProgresoHandler(alumnoRepo, condicionAlumnoRepo, condicionRepo, s.db)
	cuatrimestreHandler := handlers.NewCuatrimestreHandler(cuatrimestreRepo, s.logger)

	api := s.echo.Group("/api/v1")

	alumnos := api.Group("/alumnos")
	alumnos.GET("", alumnoHandler.GetAll)
	alumnos.GET("/:id", alumnoHandler.GetByID)
	alumnos.POST("", alumnoHandler.Create)
	alumnos.PUT("/:id", alumnoHandler.Update)
	alumnos.DELETE("/:id", alumnoHandler.Delete)
	alumnos.GET("/progreso/:id", progresoHandler.GetProgresoAlumno)

	carreras := api.Group("/carreras")
	carreras.GET("", carreraHandler.GetAll)
	carreras.GET("/:id", carreraHandler.GetByID)

	materias := api.Group("/materias")
	materias.GET("", materiaHandler.GetAll)
	materias.GET("/:id", materiaHandler.GetByID)

	condicion := api.Group("/condicion")
	condicion.GET("", condicionHandler.GetAll)

	condicion_alumno := api.Group("/condicion_alumno")
	condicion_alumno.GET("/:id", condicionAlumnoHandler.GetCondicionPorAlumno)
	condicion_alumno.POST("", condicionAlumnoHandler.SetCondicionAlumno)

	cuatrimestre := api.Group("/cuatrimestres")
	cuatrimestre.GET("", cuatrimestreHandler.GetAll)
	cuatrimestre.GET("/:id", cuatrimestreHandler.GetByID)
}

package server

import (
	"chedul-core/internals/handlers"
	"chedul-core/internals/repositories"
	"chedul-core/internals/service"
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

	alumnoService := service.NewAlumnoService(alumnoRepo, carreraRepo)
	carreraService := service.NewCarreraService(carreraRepo)
	materiaService := service.NewMateriaService(materiaRepo)

	alumnoHandler := handlers.NewAlumnoHandler(alumnoService, s.logger)
	carreraHandler := handlers.NewCarreraHandler(carreraService, s.logger)
	materiaHandler := handlers.NewMateriaHandler(materiaService, s.logger)

	api := s.echo.Group("/api/v1")

	alumnos := api.Group("/alumnos")
	alumnos.GET("", alumnoHandler.GetAll)
	alumnos.GET("/:id", alumnoHandler.GetByID)
	alumnos.POST("", alumnoHandler.Create)
	alumnos.PUT("/:id", alumnoHandler.Update)
	alumnos.DELETE("/:id", alumnoHandler.Delete)

	carreras := api.Group("/carreras")
	carreras.GET("", carreraHandler.GetAll)
	carreras.GET("/:id", carreraHandler.GetByID)

	materias := api.Group("/materias")
	materias.GET("", materiaHandler.GetAll)
	materias.GET("/:id", materiaHandler.GetByID)
}

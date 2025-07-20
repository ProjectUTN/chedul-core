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

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Green = "\033[32m"
	Cyan  = "\033[36m"
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
	api.HidePort = true
	api.HTTPErrorHandler = HttpErrorHandler

	return &Server{
		echo:   api,
		config: cfg,
		db:     db,
		logger: logger,
	}
}
func (s *Server) printBanner() {

	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		fmt.Print(Cyan)
	}

	fmt.Print(`
  _______          __     __    _____            
 / ___/ /  ___ ___/ /_ __/ /___/ ___/__  _______ 
/ /__/ _ \/ -_) _  / // / /___/ /__/ _ \/ __/ -_)
\___/_//_/\__/\_,_/\_,_/_/    \___/\___/_/  \__/ 

`)

	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		fmt.Print(Reset)
	}
}

func (s *Server) Address() string {
	return fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
}

func (s *Server) Start() error {
	s.setupMiddleware()
	s.setupRoutes()

	go func() {
		s.printBanner()
		s.config.PrettyPrint()

		address := s.Address()

		fmt.Println()
		fmt.Println(Bold + "Server en:" + Reset)
		fmt.Printf("  %sLocal%s:   %shttp://%s%s\n", Bold+Green, Reset, Bold+Cyan, address, Reset)
		fmt.Println()

		if err := s.echo.Start(address); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	s.logger.Info("SIGTERM detectado, cerrando el servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		s.logger.Error("Interrupcion al cerrar el servidor", zap.Error(err))
		return err
	}

	s.logger.Info("Cierre del servidor completado")
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

	alumnoHandler := handlers.NewAlumnoHandler(alumnoRepo, carreraRepo, s.logger, s.config)
	carreraHandler := handlers.NewCarreraHandler(carreraRepo, s.logger)
	materiaHandler := handlers.NewMateriaHandler(materiaRepo, s.logger)
	condicionHandler := handlers.NewCondicionHandler(condicionRepo, s.logger)
	condicionAlumnoHandler := handlers.NewCondicionAlumnoHandle(condicionAlumnoRepo, s.logger)
	progresoHandler := handlers.NewProgresoHandler(alumnoRepo, condicionAlumnoRepo, condicionRepo, s.db)

	api := s.echo.Group("/api/v1")

	api.POST("/signup", alumnoHandler.SignUp)
	api.POST("/login", alumnoHandler.LogIn)
	api.POST("/refresh-token", alumnoHandler.RefreshToken)

	// TODO: agregar proteccion de rutas a aquellas que lo requieran
	protectedAPI := api.Group("")
	protectedAPI.Use(RequireAuthMiddleware(s.config.JwtSecret.Expose(), s.logger))

	alumnos := protectedAPI.Group("/alumnos")
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
}

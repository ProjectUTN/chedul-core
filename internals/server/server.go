package server

import (
	"chedul-core/internals/handlers"
	"chedul-core/internals/repositories"
	"chedul-core/pkg/config"
	"chedul-core/pkg/db"
	"chedul-core/pkg/util"
	"net"

	"chedul-core/pkg/logger"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"

	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"
)

const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Green = "\033[32m"
	Cyan  = "\033[36m"
)

type Server struct {
	Echo      *echo.Echo
	Host      string
	Port      int
	JwtSecret config.Secret
	ConnPool  *bun.DB
	Logger    *zap.Logger
}

func Build(configuration config.AppConfig) *Server {

	logger, err := logger.New(configuration.Server.LogLevel)
	if err != nil {
		log.Fatal("Error iniciando logger:", err)
	}

	connPool, err := db.Open(&configuration)
	if err != nil {
		logger.Fatal("Error al conectar con la DB", zap.Error(err))
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", configuration.Server.Host, configuration.Server.Port))
	if err != nil {
		log.Fatalf("Fallo al unirse al puerto:%d. err: %s ", configuration.Server.Port, err.Error())
	}

	host := configuration.Server.Host
	port := listener.Addr().(*net.TCPAddr).Port

	api := echo.New()
	api.Listener = listener
	api.HideBanner = true
	api.HidePort = true
	api.HTTPErrorHandler = HttpErrorHandler

	server := Server{
		Echo:      api,
		Host:      host,
		JwtSecret: *configuration.JwtSecret,
		Port:      port,
		ConnPool:  connPool,
		Logger:    logger,
	}
	server.SetupMiddleware()
	server.SetupRoutes()

	return &server
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

func (s *Server) Run() error {
	go func() {
		s.printBanner()
		address := fmt.Sprintf("%s:%d", s.Host, s.Port)

		fmt.Println()
		fmt.Printf("%sPerfil:%s %s%s%s\n", Bold, Reset, Bold+Green, config.GetServerEnv(), Reset)
		fmt.Println(Bold + "Server en:" + Reset)
		fmt.Printf("  %sLocal%s:   %shttp://127.0.0.1:%d%s\n", Bold+Green, Reset, Bold+Cyan, s.Port, Reset)
		fmt.Println()

		if err := s.Echo.Start(address); err != nil && err != http.ErrServerClosed {
			s.Logger.Fatal("Fallo al iniciar servidor", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	s.Logger.Info("SIGTERM detectado, cerrando el servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Echo.Shutdown(ctx); err != nil {
		s.Logger.Error("Interrupcion al cerrar el servidor", zap.Error(err))
		return err
	}

	s.Logger.Info("Cierre del servidor completado")
	return nil
}

func (s *Server) SetupMiddleware() {
	if util.IsEnvProd() {
		s.Echo.Use(otelecho.Middleware("chedul-core"))
	}

	s.Echo.Use(TracingMiddleware(s.Logger, "chedul-core"))
	s.Echo.Use(CORSMiddleware())
	s.Echo.Use(RecoverMiddleware(s.Logger))
}

func (s *Server) SetupRoutes() {
	s.Echo.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	alumnoRepo := repositories.NewAlumnoRepository(s.ConnPool)
	carreraRepo := repositories.NewCarreraRepository(s.ConnPool)
	materiaRepo := repositories.NewMateriaRepository(s.ConnPool)
	condicionRepo := repositories.NewCondicionRepository(s.ConnPool)
	condicionAlumnoRepo := repositories.NewCondicionAlumnoRepository(s.ConnPool)

	alumnoHandler := handlers.NewAlumnoHandler(alumnoRepo, carreraRepo, s.Logger, s.JwtSecret)
	carreraHandler := handlers.NewCarreraHandler(carreraRepo, s.Logger)
	materiaHandler := handlers.NewMateriaHandler(materiaRepo, s.Logger)
	condicionHandler := handlers.NewCondicionHandler(condicionRepo, s.Logger)
	condicionAlumnoHandler := handlers.NewCondicionAlumnoHandle(condicionAlumnoRepo, s.Logger)
	progresoHandler := handlers.NewProgresoHandler(alumnoRepo, condicionAlumnoRepo, condicionRepo, s.ConnPool)

	api := s.Echo.Group("/api/v1")

	api.POST("/signup", alumnoHandler.SignUp)
	api.POST("/login", alumnoHandler.LogIn)
	api.POST("/refresh-token", alumnoHandler.RefreshToken)

	protectedAPI := api.Group("")
	protectedAPI.Use(RequireAuthMiddleware(s.JwtSecret.Expose(), s.Logger))

	alumnos := protectedAPI.Group("/alumnos")
	alumnos.GET("", alumnoHandler.GetAll)
	alumnos.GET("/:id", alumnoHandler.GetByID)
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

package server

import (
	"chedul-core/internals/handlers"
	"chedul-core/internals/repositories"
	"chedul-core/pkg/config"
	"chedul-core/pkg/db"
	"chedul-core/pkg/storage"
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
	"github.com/labstack/echo/v4/middleware"
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
	Storage   storage.Storage
	Config    config.AppConfig
}

// Build arma el servidor a partir de la configuracion: conecta la base,
// aplica las migraciones si corresponde y abre el puerto.
func Build(configuration config.AppConfig) *Server {

	logger, err := logger.New(configuration.Server.LogLevel)
	if err != nil {
		log.Fatal("Error iniciando logger:", err)
	}

	connPool, err := db.Open(&configuration)
	if err != nil {
		logger.Fatal("Error al conectar con la DB", zap.Error(err))
	}

	if configuration.Database.AutoMigrate {
		if err := db.Migrate(connPool); err != nil {
			logger.Fatal("Error al aplicar las migraciones", zap.Error(err))
		}
	}

	files, err := storage.NewLocal(configuration.Uploads.Dir)
	if err != nil {
		logger.Fatal("Error al preparar el directorio de archivos", zap.Error(err))
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", configuration.Server.Host, configuration.Server.Port))
	if err != nil {
		log.Fatalf("Fallo al unirse al puerto:%d. err: %s ", configuration.Server.Port, err.Error())
	}

	server := New(configuration, connPool, files, logger, listener)
	server.SetupMiddleware()
	server.SetupRoutes()

	return server
}

// New crea el servidor con dependencias ya construidas. Lo usan Build y los
// tests de integracion.
func New(configuration config.AppConfig, connPool *bun.DB, files storage.Storage, logger *zap.Logger, listener net.Listener) *Server {
	api := echo.New()
	api.Listener = listener
	api.HideBanner = true
	api.HidePort = true
	api.HTTPErrorHandler = HttpErrorHandler

	return &Server{
		Echo:      api,
		Host:      configuration.Server.Host,
		JwtSecret: *configuration.JwtSecret,
		Port:      listener.Addr().(*net.TCPAddr).Port,
		ConnPool:  connPool,
		Logger:    logger,
		Storage:   files,
		Config:    configuration,
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
	if s.Config.TracingEnabled() {
		s.Echo.Use(otelecho.Middleware("chedul-core"))
	}

	s.Echo.Use(TracingMiddleware(s.Logger, "chedul-core"))
	s.Echo.Use(CORSMiddleware(s.Config.Server.CorsOrigins))
	s.Echo.Use(RecoverMiddleware(s.Logger))
	// Un poco mas que el maximo de archivo para dejar lugar al resto del formulario
	s.Echo.Use(middleware.BodyLimit(fmt.Sprintf("%dM", s.Config.Uploads.MaxSizeMB+1)))
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
	aporteRepo := repositories.NewAporteRepository(s.ConnPool)
	calendarioRepo := repositories.NewCalendarioRepository(s.ConnPool)
	estudioRepo := repositories.NewEstudioRepository(s.ConnPool)
	comunidadRepo := repositories.NewComunidadRepository(s.ConnPool)

	alumnoHandler := handlers.NewAlumnoHandler(alumnoRepo, carreraRepo, s.Logger, s.JwtSecret, s.Config.Google.ClientID)
	carreraHandler := handlers.NewCarreraHandler(carreraRepo, s.Logger)
	materiaHandler := handlers.NewMateriaHandler(materiaRepo, s.Logger)
	condicionHandler := handlers.NewCondicionHandler(condicionRepo, s.Logger)
	condicionAlumnoHandler := handlers.NewCondicionAlumnoHandler(condicionAlumnoRepo, condicionRepo, materiaRepo, s.Logger)
	progresoHandler := handlers.NewProgresoHandler(alumnoRepo, materiaRepo, condicionAlumnoRepo)
	aporteHandler := handlers.NewAporteHandler(aporteRepo, materiaRepo, s.Storage, s.Config.Uploads.MaxSizeMB, !s.Config.Uploads.Disabled, s.Logger)
	calendarioHandler := handlers.NewCalendarioHandler(calendarioRepo, materiaRepo, s.Logger)
	estudioHandler := handlers.NewEstudioHandler(estudioRepo, materiaRepo, s.Logger)
	comunidadHandler := handlers.NewComunidadHandler(comunidadRepo, materiaRepo)

	api := s.Echo.Group("/api/v1")

	// Rutas publicas
	api.POST("/signup", alumnoHandler.SignUp)
	api.POST("/login", alumnoHandler.LogIn)
	api.POST("/logout", alumnoHandler.LogOut)
	api.POST("/refresh-token", alumnoHandler.RefreshToken)
	api.GET("/auth/google", alumnoHandler.GoogleConfig)
	api.POST("/auth/google", alumnoHandler.LogInGoogle)

	api.GET("/carreras", carreraHandler.GetAll)
	api.GET("/carreras/:id", carreraHandler.GetByID)

	api.GET("/materias", materiaHandler.GetAll)
	api.GET("/materias/:id", materiaHandler.GetByID)
	api.GET("/materias/:id/comisiones", materiaHandler.GetComisiones)

	api.GET("/condicion", condicionHandler.GetAll)

	// Link de calendario (.ics): lo identifica el token, no la sesion
	api.GET("/calendario/ics/:archivo", calendarioHandler.ExportarICS)

	// Rutas que requieren iniciar sesion
	protectedAPI := api.Group("")
	protectedAPI.Use(RequireAuthMiddleware(s.JwtSecret.Expose(), s.Logger))

	me := protectedAPI.Group("/alumnos/me")
	me.GET("", alumnoHandler.GetMe)
	me.PUT("", alumnoHandler.UpdateMe)
	me.DELETE("", alumnoHandler.DeleteMe)
	me.GET("/progreso", progresoHandler.GetMiProgreso)

	condicionAlumno := protectedAPI.Group("/condicion_alumno")
	condicionAlumno.GET("", condicionAlumnoHandler.GetMisCondiciones)
	condicionAlumno.PUT("/:materia_id", condicionAlumnoHandler.SetCondicion)
	condicionAlumno.DELETE("/:materia_id", condicionAlumnoHandler.DeleteCondicion)

	aportes := protectedAPI.Group("/aportes")
	aportes.GET("", aporteHandler.List)
	aportes.POST("", aporteHandler.Create)
	aportes.GET("/tags", aporteHandler.Tags)
	aportes.GET("/config", aporteHandler.Config)
	aportes.GET("/:id", aporteHandler.Get)
	aportes.PUT("/:id", aporteHandler.Update)
	aportes.DELETE("/:id", aporteHandler.Delete)
	aportes.GET("/:id/archivo", aporteHandler.Descargar)
	aportes.POST("/:id/favorito", aporteHandler.AgregarFavorito)
	aportes.DELETE("/:id/favorito", aporteHandler.QuitarFavorito)

	eventos := protectedAPI.Group("/eventos")
	eventos.GET("", calendarioHandler.ListEventos)
	eventos.POST("", calendarioHandler.CreateEvento)
	eventos.PUT("/:id", calendarioHandler.UpdateEvento)
	eventos.DELETE("/:id", calendarioHandler.DeleteEvento)

	protectedAPI.GET("/calendario-academico", calendarioHandler.ListFechasAcademicas)
	protectedAPI.GET("/calendario/suscripcion", calendarioHandler.GetSuscripcion)
	protectedAPI.POST("/calendario/suscripcion/renovar", calendarioHandler.RenovarSuscripcion)

	estudio := protectedAPI.Group("/estudio")
	estudio.GET("/sesiones", estudioHandler.ListSesiones)
	estudio.POST("/sesiones", estudioHandler.CreateSesion)
	estudio.DELETE("/sesiones/:id", estudioHandler.DeleteSesion)
	estudio.GET("/resumen", estudioHandler.Resumen)
	estudio.GET("/ranking", estudioHandler.Ranking)
	estudio.PUT("/ranking", estudioHandler.SetParticipacion)
	estudio.PUT("/meta", estudioHandler.SetMeta)
	estudio.GET("/tareas", estudioHandler.ListTareas)
	estudio.POST("/tareas", estudioHandler.CreateTarea)
	estudio.PUT("/tareas/:id", estudioHandler.UpdateTarea)
	estudio.DELETE("/tareas/:id", estudioHandler.DeleteTarea)

	comunidades := protectedAPI.Group("/comunidades")
	comunidades.GET("", comunidadHandler.List)
	comunidades.POST("", comunidadHandler.Create)
	comunidades.DELETE("/:id", comunidadHandler.Delete)
	comunidades.POST("/:id/reportar", comunidadHandler.Reportar)

	clases := protectedAPI.Group("/clases")
	clases.GET("", calendarioHandler.ListClases)
	clases.POST("", calendarioHandler.CreateClase)
	clases.PUT("/:id", calendarioHandler.UpdateClase)
	clases.DELETE("/:id", calendarioHandler.DeleteClase)
}

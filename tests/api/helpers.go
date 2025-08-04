package api

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"chedul-core/pkg/logger"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
)

type TestApp struct {
	Address string
	db      *bun.DB
	server  *server.Server
}

func (t *TestApp) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return t.server.Echo.Shutdown(ctx)
}

type spawnOptions struct {
	withDB         bool
	migrationsPath string
}

type SpawnOpts func(*spawnOptions)

func WithDB(migrationsPath string) SpawnOpts {
	return func(so *spawnOptions) {
		so.withDB = true
		so.migrationsPath = migrationsPath
	}
}

func SpawnApp(configPath string, options ...SpawnOpts) (*TestApp, error) {
	opts := &spawnOptions{}
	for _, opt := range options {
		opt(opts)
	}

	var testDB *bun.DB
	if opts.withDB {
		TestingDB, err := SetupPgContainer(opts.migrationsPath)
		if err != nil {
			return nil, fmt.Errorf("No se pudo iniciar base de datos de prueba %w", err)
		}
		testDB = TestingDB.db
	}

	configuration, err := config.Load(configPath)
	if err != nil {
		dir, _ := os.Getwd()
		return nil, fmt.Errorf("No se pudo leer la configuracion: %v", dir)
	}
	configuration.Server.Port = 0

	server := BuildWitDB(configuration, testDB)

	address := server.Echo.Listener.Addr().String()

	startupErr := make(chan error, 1)

	go func() {
		if err := server.Echo.Start(address); err != nil && err != http.ErrServerClosed {
			server.Logger.Fatal("Fallo al iniciar servidor", zap.Error(err))
			startupErr <- err
		}
	}()

	select {
	case err := <-startupErr:
		return nil, fmt.Errorf("Fallo al iniciar servidor %w", err)
	case <-time.After(100 * time.Millisecond):
		if err := waitForServer(fmt.Sprintf("http://%s", address)); err != nil {
			return nil, fmt.Errorf("El servidor no esta listo: %w", err)
		}
	}

	return &TestApp{
		Address: fmt.Sprintf("http://%s", address),
		db:      testDB,
		server:  server,
	}, nil
}

func waitForServer(address string) error {
	client := &http.Client{Timeout: time.Second}

	for range 30 {
		resp, err := client.Get(address + "/health")
		if err == nil {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("El servidor no estuvo listo dentro del plazo")
}

type TestingDB struct {
	container *postgres.PostgresContainer
	db        *bun.DB
}

// Interesante considerar esta tecnica en vez de usar testcontainers si prueban ser un cuello de botella cuando tengamos muchos tests.
// https://gajus.com/blog/setting-up-postgre-sql-for-running-integration-tests
func SetupPgContainer(
	// TODO: Resolver el tema de que se hardcodee
	migrationPath string,
) (*TestingDB, error) {
	ctx := context.Background()
	container, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithAdditionalWaitStrategy(
			wait.
				ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		return nil, err
	}

	connUri, err := container.ConnectionString(ctx, "sslmode=disable")

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connUri)))
	db := bun.NewDB(sqldb, pgdialect.New())

	goose.SetLogger(goose.NopLogger())
	goose.SetDialect("postgres")
	if err := goose.Up(sqldb, migrationPath); err != nil {
		return nil, err
	}

	return &TestingDB{
		container: container,
		db:        db,
	}, nil
}

// TODO: Refactorizar para usar el patron Opts, asi podemos definir si cargar el middleware, si cargar la db, etc.
func BuildWitDB(configuration config.AppConfig, db *bun.DB) *server.Server {
	logger, err := logger.New(configuration.Server.LogLevel)
	if err != nil {
		log.Fatal("Error iniciando logger:", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", configuration.Server.Host, configuration.Server.Port))
	if err != nil {
		log.Fatal("Fallo al unirse a un puerto aleatorio:", err)
	}

	host := configuration.Server.Host
	port := listener.Addr().(*net.TCPAddr).Port

	api := echo.New()
	api.Listener = listener
	api.HideBanner = true
	api.HidePort = true
	api.HTTPErrorHandler = server.HttpErrorHandler

	server := server.Server{
		Echo:      api,
		Host:      host,
		JwtSecret: *configuration.JwtSecret,
		Port:      port,
		ConnPool:  db,
		Logger:    logger,
	}

	// TODO: Add the middleware as an opt-in
	// server.SetupMiddleware()
	server.SetupRoutes()

	return &server
}

func getConfigPath() (string, error) {
	configPath := os.Getenv("CONFIG_DIR")
	if configPath == "" {
		return "", fmt.Errorf("CONFIG_DIR environment variable not set")
	}

	return "../" + configPath, nil
}

func CreateTestApp() (*TestApp, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	testApp, err := SpawnApp(configPath, WithDB("../../migrations"))
	if err != nil {
		return nil, err
	}

	seedDatabase(testApp)

	return testApp, nil
}

func seedDatabase(testApp *TestApp) error {
	_, err := testApp.db.Exec("INSERT INTO carrera(nombre) VALUES ('ISI'), ('Sistemas')")

	if err != nil {
		return err
	}
	return nil
}

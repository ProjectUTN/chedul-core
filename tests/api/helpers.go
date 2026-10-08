package api

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"chedul-core/pkg/logger"
	"chedul-core/pkg/storage"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const MIGRATIONS_DIR = "../../migrations"

type TestApp struct {
	Address string
	db      *bun.DB
	server  *server.Server
}

// Envejecer hace que las cuentas con correo LIKE patron parezcan creadas
// hace 10 dias (las nuevas no cuentan para reportes ni confirmaciones).
func (t *TestApp) Envejecer(patron string) {
	if _, err := t.db.Exec("update alumno set creado = now() - interval '10 days' where email like ?", patron); err != nil {
		panic(err)
	}
}

// HacerAdmin marca la cuenta con ese correo como administradora.
func (t *TestApp) HacerAdmin(email string) {
	if _, err := t.db.Exec("update alumno set es_admin = true where email = ?", email); err != nil {
		panic(err)
	}
}

func (t *TestApp) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return t.server.Echo.Shutdown(ctx)
}

type spawnOptions struct {
	withDB         bool
	migrationsPath string
	config         *config.AppConfig
}

type SpawnOpts func(*spawnOptions)

func WithDB(migrationsPath string) SpawnOpts {
	return func(so *spawnOptions) {
		so.withDB = true
		so.migrationsPath = migrationsPath
	}
}

func WithConfig(cfg *config.AppConfig) SpawnOpts {
	return func(so *spawnOptions) {
		so.config = cfg
	}
}

func createTestConfig() *config.AppConfig {
	return &config.AppConfig{
		Server: config.ServerConfig{
			Host:     "127.0.0.1",
			Port:     0,
			LogLevel: zapcore.WarnLevel,
		},
		Database: config.DatabaseConfig{
			User:       "user",
			Password:   config.NewSecret("password"),
			Name:       "test",
			RequireSsl: false,
		},
		JwtSecret: config.NewSecret("test-jwt-secret-key"),
		TracerProvider: config.HoneycombConfig{
			ServiceName: "chedul-core-test",
			ApiKey:      config.NewSecret(""),
		},
		Uploads: config.UploadsConfig{
			MaxSizeMB: 1,
		},
		Google: config.GoogleConfig{
			ClientID: "test-client.apps.googleusercontent.com",
		},
	}
}

func SpawnApp(options ...SpawnOpts) (*TestApp, error) {
	opts := &spawnOptions{
		config: createTestConfig(),
	}
	for _, opt := range options {
		opt(opts)
	}

	var testDB *bun.DB
	if opts.withDB {
		db, err := SetupTestDB(opts.migrationsPath)
		if err != nil {
			return nil, fmt.Errorf("No se pudo iniciar base de datos de prueba %w", err)
		}
		testDB = db
	}

	configuration := opts.config
	configuration.Server.Port = 0

	uploadsDir, err := os.MkdirTemp("", "chedul-uploads-*")
	if err != nil {
		return nil, err
	}
	configuration.Uploads.Dir = uploadsDir

	srv, err := BuildWithDB(*configuration, testDB)
	if err != nil {
		return nil, err
	}

	address := srv.Echo.Listener.Addr().String()

	go func() {
		if err := srv.Echo.Start(address); err != nil && err != http.ErrServerClosed {
			srv.Logger.Error("Fallo al iniciar servidor", zap.Error(err))
		}
	}()

	if err := waitForServer(fmt.Sprintf("http://%s", address)); err != nil {
		return nil, fmt.Errorf("El servidor no esta listo: %w", err)
	}

	return &TestApp{
		Address: fmt.Sprintf("http://%s", address),
		db:      testDB,
		server:  srv,
	}, nil
}

func waitForServer(address string) error {
	client := &http.Client{Timeout: time.Second}

	for range 60 {
		resp, err := client.Get(address + "/health")
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("El servidor no estuvo listo dentro del plazo")
}

// SetupTestDB crea una base de datos nueva y le aplica las migraciones.
//
// Si la variable TEST_DATABASE_URL apunta a un Postgres (por ejemplo el del
// docker-compose), se crea una base con nombre aleatorio en ese servidor. Si
// no, se levanta un contenedor con testcontainers (necesita Docker).
func SetupTestDB(migrationPath string) (*bun.DB, error) {
	connUri, err := testDatabaseUri()
	if err != nil {
		return nil, err
	}

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(connUri)))
	db := bun.NewDB(sqldb, pgdialect.New())

	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.Up(sqldb, migrationPath); err != nil {
		return nil, err
	}

	return db, nil
}

func testDatabaseUri() (string, error) {
	if base := os.Getenv("TEST_DATABASE_URL"); base != "" {
		return createDatabaseOn(base)
	}

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
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return "", err
	}

	return container.ConnectionString(ctx, "sslmode=disable")
}

func createDatabaseOn(base string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("TEST_DATABASE_URL invalida: %w", err)
	}

	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	name := "chedul_test_" + hex.EncodeToString(buf)

	admin := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(base)))
	defer admin.Close()

	if _, err := admin.Exec(fmt.Sprintf(`CREATE DATABASE %q`, name)); err != nil {
		return "", fmt.Errorf("no se pudo crear la base de prueba: %w", err)
	}

	u.Path = "/" + name
	return u.String(), nil
}

func BuildWithDB(configuration config.AppConfig, db *bun.DB) (*server.Server, error) {
	logger, err := logger.New(configuration.Server.LogLevel)
	if err != nil {
		log.Fatal("Error iniciando logger:", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", configuration.Server.Host, configuration.Server.Port))
	if err != nil {
		return nil, fmt.Errorf("Fallo al unirse a un puerto aleatorio: %w", err)
	}

	files, err := storage.NewLocal(configuration.Uploads.Dir)
	if err != nil {
		return nil, err
	}

	srv := server.New(configuration, db, files, logger, listener)
	srv.Echo.HTTPErrorHandler = server.HttpErrorHandler
	srv.SetupRoutes()

	return srv, nil
}

func CreateTestApp() (*TestApp, error) {
	return SpawnApp(WithDB(MIGRATIONS_DIR))
}

func CreateTestAppWithConfig(cfg *config.AppConfig) (*TestApp, error) {
	return SpawnApp(WithDB(MIGRATIONS_DIR), WithConfig(cfg))
}

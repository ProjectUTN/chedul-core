package api

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"net/http"

	// "github.com/google/uuid"
	"github.com/pressly/goose"
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
	withDB bool
}

type SpawnOpts func(*spawnOptions)

func WithDB() SpawnOpts {
	return func(so *spawnOptions) {
		so.withDB = true
	}
}

func SpawnApp(options ...SpawnOpts) (*TestApp, error) {
	opts := &spawnOptions{}
	for _, opt := range options {
		opt(opts)
	}

	var testDB *bun.DB
	if opts.withDB {
		TestingDB, err := SetupPgContainer("../migrations")
		if err != nil {
			return nil, fmt.Errorf("No se pudo iniciar base de datos de prueba %w", err)
		}
		testDB = TestingDB.db
	}

	configuration, err := config.Load()
	if err != nil {
		dir, _ := os.Getwd()
		return nil, fmt.Errorf("No se pudo leer la configuracion: %v", dir)
	}
	configuration.Server.Port = 0

	server := server.BuildWithoutDB(configuration)
	server.ConnPool = testDB

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

	goose.SetDialect("postgres")

	dir, _ := os.Getwd()
	fmt.Printf("------------------ %v == %v -------------", dir, migrationPath)
	err = goose.Up(sqldb, migrationPath)
	if err != nil {
		return nil, err
	}

	return &TestingDB{
		container: container,
		db:        db,
	}, nil
}

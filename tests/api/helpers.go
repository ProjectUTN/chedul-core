package api

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"context"
	"database/sql"
	"fmt"
	"log"
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
}

func SpawnAppWithDB() TestApp {
	TestingDB, err := SetupContainer()
	if err != nil {
		log.Fatal("No se pudo iniciar base de datos de prueba", err)
	}

	configuration, err := config.Load()
	if err != nil {
		dir, _ := os.Getwd()
		log.Fatal("No se pudo leer la configuracion", dir)
	}
	configuration.Server.Port = 0

	server := server.BuildWithoutDB(configuration)
	server.ConnPool = TestingDB.db

	address := fmt.Sprintf("127.0.0.1:%d", server.Port)

	go func() {
		if err := server.Echo.Start(address); err != nil && err != http.ErrServerClosed {
			server.Logger.Fatal("Fallo al iniciar servidor", zap.Error(err))
		}
	}()

	return TestApp{
		Address: fmt.Sprintf("http://%s", address),
		db:      TestingDB.db,
	}
}

func SpawnAppWithoutDB() TestApp {
	configuration, err := config.Load()
	if err != nil {
		dir, _ := os.Getwd()
		log.Fatal("No se pudo leer la configuracion", dir)
	}
	configuration.Server.Port = 0

	server := server.BuildWithoutDB(configuration)

	address := fmt.Sprintf("127.0.0.1:%d", server.Port)

	go func() {
		if err := server.Echo.Start(address); err != nil && err != http.ErrServerClosed {
			server.Logger.Fatal("Fallo al iniciar servidor", zap.Error(err))
		}
	}()

	return TestApp{
		Address: fmt.Sprintf("http://%s", address),
		db:      nil,
	}
}

type TestingDB struct {
	container *postgres.PostgresContainer
	db        *bun.DB
}

// Interesante considerar esta tecnica en vez de usar testcontainers si prueban ser un cuello de botella cuando tengamos muchos tests.
// https://gajus.com/blog/setting-up-postgre-sql-for-running-integration-tests
func SetupContainer() (*TestingDB, error) {
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
	err = goose.Up(sqldb, "../../migrations")

	if err != nil {
		return nil, err
	}

	return &TestingDB{
		container: container,
		db:        db,
	}, nil
}

package db

import (
	"chedul-core/pkg/config"
	"chedul-core/pkg/util"
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

var once sync.Once

func Open(config *config.AppConfig) (*bun.DB, error) {
	var err error

	var db *bun.DB
	once.Do(func() {
		db, err = createConnection(config)
	})

	if err != nil {
		return nil, err
	}

	if !util.IsEnvProd() {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	return db, err
}

func createConnection(config *config.AppConfig) (*bun.DB, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(config.DatabaseUrl())))

	poolConfig := config.Server.Pool

	configureConnectionPool(sqlDB, poolConfig)

	bunDB := bun.NewDB(sqlDB, pgdialect.New())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := bunDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("error al hacer ping a la base de datos: %w", err)
	}

	return bunDB, nil
}

func configureConnectionPool(sqlDB *sql.DB, cfg *config.ConnectionConfig) {
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime))
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime))
}

func GetDBStats(db *bun.DB) sql.DBStats {
	if db == nil {
		return sql.DBStats{}
	}
	return db.DB.Stats()
}

func PrintPoolStats(db *bun.DB) {
	stats := GetDBStats(db)

	log.Printf("Estadísticas del pool de la base de datos:")
	log.Printf("  Conexiones abiertas: %d", stats.OpenConnections)
	log.Printf("  En uso: %d", stats.InUse)
	log.Printf("  Inactivas: %d", stats.Idle)
	log.Printf("  Veces en espera: %d", stats.WaitCount)
	log.Printf("  Tiempo en espera: %v", stats.WaitDuration)
	log.Printf("  Cerradas por inactividad: %d", stats.MaxIdleClosed)
	log.Printf("  Cerradas por tiempo de vida: %d", stats.MaxLifetimeClosed)
}

func HealthCheck(db *bun.DB, ctx context.Context) error {
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("falló el ping a la base de datos: %w", err)
	}

	stats := GetDBStats(db)
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no hay conexiones disponibles a la base de datos")
	}

	return nil
}

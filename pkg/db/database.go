package db

import (
	"chedul-core/pkg/util"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

var once sync.Once

type ConnectionConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func DefaultConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

func Open(dbUrl string) (*bun.DB, error) {
	var err error

	var db *bun.DB
	once.Do(func() {
		db, err = createConnection(dbUrl)
	})

	if err != nil {
		return nil, err
	}

	if !util.IsEnvProd() {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	return db, err
}

func createConnection(dbUrl string) (*bun.DB, error) {

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbUrl)))

	poolConfig := getConnectionPoolConfig()
	configureConnectionPool(sqlDB, poolConfig)

	bunDB := bun.NewDB(sqlDB, pgdialect.New())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := bunDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("error al hacer ping a la base de datos: %w", err)
	}

	log.Printf("Base de datos conectada con configuración del pool: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%v",
		poolConfig.MaxOpenConns,
		poolConfig.MaxIdleConns,
		poolConfig.ConnMaxLifetime,
	)

	return bunDB, nil
}

func getConnectionPoolConfig() *ConnectionConfig {
	config := DefaultConnectionConfig()

	if maxOpen := os.Getenv("DB_MAX_OPEN_CONNS"); maxOpen != "" {
		if val, err := strconv.Atoi(maxOpen); err == nil {
			config.MaxOpenConns = val
		}
	}

	if maxIdle := os.Getenv("DB_MAX_IDLE_CONNS"); maxIdle != "" {
		if val, err := strconv.Atoi(maxIdle); err == nil {
			config.MaxIdleConns = val
		}
	}

	if maxLifetime := os.Getenv("DB_CONN_MAX_LIFETIME"); maxLifetime != "" {
		if val, err := time.ParseDuration(maxLifetime); err == nil {
			config.ConnMaxLifetime = val
		}
	}

	if maxIdleTime := os.Getenv("DB_CONN_MAX_IDLE_TIME"); maxIdleTime != "" {
		if val, err := time.ParseDuration(maxIdleTime); err == nil {
			config.ConnMaxIdleTime = val
		}
	}

	return config
}

func configureConnectionPool(sqlDB *sql.DB, config *ConnectionConfig) {
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
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
	if db == nil {
		return fmt.Errorf("la base de datos no está inicializada")
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("falló el ping a la base de datos: %w", err)
	}

	stats := GetDBStats(db)
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no hay conexiones disponibles a la base de datos")
	}

	return nil
}

func CloseDB(db *bun.DB) error {
	if db != nil {
		return db.Close()
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

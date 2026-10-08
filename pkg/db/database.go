package db

import (
	"chedul-core/migrations"
	"chedul-core/pkg/config"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

var once sync.Once

func Open(cfg *config.AppConfig) (*bun.DB, error) {
	var err error

	var db *bun.DB
	once.Do(func() {
		db, err = createConnection(cfg)
	})

	if err != nil {
		return nil, err
	}

	if !config.IsProd() {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	return db, err
}

// Migrate aplica todas las migraciones pendientes usando los archivos SQL
// embebidos en el binario.
func Migrate(db *bun.DB) error {
	goose.SetBaseFS(migrations.FS)
	defer goose.SetBaseFS(nil)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	// WithAllowMissing aplica tambien las migraciones con fecha anterior a la
	// ultima aplicada: pasa cuando dos ramas se mergean en otro orden que el
	// de sus fechas, y sin esto la API no arranca.
	if err := goose.Up(db.DB, ".", goose.WithAllowMissing()); err != nil {
		return fmt.Errorf("error aplicando migraciones: %w", err)
	}

	return nil
}

// parametrosNoSoportados son opciones de libpq que pgdriver no entiende y
// manda al servidor como parametros de sesion, que los rechaza. Neon, por
// ejemplo, agrega channel_binding=require a la URL que muestra.
var parametrosNoSoportados = []string{"channel_binding", "gssencmode"}

// limpiarDSN saca de la URL de conexion los parametros que pgdriver no soporta.
func limpiarDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil || u.RawQuery == "" {
		return dsn
	}

	query := u.Query()
	cambio := false
	for _, p := range parametrosNoSoportados {
		if query.Has(p) {
			query.Del(p)
			cambio = true
		}
	}
	if !cambio {
		return dsn
	}

	u.RawQuery = query.Encode()
	return u.String()
}

func createConnection(config *config.AppConfig) (*bun.DB, error) {
	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(limpiarDSN(config.DatabaseUrl()))))

	poolConfig := config.Server.Pool

	configureConnectionPool(sqlDB, poolConfig)

	bunDB := bun.NewDB(sqlDB, pgdialect.New())

	var lastErr error
	retries := 5
	for i := range retries {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := bunDB.PingContext(ctx)
		cancel()

		if err == nil {
			return bunDB, nil
		}
		lastErr = err
		log.Printf("⏳ Intento %d/%d: Esperando por conexion a la DB... (%v)", i+1, retries, err)
		time.Sleep(2 * time.Second)
	}

	sqlDB.Close()
	return nil, fmt.Errorf("error al hacer ping a la base de datos luego de %d: %w", retries, lastErr)
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

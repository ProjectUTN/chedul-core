package main

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"chedul-core/pkg/db"
	"chedul-core/pkg/logger"
	"log"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

type App struct {
	conn   *bun.DB
	logger *zap.Logger
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error leyendo la configuracion")
	}

	logger, err := logger.New(cfg, zap.DebugLevel)
	if err != nil {
		log.Fatal("Error iniciando logger:", err)
	}

	db, err := db.Open(cfg.DatabaseUrl())
	if err != nil {
		logger.Fatal("Error al conectar con la DB", zap.Error(err))
	}
	defer db.Close()

	server := server.New(cfg, db, logger)

	if err := server.Start(); err != nil {
		log.Fatal("Fallo al iniciar el servidor:", err)
	}

}

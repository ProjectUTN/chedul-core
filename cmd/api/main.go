package main

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"chedul-core/pkg/db"
	"chedul-core/pkg/logger"
	"log"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error leyendo la configuracion: ", err)
	}

	logger, err := logger.New(cfg)
	if err != nil {
		log.Fatal("Error iniciando logger:", err)
	}

	db, err := db.Open(cfg)
	if err != nil {
		logger.Fatal("Error al conectar con la DB", zap.Error(err))
	}

	server := server.New(cfg, db, logger)

	if err := server.Start(); err != nil {
		log.Fatal("Fallo al iniciar el servidor:", err)
	}

	defer db.Close()
}

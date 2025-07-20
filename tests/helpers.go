package tests

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"fmt"
	"log"
	"os"

	"net/http"

	// "github.com/google/uuid"
	"go.uber.org/zap"
)

type TestApp struct {
	address string
}

func SpawnApp() TestApp {
	configuration, err := config.Load()
	if err != nil {
		dir, _ := os.Getwd()
		log.Fatal("No se pudo leer la configuracion", dir)
	}
	configuration.Server.Port = 0

	server := server.Build(configuration)
	address := fmt.Sprintf("127.0.0.1:%d", server.Port)

	go func() {
		if err := server.Echo.Start(address); err != nil && err != http.ErrServerClosed {
			server.Logger.Fatal("Fallo al iniciar servidor", zap.Error(err))
		}
	}()

	return TestApp{
		address: fmt.Sprintf("http://%s", address),
	}
}

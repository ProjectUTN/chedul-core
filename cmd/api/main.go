package main

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"log"
	"os"
)

func main() {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		log.Fatal("'CONFIG_DIR' no esta configurado")
	}

	configuration, err := config.Load(configDir)
	if err != nil {
		log.Fatal("No se pudo leer la configuracion")
	}
	configuration.PrettyPrint()

	server := server.Build(configuration)

	if err := server.Run(); err != nil {
		log.Fatal("Fallo al iniciar el servidor:", err)
	}
}

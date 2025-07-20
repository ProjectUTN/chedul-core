package main

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"
	"log"
)

func main() {
	configuration, err := config.Load()
	if err != nil {
		log.Fatal("No se pudo leer la configuracion")
	}
	configuration.PrettyPrint()

	server := server.Build(configuration)

	if err := server.Run(); err != nil {
		log.Fatal("Fallo al iniciar el servidor:", err)
	}
}

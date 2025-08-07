package main

import (
	"chedul-core/internals/server"
	"chedul-core/pkg/config"

	"log"
	"os"

	otel "github.com/honeycombio/otel-config-go/otelconfig"
)

func main() {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		log.Fatal("'CONFIG_DIR' no esta configurado")
	}

	config, err := config.Load(configDir)
	if err != nil {
		log.Fatal("No se pudo leer la configuracion")
	}
	config.PrettyPrint()

	otelShutdown, err := otel.ConfigureOpenTelemetry(
		otel.WithServiceName(config.TracerProvider.ServiceName),
		otel.WithExporterProtocol(otel.Protocol(config.TracerProvider.Protocol)),
		otel.WithExporterEndpoint(config.TracerProvider.Endpoint),
		otel.WithHeaders(map[string]string{
			"x-honeycomb-team": config.TracerProvider.ApiKey.Expose(),
		}),
	)
	if err != nil {
		log.Printf("error setting up OTel SDK - %e\n", err)
	}
	defer otelShutdown()

	server := server.Build(config)

	if err := server.Run(); err != nil {
		log.Println("Fallo al iniciar el servidor:", err)
	}
}

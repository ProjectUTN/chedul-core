package config

import (
	"fmt"

	"github.com/spf13/viper"
)

var allowedEnvs = map[string]bool{
	"development": true,
	"production":  true,
	"test":        true,
}

type AppConfig struct {
	Env    string
	Server struct {
		Port string
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
}

func Load() (*AppConfig, error) {
	viper.SetConfigFile("config.yml")

	var config AppConfig
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error leyendo el archivo de configuracion: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("No se pudo serializar la configuracion: %w", err)
	}

	if !allowedEnvs[config.Env] {
		return nil, fmt.Errorf("La variable 'env' no es aceptada: %v", config.Env)
	}


	return &config, nil

}

func (self *AppConfig) DatabaseUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", self.Database.User, self.Database.Password, self.Database.Host, self.Database.Port, self.Database.Name)
}

func (self *AppConfig) IsProd() bool {
	return self.Env == "production"
}

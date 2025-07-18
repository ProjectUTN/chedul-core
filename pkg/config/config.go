package config

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

var allowedEnvs = map[string]bool{
	"development": true,
	"production":  true,
	"test":        true,
}

const (
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
	colorReset  = "\033[0m"
)

type AppConfig struct {
	Env       string         `json:"env"`
	Server    ServerConfig   `json:"server"`
	Database  DatabaseConfig `json:"database"`
	JwtSecret *Secret        `json:"jwt_secret"`
}

type ServerConfig struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	LogLevel zapcore.Level     `json:"log_level"`
	Pool     *ConnectionConfig `json:"pool"`
}

type DatabaseConfig struct {
	Host     string  `json:"host"`
	Port     int     `json:"port"`
	User     string  `json:"user"`
	Password *Secret `json:"password"`
	Name     string  `json:"name"`
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

type ConnectionConfig struct {
	MaxOpenConns    int      `json:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns    int      `json:"max_idle_conns" mapstructure:"max_idle_conns"`
	ConnMaxLifetime Duration `json:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime Duration `json:"conn_max_idletime" mapstructure:"conn_max_idletime"`
}

func DefaultConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: Duration(30 * time.Minute),
		ConnMaxIdleTime: Duration(5 * time.Minute),
	}
}

func (c *AppConfig) PrettyPrint() {
	var obj map[string]any
	tmp, _ := json.Marshal(c)
	json.Unmarshal(tmp, &obj)

	f := colorjson.NewFormatter()
	f.Indent = 2
	s, _ := f.Marshal(obj)
	fmt.Println(string(s))
}

func (c *AppConfig) validate() error {
	var errs []error

	if !allowedEnvs[c.Env] {
		errs = append(errs, fmt.Errorf("env invalido %q: debe ser development, test o production", c.Env))
	}

	if net.ParseIP(c.Server.Host) == nil {
		errs = append(errs, fmt.Errorf("server.host invalido %q: debe ser una IP valida", c.Server.Host))
	}

	if net.ParseIP(c.Database.Host) == nil {
		errs = append(errs, fmt.Errorf("database.host invalido %q: debe ser una IP valida", c.Database.Host))
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port invalido %d: debe estar entre 1 y 65535", c.Server.Port))
	}

	if c.Database.Port < 1 || c.Database.Port > 65535 {
		errs = append(errs, fmt.Errorf("database.port invalido %d: debe estar entre 1 y 65535", c.Database.Port))
	}

	if len(errs) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s%s❌ Errores de configuración:%s\n", colorRed, colorBold, colorReset))
		for _, err := range errs {
			sb.WriteString(fmt.Sprintf("  %s•%s %s\n", colorYellow, colorReset, err.Error()))
		}
		return fmt.Errorf("%s", sb.String())
	}

	return nil
}

func Load() (*AppConfig, error) {
	viper.SetConfigFile("config.yml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("Error leyendo el archivo de configuracion: %w", err)
	}

	var config AppConfig
	if err := viper.Unmarshal(&config, func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			logLevelDecodeHook(),
			durationDecodeHook(),
			SecretDecodeHook(),
		)
	}); err != nil {
		return nil, fmt.Errorf("No se pudo decodificar la configuracion: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	if config.Server.Pool == nil {
		config.Server.Pool = DefaultConnectionConfig()
	}

	return &config, nil
}

func (self *AppConfig) DatabaseUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		self.Database.User,
		self.Database.Password.Expose(),
		self.Database.Host,
		self.Database.Port,
		self.Database.Name,
	)
}

func (self *AppConfig) IsProd() bool {
	return self.Env == "production"
}

func logLevelDecodeHook() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, value any) (any, error) {
		if from.Kind() == reflect.String && to == reflect.TypeOf(zapcore.Level(0)) {
			str := value.(string)
			level, err := zapcore.ParseLevel(str)
			if err != nil {
				return nil, fmt.Errorf("logLevel invalido %q: %w", str, err)
			}
			return level, nil
		}

		return value, nil
	}
}
func durationDecodeHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(Duration(0)) {
			return data, nil
		}

		var durationStr string
		switch v := data.(type) {
		case string:
			durationStr = v
		default:
			return data, nil
		}

		dur, err := time.ParseDuration(durationStr)
		if err != nil {
			return nil, err
		}

		return Duration(dur), nil
	}
}

package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
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
	Server         ServerConfig    `json:"server"`
	Database       DatabaseConfig  `json:"database"`
	JwtSecret      *Secret         `json:"jwt_secret" mapstructure:"jwt_secret"`
	TracerProvider HoneycombConfig `json:"tracer_provider" mapstructure:"tracer_provider"`
	Uploads        UploadsConfig   `json:"uploads"`
	Google         GoogleConfig    `json:"google"`
	// Correos de los administradores (variable ADMIN_EMAILS, separados por
	// coma). Van por entorno para que no queden en el repo.
	Admins Admins `json:"admins"`
}

type Admins []string

// Incluye dice si el correo es de un administrador (sin importar mayusculas).
func (a Admins) Incluye(email string) bool {
	for _, admin := range a {
		if strings.EqualFold(strings.TrimSpace(admin), strings.TrimSpace(email)) {
			return true
		}
	}
	return false
}

// GoogleConfig habilita "Continuar con Google". El client ID es publico (va en
// el front tambien); sin el, el login con Google queda apagado.
type GoogleConfig struct {
	ClientID string `json:"client_id" mapstructure:"client_id"`
}

type ServerConfig struct {
	Host        string            `json:"host"`
	Port        int               `json:"port"`
	LogLevel    zapcore.Level     `json:"log_level"`
	Pool        *ConnectionConfig `json:"pool"`
	CorsOrigins []string          `json:"cors_origins" mapstructure:"cors_origins"`
}

type UploadsConfig struct {
	Dir       string `json:"dir"`
	MaxSizeMB int64  `json:"max_size_mb" mapstructure:"max_size_mb"`
	// Disabled apaga la subida de archivos y los aportes pasan a ser solo
	// links. Sirve en hostings sin disco persistente, como Cloud Run.
	Disabled bool `json:"disabled"`
}

type HoneycombConfig struct {
	ServiceName string  `json:"service_name" mapstructure:"service_name"`
	Protocol    string  `json:"protocol"`
	Endpoint    string  `json:"endpoint"`
	ApiKey      *Secret `json:"api_key" mapstructure:"api_key"`
}

type DatabaseConfig struct {
	Host       string  `json:"host"`
	Port       int     `json:"port"`
	User       string  `json:"user"`
	Password   *Secret `json:"password"`
	Name       string  `json:"name"`
	RequireSsl bool    `json:"require_ssl" mapstructure:"require_ssl"`
	// Url, si esta definida (variable DATABASE_URL), reemplaza al resto de los campos.
	// Es lo que entregan los Postgres administrados (Neon, Supabase, Railway, etc).
	Url         *Secret `json:"url"`
	AutoMigrate bool    `json:"auto_migrate" mapstructure:"auto_migrate"`
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

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Host:     "0.0.0.0",
			Port:     8080,
			LogLevel: zapcore.InfoLevel,
			Pool:     DefaultConnectionConfig(),
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: &Secret{value: "password"},
			Name:     "appdb",
		},
		JwtSecret: &Secret{value: "changeme"},
	}
}

func (c *AppConfig) PrettyPrint() {
	var obj map[string]any
	tmp, _ := json.Marshal(c)
	json.Unmarshal(tmp, &obj)

	f := colorjson.NewFormatter()
	f.Indent = 4
	s, _ := f.Marshal(obj)

	fmt.Println(string(s))
}

func (c *AppConfig) overrideWithEnv() *AppConfig {
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		c.Database.Password = NewSecret(dbPassword)
	}

	if dbUrl := os.Getenv("DATABASE_URL"); dbUrl != "" {
		c.Database.Url = NewSecret(dbUrl)
	}

	if apiHost := os.Getenv("API_HOST"); apiHost != "" {
		c.Server.Host = apiHost
	}

	// La mayoria de los hostings (Cloud Run, Render, Railway) indican el puerto con PORT
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Server.Port = p
		}
	}

	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		c.Server.CorsOrigins = nil
		for _, o := range strings.Split(origins, ",") {
			if o = strings.TrimSpace(o); o != "" {
				c.Server.CorsOrigins = append(c.Server.CorsOrigins, o)
			}
		}
	}

	if dir := os.Getenv("UPLOADS_DIR"); dir != "" {
		c.Uploads.Dir = dir
	}

	if v := os.Getenv("UPLOADS_DISABLED"); v != "" {
		c.Uploads.Disabled = v == "true" || v == "1"
	}

	if id := os.Getenv("GOOGLE_CLIENT_ID"); id != "" {
		c.Google.ClientID = id
	}

	if admins := os.Getenv("ADMIN_EMAILS"); admins != "" {
		c.Admins = nil
		for _, email := range strings.Split(admins, ",") {
			if email = strings.TrimSpace(email); email != "" {
				c.Admins = append(c.Admins, email)
			}
		}
	}

	return c
}

func (c *AppConfig) validate() error {
	var errs []error

	env := GetServerEnv()
	if !allowedEnvs[env] {
		errs = append(errs, fmt.Errorf("env invalido %q: debe ser development, test o production", env))
	}

	if net.ParseIP(c.Server.Host) == nil {
		errs = append(errs, fmt.Errorf("server.host invalido %q: debe ser una IP valida", c.Server.Host))
	}

	if c.Database.Url.Expose() == "" && net.ParseIP(c.Database.Host) == nil && os.Getenv("ALLOW_DB_ALIAS") != "true" {
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

func Load(configDir string) (AppConfig, error) {

	viper.SetConfigFile(configDir + "base.yml")

	if err := viper.ReadInConfig(); err != nil {
		return AppConfig{}, fmt.Errorf("Error leyendo el archivo de configuracion: %w", err)
	}

	env := GetServerEnv()
	envConfig := configDir + fmt.Sprintf("%s.yml", env)
	if _, err := os.Stat(envConfig); err == nil {
		viper.SetConfigFile(envConfig)
		if err := viper.MergeInConfig(); err != nil {
			return AppConfig{}, fmt.Errorf("Error leyendo el archivo de configuracion para %s: %w", env, err)
		}
	}

	var config AppConfig
	if err := viper.Unmarshal(&config, func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			logLevelDecodeHook(),
			durationDecodeHook(),
			SecretDecodeHook(),
		)
	}); err != nil {
		return AppConfig{}, fmt.Errorf("No se pudo decodificar la configuracion: %w", err)
	}

	if err := config.overrideWithEnv().validate(); err != nil {
		return AppConfig{}, err
	}

	if config.Server.Pool == nil {
		config.Server.Pool = DefaultConnectionConfig()
	}

	if config.Uploads.Dir == "" {
		config.Uploads.Dir = "uploads"
	}
	if config.Uploads.MaxSizeMB <= 0 {
		config.Uploads.MaxSizeMB = 25
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return AppConfig{}, fmt.Errorf("'JWT_SECRET' no puede ser leida correctamente")
	}
	if IsProd() && len(jwtSecret) < 32 {
		return AppConfig{}, fmt.Errorf("'JWT_SECRET' debe tener al menos 32 caracteres en produccion")
	}

	config.JwtSecret = NewSecret(jwtSecret)
	// Honeycomb es opcional: sin api key no se exportan trazas
	config.TracerProvider.ApiKey = NewSecret(os.Getenv("HONEYCOMB_API_KEY"))

	return config, nil
}

func (self *AppConfig) TracingEnabled() bool {
	return self.TracerProvider.ApiKey.Expose() != ""
}

func (self *AppConfig) DatabaseUrl() string {
	if url := self.Database.Url.Expose(); url != "" {
		return url
	}

	sslmode := "disable"
	if self.Database.RequireSsl {
		sslmode = "require"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		self.Database.User,
		self.Database.Password.Expose(),
		self.Database.Host,
		self.Database.Port,
		self.Database.Name,
		sslmode,
	)
}

func IsProd() bool {
	return GetServerEnv() == "production"
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

func GetServerEnv() string {
	env := os.Getenv("SERVER_ENV")
	if env == "" {
		env = "development"
	}

	return env
}

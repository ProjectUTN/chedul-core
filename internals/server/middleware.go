package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func TracingMiddleware(logger *zap.Logger, serviceName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			requestLogger := logger.With(
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.String("remote_addr", c.RealIP()),
			)

			c.Set("Logger", requestLogger)
			requestLogger.Info("Request started")

			err := next(c)
			duration := time.Since(start)

			status := c.Response().Status
			if err != nil {
				if httpErr, ok := err.(*echo.HTTPError); ok {
					status = httpErr.Code
				}
			}

			logFields := []zap.Field{
				zap.Int("status", status),
				zap.Duration("duration", duration),
				zap.Int64("response_size", c.Response().Size),
			}

			if err != nil {
				logFields = append(logFields, zap.Error(err))
				requestLogger.Error("Request failed", logFields...)
			} else {
				requestLogger.Info("Request completed", logFields...)
			}

			return err

		}
	}
}

func CORSMiddleware() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"*"},
		AllowMethods: []string{"*"},
	})
}

func RecoverMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Error("panic recovered", zap.Error(err), zap.ByteString("stack", stack))
			return nil
		},
	})
}

type CustomClaims struct {
	Sub int64 `json:"sub"` // El ID del alumno
	jwt.RegisteredClaims
}

func RequireAuthMiddleware(secretKey string, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var tokenString string
			var tokenSource string

			// 1. Intentar obtener el token del encabezado Authorization
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				tokenSource = "header"
			}

			// 2. Si no se encontró en el encabezado, intentar obtenerlo de la cookie
			if tokenString == "" {
				cookie, err := c.Cookie("Authorization")
				if err == nil && cookie.Value != "" {
					tokenString = cookie.Value
					tokenSource = "cookie"
				}
			}

			// Si el token sigue vacío después de buscar en header y cookie
			if tokenString == "" {
				logger.Warn("RequireAuthMiddleware: Token de autenticación faltante en encabezado o cookie",
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
				)
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token de autenticación faltante"})
			}

			// 3. Parsear y validar el token
			token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {

				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					logger.Error("RequireAuthMiddleware: Método de firma inesperado",
						zap.Any("alg", token.Header["alg"]),
						zap.String("path", c.Request().URL.Path),
						zap.String("method", c.Request().Method),
						zap.String("remote_ip", c.RealIP()),
						zap.String("token_source", tokenSource),
					)
					return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			})

			if err != nil {
				logger.Error("RequireAuthMiddleware: Error al parsear o validar token",
					zap.Error(err),
					zap.String("token_string", tokenString),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
					zap.String("token_source", tokenSource),
				)
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token inválido o expirado"})
			}

			// 4. Verificar si el token es válido y extraer los claims
			if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
				// 5. Establecer el alumnoID en el contexto de Echo
				c.Set("alumnoID", claims.Sub)
				logger.Info("RequireAuthMiddleware: Autenticación exitosa",
					zap.Int64("alumnoID", claims.Sub),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
					zap.String("token_source", tokenSource),
				)

				// 6. Continuar con el siguiente manejador en la cadena
				return next(c)
			} else {
				logger.Warn("RequireAuthMiddleware: Token no válido o claims incorrectos",
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
					zap.Bool("token_valid", token.Valid),
					zap.String("token_source", tokenSource),
				)
				return c.JSON(http.StatusForbidden, map[string]string{"message": "Acceso prohibido. Token no válido."})
			}
		}
	}
}

// GetAlumnoIDFromContext es una función de utilidad para obtener el alumnoID
// del contexto de Echo en tus manejadores.
func GetAlumnoIDFromContext(c echo.Context) (int64, error) {
	alumnoID, ok := c.Get("alumnoID").(int64)
	if !ok {
		return 0, fmt.Errorf("alumnoID no encontrado en el contexto o tipo incorrecto")
	}
	return alumnoID, nil
}

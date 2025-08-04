package server

import (
	"chedul-core/internals/handlers"
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
				} else if apiErr, ok := err.(handlers.ApiError); ok {
					status = apiErr.StatusCode
				} else {
					status = http.StatusInternalServerError
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
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
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
	Sub int64 `json:"sub"`
	jwt.RegisteredClaims
}

func RequireAuthMiddleware(secretKey string, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var accessTokenString string
			var tokenSource string

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				accessTokenString = strings.TrimPrefix(authHeader, "Bearer ")
				tokenSource = "header"
			} else {
				logger.Warn("RequireAuthMiddleware: Access Token faltante en encabezado Authorization",
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
				)
				return handlers.NewApiError(http.StatusUnauthorized, fmt.Errorf("Access Token faltante en encabezado Authorization"))
			}

			token, err := jwt.ParseWithClaims(accessTokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					logger.Error("RequireAuthMiddleware: Método de firma inesperado para access token",
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
				logger.Error("RequireAuthMiddleware: Error al parsear o validar access token",
					zap.Error(err),
					zap.String("token_string", accessTokenString),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
					zap.String("token_source", tokenSource),
				)

				// si el token expiro hay que enviar la peticion desde el front para generar otro
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
			}

			if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
				c.Set("alumnoID", claims.Sub)
				logger.Info("RequireAuthMiddleware: Autenticación exitosa con Access Token",
					zap.Int64("alumnoID", claims.Sub),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
					zap.String("token_source", tokenSource),
				)

				return next(c)
			} else {
				logger.Warn("RequireAuthMiddleware: Access Token no válido o claims incorrectos",
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
					zap.Bool("token_valid", token.Valid),
					zap.String("token_source", tokenSource),
				)
				return handlers.NewApiError(http.StatusForbidden, fmt.Errorf("Acceso prohibido. Access Token no válido."))
			}
		}
	}
}

func GetAlumnoIDFromContext(c echo.Context) (int64, error) {
	alumnoID, ok := c.Get("alumnoID").(int64)
	if !ok {
		return 0, fmt.Errorf("alumnoID no encontrado en el contexto o tipo incorrecto")
	}
	return alumnoID, nil
}

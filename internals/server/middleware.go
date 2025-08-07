package server

import (
	"chedul-core/internals/handlers"
	"chedul-core/pkg/util"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("chedul-core")

func TracingMiddleware(logger *zap.Logger, serviceName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, span := tracer.Start(c.Request().Context(), "tracing_middleware")
			defer span.End()

			start := time.Now()

			requestLogger := logger.With(
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.String("remote_addr", c.RealIP()),
			)

			if spanCtx := span.SpanContext(); spanCtx.IsValid() {
				requestLogger = requestLogger.With(
					zap.String("trace_id", spanCtx.TraceID().String()),
					zap.String("span_id", spanCtx.SpanID().String()),
				)
			}
			span.SetAttributes(
				attribute.String("http.method", c.Request().Method),
				attribute.String("http.url", c.Request().URL.String()),
				attribute.String("http.remote_addr", c.RealIP()),
			)

			c.Set("Logger", requestLogger)
			c.Set("tracer", tracer)
			c.SetRequest(c.Request().WithContext(ctx))

			requestLogger.Info("Request started")

			err := next(c)
			duration := time.Since(start)

			status := c.Response().Status
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				if httpErr, ok := err.(*echo.HTTPError); ok {
					status = httpErr.Code
				} else if apiErr, ok := err.(handlers.ApiError); ok {
					status = apiErr.StatusCode
				} else {
					status = http.StatusInternalServerError
				}
			} else {
				span.SetStatus(codes.Ok, "")
			}

			span.SetAttributes(
				attribute.Int("http.status_code", status),
				attribute.Int64("http.response_size", c.Response().Size),
				attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1000000),
			)

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
			ctx, span := tracer.Start(c.Request().Context(), "auth_middleware")
			defer span.End()

			if !util.IsEnvProd() {
				c.Set("alumnoID", int64(1))

				span.SetStatus(codes.Ok, "Modo desarrollo - autenticación automática")
				span.SetAttributes(
					attribute.String("auth.mode", "development"),
					attribute.Int64("auth.user_id", 1),
					attribute.String("auth.status", "auto_success"),
				)

				logger.Info("RequireAuthMiddleware: Modo desarrollo - autenticación automática",
					zap.Int64("alumnoID", 1),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
					zap.String("mode", "development"),
				)

				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			}

			var accessTokenString string
			var tokenSource string

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				accessTokenString = strings.TrimPrefix(authHeader, "Bearer ")
				tokenSource = "header"
			} else {
				span.SetStatus(codes.Error, "missing authorization header")
				span.SetAttributes(attribute.String("auth.error", "missing_token"))

				logger.Warn("RequireAuthMiddleware: Access Token faltante en encabezado Authorization",
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
				)
				return handlers.NewApiError(http.StatusUnauthorized, fmt.Errorf("Access Token faltante en encabezado Authorization"))
			}

			span.SetAttributes(
				attribute.String("auth.token_source", tokenSource),
				attribute.String("auth.method", "jwt"),
			)

			token, err := jwt.ParseWithClaims(accessTokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					span.SetAttributes(attribute.String("auth.error", "unexpected_signing_method"))
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
				span.RecordError(err)
				span.SetStatus(codes.Error, "token validation failed")
				span.SetAttributes(attribute.String("auth.error", "token_validation_failed"))

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

				span.SetStatus(codes.Ok, "authentication successful")
				span.SetAttributes(
					attribute.Int64("auth.user_id", claims.Sub),
					attribute.String("auth.status", "success"),
				)

				logger.Info("RequireAuthMiddleware: Autenticación exitosa con Access Token",
					zap.Int64("alumnoID", claims.Sub),
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("remote_ip", c.RealIP()),
					zap.String("token_source", tokenSource),
				)

				c.SetRequest(c.Request().WithContext(ctx))

				return next(c)
			} else {
				span.SetStatus(codes.Error, "invalid token or claims")
				span.SetAttributes(
					attribute.String("auth.error", "invalid_token_or_claims"),
					attribute.Bool("token.valid", token.Valid),
				)

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

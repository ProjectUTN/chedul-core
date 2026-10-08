package server

import (
	"chedul-core/internals/handlers"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

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

				var httpErr *echo.HTTPError
				var apiErr handlers.ApiError
				switch {
				case errors.As(err, &httpErr):
					status = httpErr.Code
				case errors.As(err, &apiErr):
					status = apiErr.StatusCode
				case errors.Is(err, sql.ErrNoRows):
					status = http.StatusNotFound
				default:
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

func CORSMiddleware(origins []string) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		ExposeHeaders:    []string{echo.HeaderContentDisposition},
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

// RequireAuthMiddleware exige un access token valido en el encabezado
// Authorization y guarda el id del alumno en el contexto. No hay excepciones
// por entorno: en desarrollo tambien hay que iniciar sesion.
func RequireAuthMiddleware(secretKey string, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, span := tracer.Start(c.Request().Context(), "auth_middleware")
			defer span.End()

			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if !strings.HasPrefix(authHeader, "Bearer ") {
				span.SetStatus(codes.Error, "missing authorization header")
				span.SetAttributes(attribute.String("auth.error", "missing_token"))
				return handlers.NewApiError(http.StatusUnauthorized, fmt.Errorf("Access Token faltante en encabezado Authorization"))
			}

			claims, err := handlers.ParseAccessToken(strings.TrimPrefix(authHeader, "Bearer "), secretKey)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "token validation failed")
				span.SetAttributes(attribute.String("auth.error", "token_validation_failed"))

				logger.Debug("RequireAuthMiddleware: access token rechazado",
					zap.Error(err),
					zap.String("path", c.Request().URL.Path),
				)

				// El frontend usa este mensaje para pedir un nuevo token con /refresh-token
				return handlers.NewApiError(http.StatusUnauthorized, fmt.Errorf(handlers.AccessTokenExpiradoMsg))
			}

			c.Set(handlers.ContextAlumnoID, claims.Sub)

			span.SetStatus(codes.Ok, "authentication successful")
			span.SetAttributes(
				attribute.Int64("auth.user_id", claims.Sub),
				attribute.String("auth.status", "success"),
			)

			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

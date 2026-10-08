package server

import (
	"mime"
	"net"
	"net/http"
	"strings"
	"time"

	"chedul-core/internals/handlers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// ipCliente devuelve la IP de quien hace el pedido. En produccion los pedidos
// llegan por el proxy de Vercel, que pone la IP real como primera entrada de
// X-Forwarded-For (Cloud Run agrega la de Vercel al final).
func ipCliente(c echo.Context) string {
	if xff := c.Request().Header.Get(echo.HeaderXForwardedFor); xff != "" {
		if ip := strings.TrimSpace(strings.Split(xff, ",")[0]); ip != "" {
			return ip
		}
	}
	return c.RealIP()
}

// local es un pedido hecho desde la misma maquina sin pasar por un proxy
// (tests y desarrollo). En Cloud Run la conexion siempre viene del balanceador
// de Google, nunca de loopback, asi que esto no se puede falsificar.
func local(c echo.Context) bool {
	if c.Request().Header.Get(echo.HeaderXForwardedFor) != "" {
		return false
	}
	host, _, err := net.SplitHostPort(c.Request().RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// LimitePorIP corta a quien manda demasiados pedidos de registro o ingreso:
// 20 de entrada y despues uno cada 6 segundos.
func LimitePorIP() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: local,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(10.0 / 60),
			Burst:     20,
			ExpiresIn: 15 * time.Minute,
		}),
		IdentifierExtractor: func(c echo.Context) (string, error) { return ipCliente(c), nil },
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return handlers.NewApiError(http.StatusTooManyRequests, errDemasiadosPedidos)
		},
	})
}

type errorTexto string

func (e errorTexto) Error() string { return string(e) }

const errDemasiadosPedidos = errorTexto("Demasiados intentos. Esperá unos minutos y probá de nuevo.")

// SoloJSON rechaza lo que no venga como JSON. Un formulario HTML de otro sitio
// no puede mandar application/json sin pasar por CORS, asi que esto evita que
// una pagina ajena inicie sesion en Chedul con la cuenta de otro.
func SoloJSON() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tipo, _, _ := mime.ParseMediaType(c.Request().Header.Get(echo.HeaderContentType))
			if tipo != echo.MIMEApplicationJSON {
				return handlers.NewApiError(http.StatusUnsupportedMediaType, errorTexto("El pedido tiene que ser JSON"))
			}
			return next(c)
		}
	}
}

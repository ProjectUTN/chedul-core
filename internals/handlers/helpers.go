package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type ApiError struct {
	StatusCode int `json:"statusCode"`
	Msg        any `json:"msg"`
}

func (e ApiError) Error() string {
	return fmt.Sprintf("api error: %d", e.StatusCode)
}

func NewApiError(statusCode int, err error) ApiError {
	return ApiError{
		StatusCode: statusCode,
		Msg:        err.Error(),
	}
}

func InvalidRequestData(errors map[string]string) ApiError {
	return ApiError{
		StatusCode: http.StatusUnprocessableEntity,
		Msg:        errors,
	}
}

func InvalidJSON() ApiError {
	return NewApiError(http.StatusBadRequest, fmt.Errorf("invalid JSON request data"))
}

func NotFound(recurso string) ApiError {
	return NewApiError(http.StatusNotFound, fmt.Errorf("%s no encontrado", recurso))
}

func Forbidden() ApiError {
	return NewApiError(http.StatusForbidden, fmt.Errorf("No tenés permiso para hacer esto"))
}

// Mensaje que el frontend usa para saber que tiene que pedir un nuevo access token.
const AccessTokenExpiradoMsg = "Access Token inválido o expirado"

// ContextAlumnoID es la clave donde RequireAuthMiddleware guarda el id del alumno autenticado.
const ContextAlumnoID = "alumnoID"

// AlumnoID devuelve el id del alumno autenticado. Solo se puede usar en rutas
// protegidas por RequireAuthMiddleware.
func AlumnoID(c echo.Context) (int64, error) {
	id, ok := c.Get(ContextAlumnoID).(int64)
	if !ok || id <= 0 {
		return 0, NewApiError(http.StatusUnauthorized, fmt.Errorf("No autenticado"))
	}
	return id, nil
}

// ParamID lee un parametro numerico de la ruta.
func ParamID(c echo.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, NewApiError(http.StatusBadRequest, fmt.Errorf("'%s' inválido", name))
	}
	return id, nil
}

type CustomClaims struct {
	Sub int64 `json:"sub"`
	// Tipo separa el access token del refresh token, asi el refresh (que dura
	// 30 dias) no sirve para llamar a la API.
	Tipo string `json:"typ,omitempty"`
	jwt.RegisteredClaims
}

const (
	tipoAccess  = "access"
	tipoRefresh = "refresh"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 30 * 24 * time.Hour
)

func generateToken(alumnoID int64, secretKey, tipo string, duration time.Duration) (string, error) {
	claims := &CustomClaims{
		Sub:  alumnoID,
		Tipo: tipo,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func GenerateAccessToken(alumnoID int64, secretKey string) (string, error) {
	token, err := generateToken(alumnoID, secretKey, tipoAccess, AccessTokenDuration)
	if err != nil {
		return "", fmt.Errorf("error firmando access token: %w", err)
	}
	return token, nil
}

func GenerateRefreshToken(alumnoID int64, secretKey string) (string, error) {
	token, err := generateToken(alumnoID, secretKey, tipoRefresh, RefreshTokenDuration)
	if err != nil {
		return "", fmt.Errorf("error firmando refresh token: %w", err)
	}
	return token, nil
}

// ParseToken valida un JWT firmado con HS256 y devuelve sus claims.
func ParseToken(tokenString, secretKey string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid || claims.Sub <= 0 {
		return nil, fmt.Errorf("token inválido")
	}

	return claims, nil
}

// ParseAccessToken solo acepta access tokens.
func ParseAccessToken(tokenString, secretKey string) (*CustomClaims, error) {
	claims, err := ParseToken(tokenString, secretKey)
	if err != nil {
		return nil, err
	}
	if claims.Tipo != tipoAccess {
		return nil, fmt.Errorf("no es un access token")
	}
	return claims, nil
}

// ParseRefreshToken solo acepta refresh tokens. Los emitidos antes de que
// existiera el tipo no lo tienen y se siguen aceptando hasta que venzan.
func ParseRefreshToken(tokenString, secretKey string) (*CustomClaims, error) {
	claims, err := ParseToken(tokenString, secretKey)
	if err != nil {
		return nil, err
	}
	if claims.Tipo != tipoRefresh && claims.Tipo != "" {
		return nil, fmt.Errorf("no es un refresh token")
	}
	return claims, nil
}

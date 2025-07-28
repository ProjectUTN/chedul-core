package handlers

import (
	"chedul-core/internals/domain"
	"chedul-core/pkg/config"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AlumnoHandler struct {
	alumnoRepo  domain.AlumnoRepository
	carreraRepo domain.CarreraRepository
	logger      *zap.Logger
	jwtToken    config.Secret
}

func NewAlumnoHandler(alumnoRepo domain.AlumnoRepository, carreraRepo domain.CarreraRepository, logger *zap.Logger, jwtSecret config.Secret) *AlumnoHandler {
	return &AlumnoHandler{
		alumnoRepo:  alumnoRepo,
		logger:      logger,
		carreraRepo: carreraRepo,
		jwtToken:    jwtSecret,
	}
}

func (h *AlumnoHandler) GetAll(c echo.Context) error {
	ctx := c.Request().Context()

	alumnos, err := h.alumnoRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, alumnos)
}

func (h *AlumnoHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}

	alumno, err := h.alumnoRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, alumno)
}

func (h *AlumnoHandler) SignUp(c echo.Context) error {
	ctx := c.Request().Context()
	var req domain.SignUpRequest

	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	alumno, errors := req.Validate()
	if len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	if existing, _ := h.alumnoRepo.GetByEmail(ctx, req.Email); existing != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Credenciales inválidas, Correo ya existente"})
	}

	carrera, err := h.carreraRepo.GetByName(ctx, req.Carrera)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Credenciales inválidas"})
	}
	alumno.Carrera = carrera.ID

	if err := h.alumnoRepo.Create(ctx, &alumno); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, alumno)

}

func (h *AlumnoHandler) LogIn(c echo.Context) error {
	ctx := c.Request().Context()
	var req domain.LoginRequest

	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	alumno, err := h.alumnoRepo.GetByEmail(ctx, email)
	if err != nil || alumno == nil {
		h.logger.Warn("LogIn: Intento de login fallido - credenciales inválidas", zap.String("email", email))
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Credenciales inválidas"})
	}

	if !CheckPassword(req.Password, alumno.Password) {
		h.logger.Warn("LogIn: Intento de login fallido - contraseña incorrecta", zap.String("email", email))
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Credenciales inválidas"})
	}

	jwtSecret := h.jwtToken.Expose()
	if jwtSecret == "" {
		h.logger.Error("LogIn: JwtSecret no configurado")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error de configuración del servidor"})
	}

	accessToken, err := GenerateAccessToken(alumno.ID, jwtSecret)
	if err != nil {
		h.logger.Error("LogIn: No se pudo generar el access token", zap.Error(err), zap.Int64("alumno_id", alumno.ID))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "No se pudo generar el token de acceso"})
	}

	refreshToken, err := GenerateRefreshToken(alumno.ID, jwtSecret)
	if err != nil {
		h.logger.Error("LogIn: No se pudo generar el refresh token", zap.Error(err), zap.Int64("alumno_id", alumno.ID))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "No se pudo generar el token de refresco"})
	}

	refreshTokenCookie := new(http.Cookie)
	refreshTokenCookie.Name = "refreshToken"
	refreshTokenCookie.Value = refreshToken
	refreshTokenCookie.Path = "/"
	refreshTokenCookie.Expires = time.Now().Add(30 * 24 * time.Hour)
	refreshTokenCookie.Secure = config.IsProd()
	refreshTokenCookie.HttpOnly = true
	refreshTokenCookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(refreshTokenCookie)

	h.logger.Info("LogIn: Autenticación exitosa", zap.Int64("alumno_id", alumno.ID), zap.String("email", email))

	return c.JSON(http.StatusOK, map[string]any{
		"accessToken": accessToken,
		"user": map[string]any{
			"id":     alumno.ID,
			"nombre": alumno.Nombre,
			"email":  alumno.Email,
		},
	})
}

func (h *AlumnoHandler) RefreshToken(c echo.Context) error {
	refreshTokenCookie, err := c.Cookie("refreshToken")

	if err != nil || refreshTokenCookie.Value == "" {
		h.logger.Warn("RefreshToken: Refresh Token faltante o inválido en cookie", zap.Error(err))
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Refresh Token faltante"})
	}

	refreshTokenString := refreshTokenCookie.Value
	jwtSecret := h.jwtToken.Expose()

	token, err := jwt.ParseWithClaims(refreshTokenString, &CustomClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			h.logger.Error("RefreshToken: Método de firma inesperado para refresh token", zap.Any("alg", token.Header["alg"]))
			return nil, fmt.Errorf("método de firma inesperado")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		h.logger.Error("RefreshToken: Error al parsear o validar refresh token", zap.Error(err))
		// Si el refresh token es inválido o expirado, forzar logout
		c.SetCookie(&http.Cookie{
			Name:     "refreshToken",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   config.IsProd(),
			SameSite: http.SameSiteLaxMode,
		})
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Refresh Token inválido o expirado. Por favor, inicie sesión de nuevo."})
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		h.logger.Warn("RefreshToken: Claims de refresh token inválidos o token no válido")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Refresh Token inválido"})
	}

	newAccessToken, err := GenerateAccessToken(claims.Sub, jwtSecret)
	if err != nil {
		h.logger.Error("RefreshToken: No se pudo generar un nuevo access token", zap.Error(err), zap.Int64("alumno_id", claims.Sub))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "No se pudo generar un nuevo token de acceso"})
	}

	h.logger.Info("RefreshToken: Nuevo access token generado", zap.Int64("alumno_id", claims.Sub))

	return c.JSON(http.StatusOK, map[string]string{"accessToken": newAccessToken})
}

func (h *AlumnoHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return err
	}

	var req domain.AlumnoRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	if errors := req.Validate(); len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	alumno, err := h.alumnoRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	existing, _ := h.alumnoRepo.GetByEmail(ctx, req.Email)
	if existing != nil && existing.ID != id {
		// TODO: Mejorar el error
		return InvalidJSON()
	}

	email, err := domain.NewEmail(req.Email)
	if err != nil {
		return err
	}
	alumno.Email = email

	carrera, err := h.carreraRepo.GetByName(ctx, req.Carrera)
	if err != nil {
		// TODO: mejorar el error
		return InvalidJSON()
	}
	alumno.Carrera = carrera.ID

	if err := h.alumnoRepo.Update(ctx, alumno); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, alumno)
}

func (h *AlumnoHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return InvalidJSON()
	}

	if err := h.alumnoRepo.Delete(ctx, id); err != nil {
		h.logger.Error("failed to delete alumno", zap.Error(err), zap.Int64("id", id))
		return err
	}

	return c.JSON(http.StatusOK, "Alumno deleted successfully")
}

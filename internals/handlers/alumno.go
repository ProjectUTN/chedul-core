package handlers

import (
	"chedul-core/internals/domain"
	"chedul-core/pkg/config"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"go.uber.org/zap"
)

const refreshTokenCookieName = "refreshToken"

type AlumnoHandler struct {
	alumnoRepo     domain.AlumnoRepository
	carreraRepo    domain.CarreraRepository
	logger         *zap.Logger
	jwtToken       config.Secret
	googleClientID string
	validarGoogle  ValidadorGoogle
}

func NewAlumnoHandler(alumnoRepo domain.AlumnoRepository, carreraRepo domain.CarreraRepository, logger *zap.Logger, jwtSecret config.Secret, googleClientID string) *AlumnoHandler {
	return &AlumnoHandler{
		alumnoRepo:     alumnoRepo,
		logger:         logger,
		carreraRepo:    carreraRepo,
		jwtToken:       jwtSecret,
		googleClientID: googleClientID,
		validarGoogle:  validarTokenGoogle,
	}
}

// GetMe devuelve el perfil del alumno autenticado.
func (h *AlumnoHandler) GetMe(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := AlumnoID(c)
	if err != nil {
		return err
	}

	alumno, err := h.alumnoRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NotFound("Alumno")
		}
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

	alumno, errs := req.Validate()
	if len(errs) > 0 {
		return InvalidRequestData(errs)
	}

	if existing, _ := h.alumnoRepo.GetByEmail(ctx, alumno.Email.String()); existing != nil {
		return InvalidRequestData(map[string]string{"email": "Ya existe una cuenta con ese correo"})
	}

	if _, err := h.carreraRepo.GetByID(ctx, req.CarreraID); err != nil {
		return InvalidRequestData(map[string]string{"carrera_id": "La carrera no existe"})
	}

	if err := h.alumnoRepo.Create(ctx, &alumno); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, alumno)
}

func (h *AlumnoHandler) setRefreshCookie(c echo.Context, value string, expires time.Time) {
	cookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   config.IsProd(),
		SameSite: http.SameSiteLaxMode,
	}

	// En produccion la API puede estar en otro dominio que el frontend; sin
	// SameSite=None el navegador no manda la cookie. Igual conviene que el
	// frontend haga de proxy de /api (ver vercel.json del front): Safari
	// bloquea las cookies de terceros aunque tengan SameSite=None.
	if config.IsProd() {
		cookie.SameSite = http.SameSiteNoneMode
	}

	c.SetCookie(cookie)
}

func (h *AlumnoHandler) LogIn(c echo.Context) error {
	ctx := c.Request().Context()
	var req domain.LoginRequest

	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	credencialesInvalidas := NewApiError(http.StatusUnauthorized, fmt.Errorf("Credenciales inválidas"))

	email, err := domain.ParseEmail(req.Email)
	if err != nil {
		return credencialesInvalidas
	}

	alumno, err := h.alumnoRepo.GetByEmail(ctx, email.String())
	if err != nil || alumno == nil {
		h.logger.Warn("LogIn: intento de login fallido, el correo no existe")
		return credencialesInvalidas
	}

	ok, err := domain.CheckPassword(alumno.Password.String(), req.Password)
	if err != nil || !ok {
		h.logger.Warn("LogIn: intento de login fallido, clave incorrecta", zap.Int64("alumno_id", alumno.ID))
		return credencialesInvalidas
	}

	h.logger.Info("LogIn: autenticación exitosa", zap.Int64("alumno_id", alumno.ID))

	return h.iniciarSesion(c, alumno)
}

// iniciarSesion devuelve el access token y deja el refresh token en la cookie.
func (h *AlumnoHandler) iniciarSesion(c echo.Context, alumno *domain.Alumno) error {
	jwtSecret := h.jwtToken.Expose()

	accessToken, err := GenerateAccessToken(alumno.ID, jwtSecret)
	if err != nil {
		return err
	}

	refreshToken, err := GenerateRefreshToken(alumno.ID, jwtSecret)
	if err != nil {
		return err
	}

	h.setRefreshCookie(c, refreshToken, time.Now().Add(RefreshTokenDuration))

	return c.JSON(http.StatusOK, map[string]any{
		"accessToken": accessToken,
		"user":        alumno,
	})
}

func (h *AlumnoHandler) LogOut(c echo.Context) error {
	h.setRefreshCookie(c, "", time.Unix(0, 0))
	return c.NoContent(http.StatusNoContent)
}

func (h *AlumnoHandler) RefreshToken(c echo.Context) error {
	ctx := c.Request().Context()

	refreshTokenCookie, err := c.Cookie(refreshTokenCookieName)
	if err != nil || refreshTokenCookie.Value == "" {
		return NewApiError(http.StatusUnauthorized, fmt.Errorf("Refresh Token faltante"))
	}

	jwtSecret := h.jwtToken.Expose()

	claims, err := ParseToken(refreshTokenCookie.Value, jwtSecret)
	if err != nil {
		h.setRefreshCookie(c, "", time.Unix(0, 0))
		return NewApiError(http.StatusUnauthorized, fmt.Errorf("Refresh Token inválido o expirado. Por favor, inicie sesión de nuevo."))
	}

	// Si el alumno borro su cuenta el refresh token deja de servir
	alumno, err := h.alumnoRepo.GetByID(ctx, claims.Sub)
	if err != nil {
		h.setRefreshCookie(c, "", time.Unix(0, 0))
		return NewApiError(http.StatusUnauthorized, fmt.Errorf("Refresh Token inválido"))
	}

	newAccessToken, err := GenerateAccessToken(alumno.ID, jwtSecret)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{
		"accessToken": newAccessToken,
		"user":        alumno,
	})
}

// UpdateMe actualiza el nombre y la carrera del alumno autenticado.
func (h *AlumnoHandler) UpdateMe(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := AlumnoID(c)
	if err != nil {
		return err
	}

	var req domain.ActualizarAlumnoRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	nombre, errs := req.Validate()
	if len(errs) > 0 {
		return InvalidRequestData(errs)
	}

	if _, err := h.carreraRepo.GetByID(ctx, req.CarreraID); err != nil {
		return InvalidRequestData(map[string]string{"carrera_id": "La carrera no existe"})
	}

	alumno, err := h.alumnoRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	alumno.Nombre = nombre
	alumno.Carrera = req.CarreraID

	if err := h.alumnoRepo.Update(ctx, alumno); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, alumno)
}

// CambiarPassword cambia la clave del alumno. Pide la actual para que no se
// pueda cambiar con una sesion abierta ajena.
func (h *AlumnoHandler) CambiarPassword(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := AlumnoID(c)
	if err != nil {
		return err
	}

	var req struct {
		Actual string `json:"actual"`
		Nueva  string `json:"nueva"`
	}
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	alumno, err := h.alumnoRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if ok, err := domain.CheckPassword(alumno.Password.String(), req.Actual); err != nil || !ok {
		return InvalidRequestData(map[string]string{"actual": "La contraseña actual no es correcta"})
	}
	nueva, err := domain.ParsePassword(req.Nueva)
	if err != nil {
		return InvalidRequestData(map[string]string{"nueva": err.Error()})
	}

	alumno.Password = nueva
	if err := h.alumnoRepo.Update(ctx, alumno); err != nil {
		return err
	}
	h.logger.Info("Clave cambiada", zap.Int64("alumno_id", id))
	return c.NoContent(http.StatusNoContent)
}

// DeleteMe borra la cuenta del alumno autenticado.
func (h *AlumnoHandler) DeleteMe(c echo.Context) error {
	ctx := c.Request().Context()

	id, err := AlumnoID(c)
	if err != nil {
		return err
	}

	if err := h.alumnoRepo.Delete(ctx, id); err != nil {
		h.logger.Error("Fallo al eliminar alumno", zap.Error(err), zap.Int64("id", id))
		return err
	}

	h.setRefreshCookie(c, "", time.Unix(0, 0))
	return c.NoContent(http.StatusNoContent)
}

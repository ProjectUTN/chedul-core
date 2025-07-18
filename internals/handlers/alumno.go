package handlers

import (
	"chedul-core/internals/domain"
	"chedul-core/pkg/config"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type AlumnoHandler struct {
	alumnoRepo  domain.AlumnoRepository
	carreraRepo domain.CarreraRepository
	logger      *zap.Logger
	appConfig   *config.AppConfig 
}

func NewAlumnoHandler(alumnoRepo domain.AlumnoRepository, carreraRepo domain.CarreraRepository, logger *zap.Logger, cfg *config.AppConfig) *AlumnoHandler {
	return &AlumnoHandler{
		alumnoRepo:  alumnoRepo,
		logger:      logger,
		carreraRepo: carreraRepo,
		appConfig:   cfg,
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

func (h *AlumnoHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var req domain.AlumnoRequest
	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}
	

	if errors := req.Validate(); len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	if existing, _ := h.alumnoRepo.GetByEmail(ctx, req.Email); existing != nil {
		// TODO: Cambiar el error
		return InvalidJSON()
	}

	carrera, err := h.carreraRepo.GetByName(ctx, req.Carrera)
	if err != nil {
		// TODO: Cambiar el error
		return InvalidJSON()
	}

	alumno := &domain.Alumno{
		Nombre:  strings.TrimSpace(req.Nombre),
		Email:   strings.ToLower(strings.TrimSpace(req.Email)),
		Carrera: carrera.ID,
	}

	if err := h.alumnoRepo.Create(ctx, alumno); err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, alumno)
}

func  (h *AlumnoHandler) SignUp(c echo.Context) error {
	ctx := c.Request().Context()
	var req domain.SignUpRequest

	if err := c.Bind(&req); err != nil {
		return InvalidJSON()
	}

	if errors := req.Validate(); len(errors) > 0 {
		return InvalidRequestData(errors)
	}

	if existing, _ := h.alumnoRepo.GetByEmail(ctx, req.Email); existing != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Credenciales inválidas"})
	}

	carrera, err := h.carreraRepo.GetByName(ctx, req.Carrera)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Credenciales inválidas"})
	}

	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error al encriptar la contraseña"})
	}

	alumno := &domain.Alumno{
		Nombre:   req.Nombre,
		Email:    strings.ToLower(req.Email),
		Carrera:  carrera.ID,
		Password: hashedPassword,
	}

	if err := h.alumnoRepo.Create(ctx, alumno); err != nil {
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
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Credenciales inválidas"})
	}

	if !CheckPassword(req.Password, alumno.Password) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Credenciales inválidas"})
	}
	
	h.logger.Info("LogIn: JwtSecret cargado", zap.String("value", h.appConfig.JwtSecret))

	secretKeyBytes, err := base64.StdEncoding.DecodeString(h.appConfig.JwtSecret)
	if err != nil {
		h.logger.Error("LogIn: Error decodificando JwtSecret", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error interno al decodificar la clave"})
	}
	h.logger.Info("LogIn: Clave decodificada", zap.String("key", fmt.Sprintf("%x", secretKeyBytes)))

	token, err := GenerateJWT(alumno.ID, h.appConfig.JwtSecret)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "No se pudo generar el token"})
	}

	cookie := new(http.Cookie)
	cookie.Name = "Authorization"
	cookie.Value = token
	cookie.Path = "/"
	cookie.Expires = time.Now().Add(30 * 24 * time.Hour)
	cookie.Secure = h.appConfig.IsProd()
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":    alumno.ID,
			"nombre": alumno.Nombre,
			"email": alumno.Email,
		},
	})
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
	alumno.Email = req.Email

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


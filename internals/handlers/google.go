package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/api/idtoken"

	"chedul-core/internals/domain"
)

// DatosGoogle es lo que se usa del token de Google para identificar al alumno.
type DatosGoogle struct {
	Email           string
	EmailVerificado bool
	Nombre          string
	// given_name y family_name, si Google los manda
	NombrePila string
	Apellido   string
}

// ValidadorGoogle verifica el ID token que devuelve el boton de Google.
type ValidadorGoogle func(ctx context.Context, token, clientID string) (*DatosGoogle, error)

func validarTokenGoogle(ctx context.Context, token, clientID string) (*DatosGoogle, error) {
	payload, err := idtoken.Validate(ctx, token, clientID)
	if err != nil {
		return nil, err
	}
	datos := &DatosGoogle{}
	datos.Email, _ = payload.Claims["email"].(string)
	datos.EmailVerificado, _ = payload.Claims["email_verified"].(bool)
	datos.Nombre, _ = payload.Claims["name"].(string)
	datos.NombrePila, _ = payload.Claims["given_name"].(string)
	datos.Apellido, _ = payload.Claims["family_name"].(string)
	return datos, nil
}

// SetValidadorGoogle reemplaza la verificacion contra Google (para tests).
func (h *AlumnoHandler) SetValidadorGoogle(v ValidadorGoogle) {
	h.validarGoogle = v
}

// GoogleConfig le dice al front si mostrar el boton y con que client ID.
func (h *AlumnoHandler) GoogleConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"client_id": h.googleClientID})
}

type loginGoogleRequest struct {
	Credential string `json:"credential"`
}

// LogInGoogle inicia sesion con el ID token de Google. Si no hay cuenta con
// ese mail se crea una en la primera carrera, con una clave al azar que el
// alumno no conoce (puede seguir entrando con Google).
func (h *AlumnoHandler) LogInGoogle(c echo.Context) error {
	ctx := c.Request().Context()

	if h.googleClientID == "" {
		return NewApiError(http.StatusNotFound, fmt.Errorf("El login con Google no está habilitado"))
	}

	var req loginGoogleRequest
	if err := c.Bind(&req); err != nil || req.Credential == "" {
		return InvalidJSON()
	}

	noValido := NewApiError(http.StatusUnauthorized, fmt.Errorf("No se pudo verificar tu cuenta de Google"))
	datos, err := h.validarGoogle(ctx, req.Credential, h.googleClientID)
	if err != nil {
		h.logger.Warn("LogInGoogle: token invalido", zap.Error(err))
		return noValido
	}
	if !datos.EmailVerificado {
		return noValido
	}

	email, err := domain.ParseEmail(datos.Email)
	if err != nil {
		return noValido
	}

	alumno, _ := h.alumnoRepo.GetByEmail(ctx, email.String())
	if alumno == nil {
		alumno, err = h.crearAlumnoGoogle(ctx, email, datos)
		if err != nil {
			return err
		}
		if err := h.alumnoRepo.VincularGoogle(ctx, alumno.ID, nil); err != nil {
			return err
		}
		h.logger.Info("LogInGoogle: cuenta nueva", zap.Int64("alumno_id", alumno.ID))
	} else if !alumno.GoogleVinculado {
		// La cuenta se creo con clave y el registro no verifica el correo:
		// pudo haberla creado otra persona con este mail. Google si lo
		// verifico, asi que el dueño es quien entra ahora: se borra la clave
		// y se cierran las otras sesiones.
		clave, err := claveAlAzar()
		if err != nil {
			return err
		}
		if err := h.alumnoRepo.VincularGoogle(ctx, alumno.ID, &clave); err != nil {
			return err
		}
		alumno, err = h.alumnoRepo.GetByID(ctx, alumno.ID)
		if err != nil {
			return err
		}
		h.logger.Info("LogInGoogle: cuenta con clave vinculada a Google", zap.Int64("alumno_id", alumno.ID))
	}

	h.logger.Info("LogInGoogle: autenticación exitosa", zap.Int64("alumno_id", alumno.ID))
	return h.iniciarSesion(c, alumno, "google")
}

func (h *AlumnoHandler) crearAlumnoGoogle(ctx context.Context, email domain.Email, datos *DatosGoogle) (*domain.Alumno, error) {
	// Con nombre y apellido separados se guardan asi; si no, todo va en el nombre
	nombreGoogle, apellidoGoogle := datos.Nombre, ""
	if datos.NombrePila != "" && datos.Apellido != "" {
		nombreGoogle, apellidoGoogle = datos.NombrePila, datos.Apellido
	}
	apellido, _ := domain.ParseApellido(apellidoGoogle)
	nombre, err := domain.NewUsername(nombreGoogle)
	if err != nil {
		apellido = ""
		nombre, err = domain.NewUsername(strings.Split(email.String(), "@")[0])
		if err != nil {
			return nil, err
		}
	}

	carreras, err := h.carreraRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	if len(carreras) == 0 {
		return nil, fmt.Errorf("no hay carreras cargadas")
	}

	clave, err := claveAlAzar()
	if err != nil {
		return nil, err
	}

	alumno := &domain.Alumno{
		Nombre:   nombre,
		Apellido: apellido,
		Email:    email,
		Carrera:  carreras[0].ID,
		Password: clave,
	}
	if err := h.alumnoRepo.Create(ctx, alumno); err != nil {
		return nil, err
	}
	return alumno, nil
}

func claveAlAzar() (domain.Password, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return domain.Password{}, err
	}
	// Prefijo fijo para cumplir con mayusculas, minusculas, numeros y simbolos
	return domain.ParsePassword("Gg1!" + base64.RawURLEncoding.EncodeToString(b))
}

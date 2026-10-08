package handlers

import (
	"chedul-core/internals/domain"
	"chedul-core/pkg/config"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// alumnosEnMemoria es un AlumnoRepository minimo para probar el login con Google
type alumnosEnMemoria struct {
	domain.AlumnoRepository
	alumno *domain.Alumno
}

func (r *alumnosEnMemoria) GetByEmail(_ context.Context, email string) (*domain.Alumno, error) {
	if r.alumno != nil && r.alumno.Email.String() == email {
		copia := *r.alumno
		return &copia, nil
	}
	return nil, nil
}

func (r *alumnosEnMemoria) GetByID(_ context.Context, _ int64) (*domain.Alumno, error) {
	copia := *r.alumno
	return &copia, nil
}

func (r *alumnosEnMemoria) VincularGoogle(_ context.Context, _ int64, clave *domain.Password) error {
	r.alumno.GoogleVinculado = true
	if clave != nil {
		r.alumno.Password = *clave
		r.alumno.VersionSesion++
	}
	return nil
}

func (r *alumnosEnMemoria) RegistrarAcceso(context.Context, int64, string) error { return nil }

func (r *alumnosEnMemoria) loginGoogle(t *testing.T) {
	h := NewAlumnoHandler(r, nil, zap.NewNop(), *config.NewSecret("secreto-de-prueba-secreto-de-prueba"), "cliente")
	h.SetValidadorGoogle(func(context.Context, string, string) (*DatosGoogle, error) {
		return &DatosGoogle{Email: "duenio@gmail.com", EmailVerificado: true, Nombre: "Dueño"}, nil
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/google", strings.NewReader(`{"credential":"x"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	require.NoError(t, h.LogInGoogle(echo.New().NewContext(req, rec)))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
}

// Alguien crea una cuenta con el mail de otro y una clave suya. Cuando el
// dueño entra con Google, esa clave deja de servir y las sesiones se cierran.
func TestLoginGoogleBorraLaClaveDeUnaCuentaAjena(t *testing.T) {
	email, _ := domain.ParseEmail("duenio@gmail.com")
	clave, err := domain.ParsePassword("clave-del-intruso")
	require.NoError(t, err)
	repo := &alumnosEnMemoria{alumno: &domain.Alumno{ID: 1, Email: email, Password: clave}}

	repo.loginGoogle(t)

	assert.True(t, repo.alumno.GoogleVinculado)
	assert.Equal(t, 1, repo.alumno.VersionSesion, "se cierran las sesiones anteriores")
	ok, _ := domain.CheckPassword(repo.alumno.Password.String(), "clave-del-intruso")
	assert.False(t, ok, "la clave del intruso ya no sirve")

	// La segunda vez ya esta vinculada: no se toca nada
	repo.loginGoogle(t)
	assert.Equal(t, 1, repo.alumno.VersionSesion)
}

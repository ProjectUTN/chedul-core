package domain

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
)

type Alumno struct {
	ID     int64    `json:"id"`
	Nombre Username `json:"nombre"`
	// Puede estar vacio en las cuentas de antes de que existiera el campo
	Apellido string   `json:"apellido"`
	Email    Email    `json:"email"`
	Carrera  int64    `json:"carrera"`
	Password Password `json:"-"`
	// Solo los administradores reciben este campo (y ven el panel). Se
	// completa con la lista de config.Admins, no se guarda en la base.
	EsAdmin bool `json:"es_admin,omitempty"`
	// Ya entro alguna vez con Google
	GoogleVinculado bool `json:"-"`
	// Va dentro de los tokens; al subirla los refresh tokens viejos dejan de servir
	VersionSesion int `json:"-"`
}

// ActualizarAlumnoRequest es lo que el alumno puede cambiar de su propio perfil.
type ActualizarAlumnoRequest struct {
	Nombre string `json:"nombre"`
	// nil deja el apellido como estaba
	Apellido  *string `json:"apellido"`
	CarreraID int64   `json:"carrera_id"`
}

type SignUpRequest struct {
	Nombre    string `json:"nombre"`
	Apellido  string `json:"apellido"`
	Email     string `json:"email"`
	CarreraID int64  `json:"carrera_id"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (self *ActualizarAlumnoRequest) Validate() (Username, map[string]string) {
	errors := make(map[string]string)

	nombre, err := NewUsername(self.Nombre)
	if err != nil {
		errors["nombre"] = err.Error()
	}

	if self.Apellido != nil {
		if _, err := ParseApellido(*self.Apellido); err != nil {
			errors["apellido"] = err.Error()
		}
	}

	if self.CarreraID <= 0 {
		errors["carrera_id"] = "Carrera es requerida"
	}

	return nombre, errors
}

func (self *SignUpRequest) Validate() (Alumno, map[string]string) {
	errors := make(map[string]string)

	nombre, err := NewUsername(self.Nombre)
	if err != nil {
		errors["nombre"] = err.Error()
	}

	apellido, err := ParseApellido(self.Apellido)
	if err != nil {
		errors["apellido"] = err.Error()
	}

	email, err := ParseEmail(self.Email)
	if err != nil {
		errors["email"] = err.Error()
	}

	clave, err := ParsePassword(self.Password)
	if err != nil {
		errors["password"] = err.Error()
	}

	if self.CarreraID <= 0 {
		errors["carrera_id"] = "Carrera es requerida"
	}

	return Alumno{
		Nombre:   nombre,
		Apellido: apellido,
		Email:    email,
		Carrera:  self.CarreraID,
		Password: clave,
	}, errors
}

type AlumnoRepository interface {
	GetByID(ctx context.Context, id int64) (*Alumno, error)
	GetByEmail(ctx context.Context, email string) (*Alumno, error)
	Create(ctx context.Context, alumno *Alumno) error
	Update(ctx context.Context, alumno *Alumno) error
	Delete(ctx context.Context, id int64) error
	// RegistrarAcceso guarda un inicio de sesion (metodo: clave, google, registro)
	RegistrarAcceso(ctx context.Context, id int64, metodo string) error
	// MarcarActivo actualiza la ultima vez que el alumno uso la app
	MarcarActivo(ctx context.Context, id int64) error
	// VincularGoogle marca que el alumno entro con Google. Si nuevaClave no es
	// nil reemplaza la clave y sube la version de sesion.
	VincularGoogle(ctx context.Context, id int64, nuevaClave *Password) error
	// SubirVersionSesion invalida los refresh tokens emitidos hasta ahora
	SubirVersionSesion(ctx context.Context, id int64) (int, error)
	// Contar devuelve cuantos alumnos hay registrados
	Contar(ctx context.Context) (int, error)
}

// ParseApellido recorta el apellido y lo valida. Vacio es valido (no es
// obligatorio para las cuentas viejas ni para quien entra con Google).
func ParseApellido(s string) (string, error) {
	s = strings.TrimSpace(s)
	if uniseg.GraphemeClusterCount(s) > 100 {
		return "", fmt.Errorf("El apellido no puede tener más de 100 caracteres")
	}
	if strings.ContainsAny(s, "/()\"<>\\{}") {
		return "", fmt.Errorf("El apellido tiene caracteres no permitidos")
	}
	return s, nil
}

// NombreCorto es como se muestra un alumno a los demas (aportes, ranking):
// nombre de pila e inicial del apellido, "Juan P.". Si la cuenta no tiene
// apellido, se usa la ultima palabra del nombre como antes.
func NombreCorto(nombre, apellido string) string {
	partes := strings.Fields(nombre)
	if len(partes) == 0 {
		return "Anónimo"
	}
	resto := strings.TrimSpace(apellido)
	if resto == "" {
		if len(partes) == 1 {
			return partes[0]
		}
		resto = partes[len(partes)-1]
	}
	for _, r := range resto {
		if unicode.IsLetter(r) {
			return partes[0] + " " + strings.ToUpper(string(r)) + "."
		}
	}
	return partes[0]
}

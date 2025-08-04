package domain

import (
	"context"
	"strings"
)

type Alumno struct {
	ID       int64    `json:"id"`
	Nombre   UserName `json:"nombre"`
	Email    Email    `json:"email"`
	Carrera  int64    `json:"carrera"`
	Password Password `json:"-"`
}

// TODO: Solo se utiliza en Update(), ver y cambiarlo
type AlumnoRequest struct {
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Carrera  string `json:"carrera"`
	Password string `json:"password"`
}

type SignUpRequest struct {
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
	Carrera  string `json:"carrera"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TODO: Deuda Tecnica
func (self *AlumnoRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if _, err := NewUserName(self.Nombre); err != nil {
		errors["nombre"] = err.Error()
	}

	if _, err := ParseEmail(self.Email); err != nil {
		errors["email"] = err.Error()
	}

	if strings.TrimSpace(self.Carrera) == "" {
		errors["carrera"] = "Carrera es requerida"
	}

	// TODO: Expandir los requisitos de una contrasena
	if len(self.Password) < 6 {
		errors["password"] = "La contraseña debe tener al menos 6 caracteres"
	}

	return errors
}

func (self *SignUpRequest) Validate() (Alumno, map[string]string) {
	errors := make(map[string]string)

	nombre, err := NewUserName(self.Nombre)
	if err != nil {
		errors["nombre"] = err.Error()
	}

	email, err := ParseEmail(self.Email)
	if err != nil {
		errors["email"] = err.Error()
	}

	clave, err := NewPassword(self.Password)
	if err != nil {
		errors["password"] = err.Error()
	}

	if strings.TrimSpace(self.Carrera) == "" {
		errors["carrera"] = "Carrera es requerida"
	}

	return Alumno{
		Nombre:   nombre,
		Email:    email,
		Password: clave,
	}, errors
}

type AlumnoRepository interface {
	GetAll(ctx context.Context) ([]Alumno, error)
	GetByID(ctx context.Context, id int64) (*Alumno, error)
	GetByEmail(ctx context.Context, email string) (*Alumno, error)
	Create(ctx context.Context, alumno *Alumno) error
	Update(ctx context.Context, alumno *Alumno) error
	Delete(ctx context.Context, id int64) error
}

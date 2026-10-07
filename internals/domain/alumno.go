package domain

import (
	"context"
)

type Alumno struct {
	ID       int64    `json:"id"`
	Nombre   Username `json:"nombre"`
	Email    Email    `json:"email"`
	Carrera  int64    `json:"carrera"`
	Password Password `json:"-"`
}

// ActualizarAlumnoRequest es lo que el alumno puede cambiar de su propio perfil.
type ActualizarAlumnoRequest struct {
	Nombre    string `json:"nombre"`
	CarreraID int64  `json:"carrera_id"`
}

type SignUpRequest struct {
	Nombre    string `json:"nombre"`
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
}

package domain

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/rivo/uniseg"
	"golang.org/x/crypto/bcrypt"
)

type Alumno struct {
	ID       int64    `json:"id"`
	Nombre   UserName `json:"nombre"`
	Email    Email    `json:"email"`
	Carrera  int64    `json:"carrera"`
	Password string
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

	if _, err := NewEmail(self.Email); err != nil {
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

	email, err := NewEmail(self.Email)
	if err != nil {
		errors["email"] = err.Error()
	}

	if strings.TrimSpace(self.Carrera) == "" {
		errors["carrera"] = "Carrera es requerida"
	}

	// TODO: Expandir los requisitos de una contrasena
	if len(self.Password) < 6 {
		errors["password"] = "La contraseña debe tener al menos 6 caracteres"
	}

	hashedPassword, err := HashPassword(self.Password)
	if err != nil {
		errors["password"] = "Error al encriptar la contraseña"
	}

	return Alumno{
		Nombre:   nombre,
		Email:    email,
		Password: hashedPassword,
	}, nil
}

type AlumnoRepository interface {
	GetAll(ctx context.Context) ([]Alumno, error)
	GetByID(ctx context.Context, id int64) (*Alumno, error)
	GetByEmail(ctx context.Context, email string) (*Alumno, error)
	Create(ctx context.Context, alumno *Alumno) error
	Update(ctx context.Context, alumno *Alumno) error
	Delete(ctx context.Context, id int64) error
}

type Email struct {
	value string
}

func NewEmail(s string) (Email, error) {
	addr, err := mail.ParseAddress(s)

	if err != nil {
		return Email{}, fmt.Errorf("formato invalido: %v", err)
	}

	return Email{value: addr.Address}, nil
}

func (e *Email) String() string {
	return e.value
}

type UserName struct {
	value string
}

func (u *UserName) String() string {
	return u.value
}

func NewUserName(s string) (UserName, error) {
	emptyOrWhitespace := strings.TrimSpace(s) == ""
	isTooLong := uniseg.GraphemeClusterCount(s) > 256
	forbiddenChars := []rune{'/', '(', ')', '"', '<', '>', '\\', '{', '}'}
	hasInvalidChar := false
	for _, char := range s {
		for _, forbidden := range forbiddenChars {
			if char == forbidden {
				hasInvalidChar = true
			}
		}
	}

	if hasInvalidChar || emptyOrWhitespace || isTooLong {
		return UserName{}, fmt.Errorf("''%s' es un nombre invalido", s)
	}

	return UserName{value: s}, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

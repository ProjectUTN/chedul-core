package domain

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	"github.com/rivo/uniseg"
	"golang.org/x/crypto/bcrypt"
)

type Alumno struct {
	ID       int64    `json:"id"`
	Nombre   UserName `json:"nombre"`
	Email    Email    `json:"email"`
	Carrera  int64    `json:"carrera"`
	Password Password
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

type Password struct {
	inner string
}

func NewPassword(s string) (Password, error) {
	if len(s) < 8 {
		return Password{}, fmt.Errorf("La clave debe tener minimo 8 caracteres")
	}

	if len(s) > 128 {
		return Password{}, fmt.Errorf("la clave no puede exceder 128 caracteres")
	}

	if !hasCharacterDiversity(s) {
		return Password{}, fmt.Errorf("la clave debe contener al menos 3 de los siguientes tipos: mayúsculas, minúsculas, números y símbolos")
	}

	password, err := hash(s)
	if err != nil {
		return Password{}, fmt.Errorf("La clave no pudo ser encriptada")
	}

	return Password{
		inner: password,
	}, nil
}

// TODO: No me gusta tener algo asi. Usar con cuidado
func NewPasswordFromEncrypted(s string) Password {
	return Password{
		inner: s,
	}
}

func (p *Password) String() string {
	return p.inner
}

// FIX: Reemplazar esto con argon2id
func hash(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	return string(hash), err
}

func hasCharacterDiversity(password string) bool {
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	count := 0

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSymbol {
		count++
	}

	return count >= 3
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

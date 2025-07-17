package domain

import (
	"context"
	"fmt"
	"net/mail"
	"strings"

	"github.com/rivo/uniseg"
)

type Alumno struct {
	ID      int64  `json:"id"`
	Nombre  string `json:"nombre"`
	Email   string `json:"email"`
	Carrera int64  `json:"carrera"`
}

type AlumnoRequest struct {
	Nombre  string `json:"nombre"`
	Email   string `json:"email"`
	Carrera string `json:"carrera"`
}

func (self *AlumnoRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if _, err := newAlumnoName(self.Nombre); err != nil {
		errors["nombre"] = err.Error()
	}

	if _, err := newAlumnoEmail(self.Email); err != nil {
		errors["email"] = err.Error()
	}

	if self.Carrera == "" {
		errors["carrera"] = fmt.Sprintf("'%v' no es una carrera valida", self.Carrera)
	}

	return errors
}

type AlumnoRepository interface {
	GetAll(ctx context.Context) ([]Alumno, error)
	GetByID(ctx context.Context, id int64) (*Alumno, error)
	GetByEmail(ctx context.Context, email string) (*Alumno, error)
	Create(ctx context.Context, alumno *Alumno) error
	Update(ctx context.Context, alumno *Alumno) error
	Delete(ctx context.Context, id int64) error
}

type AlumnoService interface {
	GetAll(ctx context.Context) ([]Alumno, error)
	GetByID(ctx context.Context, id int64) (*Alumno, error)
	Create(ctx context.Context, req AlumnoRequest) (*Alumno, error)
	Update(ctx context.Context, id int64, req AlumnoRequest) (*Alumno, error)
	Delete(ctx context.Context, id int64) error
}

type alumnoEmail string

func newAlumnoEmail(s string) (alumnoEmail, error) {
	addr, err := mail.ParseAddress(s)

	if err != nil {
		return "", fmt.Errorf("formato invalido: %v", err)
	}

	return alumnoEmail(addr.Address), nil
}

type alumnoName string

func newAlumnoName(s string) (alumnoName, error) {
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
		return "", fmt.Errorf("''%s' es un nombre invalido", s)
	}

	return alumnoName(s), nil
}

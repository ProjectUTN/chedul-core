package domain

import (
	"context"
	"fmt"
)

const (
	NotaAprobadaMin = 6
	NotaMax         = 10
)

type CondicionAlumno struct {
	ID          int64 `json:"id"`
	CondicionID int64 `json:"condicion_id"`
	Nota        *int  `json:"nota"`
	AlumnoID    int64 `json:"alumno_id"`
	MateriaID   int64 `json:"materia_id"`
}

type SetCondicionRequest struct {
	CondicionID int64 `json:"condicion_id"`
	Nota        *int  `json:"nota"`
}

func (r *SetCondicionRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if r.CondicionID <= 0 {
		errors["condicion_id"] = "Condición es requerida"
	}

	// En la UTN se aprueba con 6 o mas
	if r.Nota != nil && (*r.Nota < NotaAprobadaMin || *r.Nota > NotaMax) {
		errors["nota"] = fmt.Sprintf("La nota de una aprobada va de 6 a 10, se recibió %d", *r.Nota)
	}

	return errors
}

// CondicionPorAlumno es la condicion de un alumno en una materia, con los
// nombres resueltos para mostrar en el frontend.
type CondicionPorAlumno struct {
	MateriaID   int64  `json:"materia_id" bun:"materia_id"`
	Materia     string `json:"materia" bun:"materia"`
	CondicionID int64  `json:"condicion_id" bun:"condicion_id"`
	Condicion   string `json:"condicion" bun:"condicion"`
	Nota        *int   `json:"nota" bun:"nota"`
}

type CondicionAlumnoRepository interface {
	GetCondicionPorAlumno(ctx context.Context, alumnoID int64) ([]CondicionPorAlumno, error)
	// Set crea o reemplaza la condicion del alumno en la materia
	Set(ctx context.Context, condicionAlumno *CondicionAlumno) error
	Delete(ctx context.Context, alumnoID, materiaID int64) error
}

package domain

import (
	"context"
)

type CondicionAlumno struct {
	ID          int64 `json:"id"`
	CondicionID int64 `json:"condicion_id"`
	Nota        int   `json:"nota"`
	AlumnoID    int64 `json:"alumno_id"`
	MateriaID   int64 `json:"materia_id"`
}

type SetCondicionRequest struct {
	CondicionID int64 `json:"condicion_id"`
	Nota        int   `json:"nota"`
}

type CondicionPorAlumno struct {
	Nota      *int       `json:"nota"`
	Materia   *Materia   `json:"materia"`
	Condicion *Condicion `json:"condicion"`
}

type CondicionAlumnoRepository interface {
	GetAll(ctx context.Context) ([]CondicionAlumno, error)
	GetByID(ctx context.Context, id int64) (*CondicionAlumno, error)
	Create(ctx context.Context, condicionAlumno *CondicionAlumno) error
	Update(ctx context.Context, condicionAlumno *CondicionAlumno) error
	GetCondicionPorAlumno(ctx context.Context, id int64) ([]CondicionPorAlumno, error)
	// Delete(ctx context.Context, id int64) error
}

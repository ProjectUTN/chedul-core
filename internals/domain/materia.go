package domain

import "context"

type Materia struct {
	ID             int64   `json:"id"`
	Nombre         string  `json:"nombre"`
	CargaHoraria   int64   `json:"carga_horaria"`
	CorrelativaID  *int64  `json:"correlativa_id"`
	Nivel          int64   `json:"nivel"`
	Area           string  `json:"area"`
	Tipo           string  `json:"tipo"`
	CuatrimestreID int64   `json:"cuatrimestre_id"`
	Horas          float64 `json:"horas"`
	Bloque         string  `json:"bloque"`
	Programa       string  `json:"programa"`
}

type MateriaRepository interface {
	GetAll(ctx context.Context) ([]Materia, error)
	GetByID(ctx context.Context, id int64) (*Materia, error)
}

type MateriaService interface {
	GetAll(ctx context.Context) ([]Materia, error)
	GetByID(ctx context.Context, id int64) (*Materia, error)
}

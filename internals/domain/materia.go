package domain

import "context"

const (
	CorrelativaRegular  = "regular"
	CorrelativaAprobada = "aprobada"
)

// Correlativa indica que para cursar una materia hay que tener otra
// regularizada ("regular") o aprobada ("aprobada").
type Correlativa struct {
	MateriaID int64  `json:"materia_id"`
	Nombre    string `json:"nombre"`
	Tipo      string `json:"tipo"`
}

type Materia struct {
	ID           int64         `json:"id"`
	Codigo       string        `json:"codigo"`
	Nombre       string        `json:"nombre"`
	CargaHoraria int64         `json:"carga_horaria"`
	Nivel        int64         `json:"nivel"`
	Area         string        `json:"area"`
	Tipo         string        `json:"tipo"`
	Cuatrimestre string        `json:"cuatrimestre"`
	Horas        float64       `json:"horas"`
	Bloque       string        `json:"bloque"`
	Programa     string        `json:"programa"`
	Correlativas []Correlativa `json:"correlativas"`
}

// HorarioComision es un dia y franja horaria en que cursa una comision.
// Dia: 1 = lunes ... 7 = domingo, igual que en las clases del alumno.
type HorarioComision struct {
	Dia        int    `json:"dia"`
	HoraInicio string `json:"hora_inicio"`
	HoraFin    string `json:"hora_fin"`
	Aula       string `json:"aula"`
}

// Comision es un curso de una materia (K1.1, K3.2, ...) con sus horarios.
type Comision struct {
	ID           int64             `json:"id"`
	Codigo       string            `json:"codigo"`
	Cuatrimestre string            `json:"cuatrimestre"`
	Horarios     []HorarioComision `json:"horarios"`
}

type MateriaRepository interface {
	GetAll(ctx context.Context) ([]Materia, error)
	GetByCarrera(ctx context.Context, carreraID int64) ([]Materia, error)
	GetByID(ctx context.Context, id int64) (*Materia, error)
	GetComisiones(ctx context.Context, materiaID int64) ([]Comision, error)
}

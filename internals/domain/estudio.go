package domain

import (
	"context"
	"errors"
	"slices"
	"strings"
	"unicode/utf8"
)

var (
	ErrSesionNoEncontrada = errors.New("sesion de estudio no encontrada")
	ErrTareaNoEncontrada  = errors.New("tarea de estudio no encontrada")
)

// ModosEstudio son las formas de medir una sesion: pomodoro (bloques fijos)
// o libre (cronometro).
var ModosEstudio = []string{"pomodoro", "libre"}

// Una sesion dura entre 1 minuto y 12 horas
const (
	SesionMinutosMin = 1
	SesionMinutosMax = 720
	// La meta diaria va de 15 minutos a 12 horas
	MetaMinutosMin = 15
	MetaMinutosMax = 720
	TituloTareaMax = 200
	// ZonaHoraria define en que dia cae cada sesion
	ZonaHoraria = "America/Argentina/Buenos_Aires"
)

// SesionEstudio es un rato de estudio que el alumno termino. Fin es cuando
// termino, en RFC 3339.
type SesionEstudio struct {
	ID      int64           `json:"id"`
	Modo    string          `json:"modo"`
	Minutos int             `json:"minutos"`
	Fin     string          `json:"fin"`
	Materia *MateriaResumen `json:"materia"`
}

type DatosSesion struct {
	MateriaID *int64 `json:"materia_id"`
	Modo      string `json:"modo"`
	Minutos   int    `json:"minutos"`
}

func (d *DatosSesion) Normalizar() {
	d.MateriaID = normalizarMateria(d.MateriaID)
}

func (d *DatosSesion) Validate() map[string]string {
	errs := make(map[string]string)
	if !slices.Contains(ModosEstudio, d.Modo) {
		errs["modo"] = "El modo tiene que ser pomodoro o libre"
	}
	if d.Minutos < SesionMinutosMin || d.Minutos > SesionMinutosMax {
		errs["minutos"] = "La sesión tiene que durar entre 1 minuto y 12 horas"
	}
	return errs
}

// MinutosPorDia son los minutos estudiados en un dia (AAAA-MM-DD).
type MinutosPorDia struct {
	Fecha   string `json:"fecha" bun:"fecha"`
	Minutos int    `json:"minutos" bun:"minutos"`
}

// MinutosPorMateria agrupa lo estudiado por materia; sin materia va con ID nil.
type MinutosPorMateria struct {
	MateriaID *int64 `json:"materia_id" bun:"materia_id"`
	Nombre    string `json:"nombre" bun:"nombre"`
	Minutos   int    `json:"minutos" bun:"minutos"`
}

// PuestoRanking es un alumno en el ranking de la semana. Solo se muestra el
// nombre de pila y la inicial del apellido.
type PuestoRanking struct {
	Posicion int    `json:"posicion"`
	Nombre   string `json:"nombre"`
	Minutos  int    `json:"minutos"`
	SoyYo    bool   `json:"soy_yo"`
}

type FilaRanking struct {
	AlumnoID int64  `bun:"alumno_id"`
	Nombre   string `bun:"nombre"`
	Minutos  int    `bun:"minutos"`
}

// TareaEstudio es algo que el alumno anota para estudiar, opcionalmente de una materia.
type TareaEstudio struct {
	ID      int64           `json:"id"`
	Titulo  string          `json:"titulo"`
	Hecha   bool            `json:"hecha"`
	Creada  string          `json:"creada"`
	Materia *MateriaResumen `json:"materia"`
}

type DatosTarea struct {
	Titulo    string `json:"titulo"`
	MateriaID *int64 `json:"materia_id"`
	Hecha     bool   `json:"hecha"`
}

func (d *DatosTarea) Normalizar() {
	d.Titulo = strings.TrimSpace(d.Titulo)
	d.MateriaID = normalizarMateria(d.MateriaID)
}

func (d *DatosTarea) Validate() map[string]string {
	errs := make(map[string]string)
	if d.Titulo == "" {
		errs["titulo"] = "Escribí qué tenés que hacer"
	} else if utf8.RuneCountInString(d.Titulo) > TituloTareaMax {
		errs["titulo"] = "Máximo 200 caracteres"
	}
	return errs
}

type EstudioRepository interface {
	CreateSesion(ctx context.Context, alumnoID int64, datos DatosSesion) (int64, error)
	GetSesion(ctx context.Context, alumnoID, id int64) (*SesionEstudio, error)
	ListSesiones(ctx context.Context, alumnoID int64, limite int) ([]SesionEstudio, error)
	DeleteSesion(ctx context.Context, alumnoID, id int64) error

	// Desde y hasta son dias AAAA-MM-DD en la zona horaria de Argentina, inclusive
	MinutosPorDia(ctx context.Context, alumnoID int64, desde, hasta string) ([]MinutosPorDia, error)
	MinutosPorMateria(ctx context.Context, alumnoID int64, desde, hasta string) ([]MinutosPorMateria, error)

	EnRanking(ctx context.Context, alumnoID int64) (bool, error)
	SetEnRanking(ctx context.Context, alumnoID int64, participar bool) error
	// Ranking devuelve a todos los que participan con sus minutos del rango,
	// de mas a menos.
	Ranking(ctx context.Context, desde, hasta string) ([]FilaRanking, error)

	MetaDiaria(ctx context.Context, alumnoID int64) (int, error)
	SetMetaDiaria(ctx context.Context, alumnoID int64, minutos int) error

	ListTareas(ctx context.Context, alumnoID int64) ([]TareaEstudio, error)
	GetTarea(ctx context.Context, alumnoID, id int64) (*TareaEstudio, error)
	CreateTarea(ctx context.Context, alumnoID int64, datos DatosTarea) (int64, error)
	UpdateTarea(ctx context.Context, alumnoID, id int64, datos DatosTarea) error
	DeleteTarea(ctx context.Context, alumnoID, id int64) error
}

package domain

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rivo/uniseg"
)

var (
	ErrEventoNoEncontrado = errors.New("evento no encontrado")
	ErrClaseNoEncontrada  = errors.New("clase no encontrada")
)

const (
	FormatoFecha = "2006-01-02"
	FormatoHora  = "15:04"

	calendarioTituloMax      = 120
	eventoDescripcionMax     = 1000
	claseAulaMax             = 60
	EventosRangoMaximoEnDias = 400
)

// TiposEvento son los tipos de evento que acepta el calendario.
var TiposEvento = []string{"parcial", "final", "entrega", "recordatorio", "otro"}

type MateriaResumen struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

// Evento es una fecha puntual en el calendario del alumno.
type Evento struct {
	ID          int64           `json:"id"`
	Titulo      string          `json:"titulo"`
	Tipo        string          `json:"tipo"`
	Fecha       string          `json:"fecha"`
	Hora        *string         `json:"hora"`
	Descripcion string          `json:"descripcion"`
	Materia     *MateriaResumen `json:"materia"`
}

// DatosEvento son los campos que el alumno completa al crear o editar un evento.
type DatosEvento struct {
	Titulo      string  `json:"titulo"`
	Tipo        string  `json:"tipo"`
	Fecha       string  `json:"fecha"`
	Hora        *string `json:"hora"`
	Descripcion string  `json:"descripcion"`
	MateriaID   *int64  `json:"materia_id"`
}

func normalizarOpcional(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func normalizarMateria(id *int64) *int64 {
	if id == nil || *id <= 0 {
		return nil
	}
	return id
}

func validarTitulo(titulo string, errs map[string]string) {
	largo := uniseg.GraphemeClusterCount(titulo)
	if largo == 0 {
		errs["titulo"] = "El título es requerido"
	} else if largo > calendarioTituloMax {
		errs["titulo"] = "El título no puede tener más de 120 caracteres"
	}
}

func esHora(s string) bool {
	_, err := time.Parse(FormatoHora, s)
	return err == nil && len(s) == len(FormatoHora)
}

func (d *DatosEvento) Normalizar() {
	d.Titulo = strings.TrimSpace(d.Titulo)
	d.Tipo = strings.ToLower(strings.TrimSpace(d.Tipo))
	d.Fecha = strings.TrimSpace(d.Fecha)
	d.Descripcion = strings.TrimSpace(d.Descripcion)
	d.Hora = normalizarOpcional(d.Hora)
	d.MateriaID = normalizarMateria(d.MateriaID)
	if d.Tipo == "" {
		d.Tipo = "otro"
	}
}

func (d *DatosEvento) Validate() map[string]string {
	errs := make(map[string]string)

	validarTitulo(d.Titulo, errs)

	tipoValido := false
	for _, t := range TiposEvento {
		if d.Tipo == t {
			tipoValido = true
			break
		}
	}
	if !tipoValido {
		errs["tipo"] = "Tipo inválido. Puede ser parcial, final, entrega, recordatorio u otro"
	}

	if _, err := time.Parse(FormatoFecha, d.Fecha); err != nil {
		errs["fecha"] = "La fecha tiene que tener el formato AAAA-MM-DD"
	}

	if d.Hora != nil && !esHora(*d.Hora) {
		errs["hora"] = "La hora tiene que tener el formato HH:MM"
	}

	if uniseg.GraphemeClusterCount(d.Descripcion) > eventoDescripcionMax {
		errs["descripcion"] = "La descripción no puede tener más de 1000 caracteres"
	}

	return errs
}

// Clase es un bloque del horario semanal de cursada. Dia va de 1 (lunes) a 7 (domingo).
type Clase struct {
	ID         int64           `json:"id"`
	Titulo     string          `json:"titulo"`
	Dia        int             `json:"dia"`
	HoraInicio string          `json:"hora_inicio"`
	HoraFin    string          `json:"hora_fin"`
	Aula       string          `json:"aula"`
	Materia    *MateriaResumen `json:"materia"`
	// Comision de la que se cargo la clase; nil si se cargo a mano
	ComisionID *int64 `json:"comision_id"`
	// Cuatrimestre de esa comision (1C, 2C); nil si se cargo a mano
	Cuatrimestre *string `json:"cuatrimestre"`
}

type DatosClase struct {
	Titulo     string `json:"titulo"`
	Dia        int    `json:"dia"`
	HoraInicio string `json:"hora_inicio"`
	HoraFin    string `json:"hora_fin"`
	Aula       string `json:"aula"`
	MateriaID  *int64 `json:"materia_id"`
	ComisionID *int64 `json:"comision_id"`
}

func (d *DatosClase) Normalizar() {
	d.Titulo = strings.TrimSpace(d.Titulo)
	d.HoraInicio = strings.TrimSpace(d.HoraInicio)
	d.HoraFin = strings.TrimSpace(d.HoraFin)
	d.Aula = strings.TrimSpace(d.Aula)
	d.MateriaID = normalizarMateria(d.MateriaID)
	d.ComisionID = normalizarMateria(d.ComisionID)
}

func (d *DatosClase) Validate() map[string]string {
	errs := make(map[string]string)

	validarTitulo(d.Titulo, errs)

	if d.Dia < 1 || d.Dia > 7 {
		errs["dia"] = "El día tiene que ir de 1 (lunes) a 7 (domingo)"
	}

	inicioOK := esHora(d.HoraInicio)
	finOK := esHora(d.HoraFin)
	if !inicioOK {
		errs["hora_inicio"] = "La hora tiene que tener el formato HH:MM"
	}
	if !finOK {
		errs["hora_fin"] = "La hora tiene que tener el formato HH:MM"
	}
	// Con formato HH:MM la comparacion de strings respeta el orden horario
	if inicioOK && finOK && d.HoraFin <= d.HoraInicio {
		errs["hora_fin"] = "Tiene que terminar después de empezar"
	}

	if uniseg.GraphemeClusterCount(d.Aula) > claseAulaMax {
		errs["aula"] = "El aula no puede tener más de 60 caracteres"
	}

	return errs
}

type CalendarioRepository interface {
	ListEventos(ctx context.Context, alumnoID int64, desde, hasta string) ([]Evento, error)
	GetEvento(ctx context.Context, alumnoID, id int64) (*Evento, error)
	CreateEvento(ctx context.Context, alumnoID int64, datos DatosEvento) (int64, error)
	UpdateEvento(ctx context.Context, alumnoID, id int64, datos DatosEvento) error
	DeleteEvento(ctx context.Context, alumnoID, id int64) error

	ListClases(ctx context.Context, alumnoID int64) ([]Clase, error)
	GetClase(ctx context.Context, alumnoID, id int64) (*Clase, error)
	CreateClase(ctx context.Context, alumnoID int64, datos DatosClase) (int64, error)
	UpdateClase(ctx context.Context, alumnoID, id int64, datos DatosClase) error
	DeleteClase(ctx context.Context, alumnoID, id int64) error
}

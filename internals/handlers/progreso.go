package handlers

import (
	"chedul-core/internals/domain"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

const (
	condicionAprobada     = "Aprobada"
	condicionRegularizada = "Regularizada"
	condicionCursando     = "Cursando"
	condicionPendiente    = "Pendiente"
)

type ProgresoHandler struct {
	alumnoRepo          domain.AlumnoRepository
	materiaRepo         domain.MateriaRepository
	condicionAlumnoRepo domain.CondicionAlumnoRepository
}

func NewProgresoHandler(alumnoRepo domain.AlumnoRepository, materiaRepo domain.MateriaRepository, condicionAlumnoRepo domain.CondicionAlumnoRepository) *ProgresoHandler {
	return &ProgresoHandler{
		alumnoRepo:          alumnoRepo,
		materiaRepo:         materiaRepo,
		condicionAlumnoRepo: condicionAlumnoRepo,
	}
}

type MateriaSimple struct {
	ID           int64  `json:"id"`
	Nombre       string `json:"nombre"`
	Nivel        int64  `json:"nivel"`
	EstadoActual string `json:"estado_actual"`
	Nota         *int   `json:"nota"`
	Mensaje      string `json:"mensaje"`
	Tipo         string `json:"tipo"`
}

type ProgresoResponse struct {
	AlumnoID                        int64           `json:"alumno_id"`
	MateriasAprobadas               []MateriaSimple `json:"materias_aprobadas"`
	MateriasRegularizadas           []MateriaSimple `json:"materias_regularizadas"`
	MateriasCursando                []MateriaSimple `json:"materias_cursando"`
	MateriasPendientesDisponibles   []MateriaSimple `json:"materias_pendientes_disponibles"`
	MateriasPendientesNoDisponibles []MateriaSimple `json:"materias_pendientes_no_disponibles"`
	ObligatoriasTotal               int             `json:"obligatorias_total"`
	ObligatoriasAprobadas           int             `json:"obligatorias_aprobadas"`
	PorcentajeAprobadas             float64         `json:"porcentaje_aprobadas"`
	Promedio                        *float64        `json:"promedio"`
}

// cumpleCorrelativa indica si la condicion del alumno en la materia requerida
// alcanza para el tipo de correlativa.
func cumpleCorrelativa(tipo, condicionRequerida string) bool {
	switch tipo {
	case domain.CorrelativaAprobada:
		return condicionRequerida == condicionAprobada
	case domain.CorrelativaRegular:
		return condicionRequerida == condicionAprobada || condicionRequerida == condicionRegularizada
	}
	return false
}

// CalcularProgreso clasifica las materias de la carrera segun la condicion del
// alumno y las correlativas de cada una.
func CalcularProgreso(alumnoID int64, materias []domain.Materia, condiciones []domain.CondicionPorAlumno) ProgresoResponse {
	resp := ProgresoResponse{
		AlumnoID:                        alumnoID,
		MateriasAprobadas:               []MateriaSimple{},
		MateriasRegularizadas:           []MateriaSimple{},
		MateriasCursando:                []MateriaSimple{},
		MateriasPendientesDisponibles:   []MateriaSimple{},
		MateriasPendientesNoDisponibles: []MateriaSimple{},
	}

	estados := make(map[int64]domain.CondicionPorAlumno, len(condiciones))
	for _, c := range condiciones {
		estados[c.MateriaID] = c
	}

	sumaNotas, cantidadNotas := 0, 0

	for _, materia := range materias {
		simple := MateriaSimple{
			ID:     materia.ID,
			Nombre: materia.Nombre,
			Nivel:  materia.Nivel,
			Tipo:   materia.Tipo,
		}

		if materia.Tipo == "Obligatoria" {
			resp.ObligatoriasTotal++
		}

		if estado, ok := estados[materia.ID]; ok {
			simple.EstadoActual = estado.Condicion
			simple.Nota = estado.Nota

			switch estado.Condicion {
			case condicionAprobada:
				simple.Mensaje = "Ya aprobada"
				resp.MateriasAprobadas = append(resp.MateriasAprobadas, simple)
				if materia.Tipo == "Obligatoria" {
					resp.ObligatoriasAprobadas++
				}
				if estado.Nota != nil {
					sumaNotas += *estado.Nota
					cantidadNotas++
				}
			case condicionRegularizada:
				simple.Mensaje = "Regularizada, falta el final"
				resp.MateriasRegularizadas = append(resp.MateriasRegularizadas, simple)
			case condicionCursando:
				simple.Mensaje = "Actualmente cursando"
				resp.MateriasCursando = append(resp.MateriasCursando, simple)
			}
			continue
		}

		simple.EstadoActual = condicionPendiente

		var faltantes []string
		for _, correlativa := range materia.Correlativas {
			requerida := estados[correlativa.MateriaID].Condicion
			if cumpleCorrelativa(correlativa.Tipo, requerida) {
				continue
			}

			if correlativa.Tipo == domain.CorrelativaAprobada {
				faltantes = append(faltantes, fmt.Sprintf("Necesita aprobar %s", correlativa.Nombre))
			} else {
				faltantes = append(faltantes, fmt.Sprintf("Necesita regularizar %s", correlativa.Nombre))
			}
		}

		if len(faltantes) == 0 {
			simple.Mensaje = "Disponible para cursar"
			resp.MateriasPendientesDisponibles = append(resp.MateriasPendientesDisponibles, simple)
		} else {
			simple.Mensaje = strings.Join(faltantes, ". ")
			resp.MateriasPendientesNoDisponibles = append(resp.MateriasPendientesNoDisponibles, simple)
		}
	}

	if resp.ObligatoriasTotal > 0 {
		resp.PorcentajeAprobadas = float64(resp.ObligatoriasAprobadas) / float64(resp.ObligatoriasTotal) * 100
	}

	if cantidadNotas > 0 {
		promedio := float64(sumaNotas) / float64(cantidadNotas)
		resp.Promedio = &promedio
	}

	return resp
}

// GetMiProgreso devuelve el progreso del alumno autenticado en su carrera.
func (h *ProgresoHandler) GetMiProgreso(c echo.Context) error {
	ctx := c.Request().Context()

	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	alumno, err := h.alumnoRepo.GetByID(ctx, alumnoID)
	if err != nil {
		return err
	}

	materias, err := h.materiaRepo.GetByCarrera(ctx, alumno.Carrera)
	if err != nil {
		return err
	}

	condiciones, err := h.condicionAlumnoRepo.GetCondicionPorAlumno(ctx, alumnoID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, CalcularProgreso(alumnoID, materias, condiciones))
}

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
	// Solo para electivas: el alumno no la va a cursar y no aparece en ninguna lista
	condicionNoMeInteresa = "No me interesa"
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
	Ordenanza531                    Ordenanza531    `json:"ordenanza_531"`
}

// Ordenanza531 dice si el alumno puede pedir la excepcion de correlativas del
// punto 5.3.1 del Reglamento de Estudios (Ord. 1549, modificada por la 1872):
// si la carga horaria semanal de las materias que le faltan cursar no supera
// la del ultimo nivel, las cursa sin correlativas (para rendir el final si se
// piden). Se cuentan horas semanales (carga_horaria), no las horas totales.
// Las electivas y las que no tienen cursado (la Practica Supervisada) no se
// cuentan, porque no se sabe cuales va a elegir el alumno.
type Ordenanza531 struct {
	Puede          bool            `json:"puede"`
	HorasFaltantes float64         `json:"horas_faltantes"`
	HorasLimite    float64         `json:"horas_limite"`
	Faltantes      []MateriaSimple `json:"faltantes"`
}

const (
	tipoObligatoria = "Obligatoria"
	tipoElectiva    = "Electiva"
)

// cuentaPara531 indica si la materia entra en la cuenta: obligatoria y con cursado.
func cuentaPara531(m domain.Materia) bool {
	return m.Tipo == tipoObligatoria && m.CargaHoraria > 0
}

// CalcularOrdenanza531 compara las horas de las obligatorias sin aprobar con
// las del ultimo nivel de la carrera.
func CalcularOrdenanza531(materias []domain.Materia, estados map[int64]domain.CondicionPorAlumno) Ordenanza531 {
	resp := Ordenanza531{Faltantes: []MateriaSimple{}}

	var ultimoNivel int64
	for _, m := range materias {
		if cuentaPara531(m) && m.Nivel > ultimoNivel {
			ultimoNivel = m.Nivel
		}
	}

	for _, m := range materias {
		if !cuentaPara531(m) {
			continue
		}
		if m.Nivel == ultimoNivel {
			resp.HorasLimite += float64(m.CargaHoraria)
		}
		// Solo cuenta lo que falta cursar: lo aprobado, regularizado o que
		// ya se esta cursando no necesita la excepcion
		estado := estados[m.ID]
		if estado.Condicion == condicionAprobada || estado.Condicion == condicionRegularizada || estado.Condicion == condicionCursando {
			continue
		}
		resp.HorasFaltantes += float64(m.CargaHoraria)
		condicion := estado.Condicion
		if condicion == "" {
			condicion = condicionPendiente
		}
		resp.Faltantes = append(resp.Faltantes, MateriaSimple{
			ID:           m.ID,
			Nombre:       m.Nombre,
			Nivel:        m.Nivel,
			EstadoActual: condicion,
			Tipo:         m.Tipo,
		})
	}

	resp.Puede = resp.HorasLimite > 0 && resp.HorasFaltantes > 0 && resp.HorasFaltantes <= resp.HorasLimite
	return resp
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

		if materia.Tipo == tipoObligatoria {
			resp.ObligatoriasTotal++
		}

		if estado, ok := estados[materia.ID]; ok {
			simple.EstadoActual = estado.Condicion
			simple.Nota = estado.Nota

			switch estado.Condicion {
			case condicionAprobada:
				simple.Mensaje = "Ya aprobada"
				resp.MateriasAprobadas = append(resp.MateriasAprobadas, simple)
				if materia.Tipo == tipoObligatoria {
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
			case condicionNoMeInteresa:
				// Electiva descartada: no va ni en pendientes ni en "podés cursar"
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

	resp.Ordenanza531 = CalcularOrdenanza531(materias, estados)

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

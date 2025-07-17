package handlers

import (
	"chedul-core/internals/domain"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/uptrace/bun"
)

// TODO: Ver esto
type ProgresoHandler struct {
	db                  *bun.DB
	alumnoRepo          domain.AlumnoRepository
	condicionAlumnoRepo domain.CondicionAlumnoRepository
	condicionRepo       domain.CondicionRepository
}

func NewProgresoHandler(alumnoRepo domain.AlumnoRepository, condicionAlumnorepo domain.CondicionAlumnoRepository, condicionRepo domain.CondicionRepository, db *bun.DB) *ProgresoHandler {
	return &ProgresoHandler{
		db:                  db,
		alumnoRepo:          alumnoRepo,
		condicionAlumnoRepo: condicionAlumnorepo,
		condicionRepo:       condicionRepo,
	}
}

type MateriaSimple struct {
	ID           int64  `json:"id"`
	Nombre       string `json:"nombre"`
	Nivel        int64  `json:"nivel"`
	EstadoActual string `json:"estado_actual"`
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
	// Se pueden añadir más estadísticas como las electivas si es necesario.
}

func (h *ProgresoHandler) GetProgresoAlumno(c echo.Context) error {
	ctx := c.Request().Context()

	// 1. Obtener el ID del alumno desde la URL
	alumnoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de alumno inválido"))
	}

	// 2. Obtener la carrera del alumno
	alumno, err := h.alumnoRepo.GetByID(ctx, alumnoID)
	if err != nil {
		return err
	}

	// 3. Obtener TODAS las materias de la carrera del alumno
	var materiasDeLaCarrera []domain.Materia
	err = h.db.NewSelect().
		Model(&materiasDeLaCarrera).
		Join("JOIN materiasporcarrera AS mpc ON mpc.materia_id = materia.id").
		Where("mpc.carrera_id = ?", alumno.Carrera).
		Scan(ctx)

	if err != nil {
		return err
	}

	// 4. Obtener TODOS los estados que el alumno ya tiene para esas materias
	// TODO: CondicionPorAlumno o CondicionAlumno?
	condicionDelAlumno, err := h.condicionAlumnoRepo.GetCondicionPorAlumno(ctx, alumnoID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// 5. Pre-cargar datos para evitar consultas en bucle (N+1)
	// Mapa de todas las condiciones (ID -> Nombre)
	condiciones, err := h.condicionRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	condicionesMap := make(map[int64]string)
	for _, c := range condiciones {
		condicionesMap[c.ID] = c.Condicion
	}

	type EstadoInfo struct {
		Nombre string
		Nota   *int
	}

	estadosMap := make(map[int64]EstadoInfo)
	for _, estado := range condicionDelAlumno {
		estadosMap[estado.Materia.ID] = EstadoInfo{
			Nombre: condicionesMap[estado.Condicion.ID],
			Nota:   estado.Nota,
		}
	}

	// Mapa de nombres de materias para mensajes de correlativas
	nombresMateriasMap := make(map[int64]string)
	for _, m := range materiasDeLaCarrera {
		nombresMateriasMap[m.ID] = m.Nombre
	}

	// 6. Clasificar cada materia
	var resp ProgresoResponse
	resp.AlumnoID = alumnoID

	for _, materia := range materiasDeLaCarrera {
		materiaSimple := MateriaSimple{
			ID:     materia.ID,
			Nombre: materia.Nombre,
			Nivel:  materia.Nivel,
			Tipo:   materia.Tipo,
		}

		if estado, ok := estadosMap[materia.ID]; ok {
			// El alumno tiene un estado para esta materia
			materiaSimple.EstadoActual = estado.Nombre
			switch estado.Nombre {
			case "Aprobada":
				materiaSimple.Mensaje = "Ya aprobada"
				resp.MateriasAprobadas = append(resp.MateriasAprobadas, materiaSimple)
			case "Regularizada":
				materiaSimple.Mensaje = "Regularizada"
				resp.MateriasRegularizadas = append(resp.MateriasRegularizadas, materiaSimple)
			case "Cursando":
				materiaSimple.Mensaje = "Actualmente cursando"
				resp.MateriasCursando = append(resp.MateriasCursando, materiaSimple)
			}
		} else {
			// El alumno no tiene estado para esta materia (está pendiente)
			materiaSimple.EstadoActual = "Pendiente"
			puedeCursar := true
			mensaje := "Disponible para cursar"

			if materia.CorrelativaID != nil {
				if estadoCorrelativa, correlativaExiste := estadosMap[*materia.CorrelativaID]; !correlativaExiste || estadoCorrelativa.Nombre != "Aprobada" {
					puedeCursar = false
					nombreCorrelativa := nombresMateriasMap[*materia.CorrelativaID]
					mensaje = fmt.Sprintf("Necesita aprobar %s", nombreCorrelativa)
				}
			}

			materiaSimple.Mensaje = mensaje
			if puedeCursar {
				resp.MateriasPendientesDisponibles = append(resp.MateriasPendientesDisponibles, materiaSimple)
			} else {
				resp.MateriasPendientesNoDisponibles = append(resp.MateriasPendientesNoDisponibles, materiaSimple)
			}
		}
	}

	// 7. Calcular estadísticas
	for _, m := range materiasDeLaCarrera {
		if m.Tipo == "Obligatoria" {
			resp.ObligatoriasTotal++
		}
	}
	// Contar solo las aprobadas que son obligatorias
	for _, aprobada := range resp.MateriasAprobadas {
		if aprobada.Tipo == "Obligatoria" {
			resp.ObligatoriasAprobadas++
		}
	}

	if resp.ObligatoriasTotal > 0 {
		resp.PorcentajeAprobadas = (float64(resp.ObligatoriasAprobadas) / float64(resp.ObligatoriasTotal)) * 100
	}

	return c.JSON(http.StatusOK, resp)
}

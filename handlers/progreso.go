package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// MateriaSimple es una estructura simplificada para las respuestas.
type MateriaSimple struct {
	ID           int64  `json:"id"`
	Nombre       string `json:"nombre"`
	Nivel        int    `json:"nivel"`
	EstadoActual string `json:"estado_actual"`
	Mensaje      string `json:"mensaje"`
	Tipo         string `json:"tipo"`
}

// ProgresoResponse es la estructura completa que se enviará como JSON.
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

// HandleGetProgresoAlumno calcula y devuelve el estado de todas las materias para un alumno.
func HandleGetProgresoAlumno(c echo.Context) error {
	ctx := c.(*AppContext)
	conn := ctx.conn

	// 1. Obtener el ID del alumno desde la URL
	alumnoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return NewApiError(http.StatusBadRequest, fmt.Errorf("ID de alumno inválido"))
	}

	// 2. Obtener la carrera del alumno
	var alumno Alumno
	if err := conn.NewSelect().Model(&alumno).Where("id = ?", alumnoID).Scan(ctx.Request().Context()); err != nil {
		return NewApiError(http.StatusNotFound, fmt.Errorf("alumno no encontrado"))
	}

	// 3. Obtener TODAS las materias de la carrera del alumno
	var materiasDeLaCarrera []Materia
	err = conn.NewSelect().
		Model(&materiasDeLaCarrera).
		Join("JOIN materiasporcarrera AS mpc ON mpc.materia_id = materia.id").
		Where("mpc.carrera_id = ?", alumno.Carrera).
		Scan(ctx.Request().Context())
	if err != nil {
		return err
	}

	// 4. Obtener TODOS los estados que el alumno ya tiene para esas materias
	var estadosDelAlumno []CondicionAlumno
	err = conn.NewSelect().Model(&estadosDelAlumno).Where("alumno_id = ?", alumnoID).Scan(ctx.Request().Context())
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// 5. Pre-cargar datos para evitar consultas en bucle (N+1)
	// Mapa de todas las condiciones (ID -> Nombre)
	var condiciones []Condicion
	_ = conn.NewSelect().Model(&condiciones).Scan(ctx.Request().Context())
	condicionesMap := make(map[int64]string)
	for _, c := range condiciones {
		condicionesMap[c.ID] = c.Condicion
	}

	// Mapa de los estados del alumno (MateriaID -> Estado)
	type EstadoInfo struct {
		Nombre string
		Nota   *int
	}
	estadosMap := make(map[int64]EstadoInfo)
	for _, estado := range estadosDelAlumno {
		estadosMap[estado.MateriaID] = EstadoInfo{
			Nombre: condicionesMap[estado.CondicionID],
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

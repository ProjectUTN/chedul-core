package handlers

import (
	"chedul-core/internals/domain"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	sesionesRecientes = 20
	// 26 semanas: alcanza para el calendario de actividad y la cotizacion
	diasDelGrafico    = 182
	diasDeRacha       = 366
	puestosDelRanking = 20
)

type EstudioHandler struct {
	repo        domain.EstudioRepository
	materiaRepo domain.MateriaRepository
	logger      *zap.Logger
	// ahora se puede cambiar en los tests
	ahora func() time.Time
}

func NewEstudioHandler(repo domain.EstudioRepository, materiaRepo domain.MateriaRepository, logger *zap.Logger) *EstudioHandler {
	return &EstudioHandler{repo: repo, materiaRepo: materiaRepo, logger: logger, ahora: time.Now}
}

// hoy es la fecha de hoy en Argentina, que define en que dia cae cada sesion
func (h *EstudioHandler) hoy() time.Time {
	zona, err := time.LoadLocation(domain.ZonaHoraria)
	if err != nil {
		zona = time.FixedZone("ART", -3*60*60)
	}
	a := h.ahora().In(zona)
	return time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, time.UTC)
}

// lunesDe devuelve el lunes de la semana de la fecha
func lunesDe(fecha time.Time) time.Time {
	return fecha.AddDate(0, 0, -((int(fecha.Weekday()) + 6) % 7))
}

func dia(fecha time.Time) string {
	return fecha.Format(domain.FormatoFecha)
}

func (h *EstudioHandler) CreateSesion(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	var datos domain.DatosSesion
	if err := c.Bind(&datos); err != nil {
		return InvalidJSON()
	}
	datos.Normalizar()
	errs := datos.Validate()
	if datos.MateriaID != nil {
		if _, err := h.materiaRepo.GetByID(c.Request().Context(), *datos.MateriaID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			errs["materia_id"] = "La materia no existe"
		}
	}
	if len(errs) > 0 {
		return InvalidRequestData(errs)
	}

	ctx := c.Request().Context()
	id, err := h.repo.CreateSesion(ctx, alumnoID, datos)
	if err != nil {
		return err
	}
	sesion, err := h.repo.GetSesion(ctx, alumnoID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, sesion)
}

func (h *EstudioHandler) ListSesiones(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	sesiones, err := h.repo.ListSesiones(c.Request().Context(), alumnoID, sesionesRecientes)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sesiones)
}

// UpdateSesion cambia la materia de una sesion; el resto (cuando empezo,
// cuando termino, cuanto duro) queda como se registro.
func (h *EstudioHandler) UpdateSesion(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}
	var datos domain.DatosMateriaSesion
	if err := c.Bind(&datos); err != nil {
		return InvalidJSON()
	}
	datos.Normalizar()
	ctx := c.Request().Context()
	if datos.MateriaID != nil {
		if _, err := h.materiaRepo.GetByID(ctx, *datos.MateriaID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			return InvalidRequestData(map[string]string{"materia_id": "La materia no existe"})
		}
	}
	err = h.repo.SetMateriaSesion(ctx, alumnoID, id, datos.MateriaID)
	if errors.Is(err, domain.ErrSesionNoEncontrada) {
		return NotFound("Sesión")
	}
	if err != nil {
		return err
	}
	sesion, err := h.repo.GetSesion(ctx, alumnoID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, sesion)
}

func (h *EstudioHandler) DeleteSesion(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}
	err = h.repo.DeleteSesion(c.Request().Context(), alumnoID, id)
	if errors.Is(err, domain.ErrSesionNoEncontrada) {
		return NotFound("Sesión")
	}
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type ResumenEstudio struct {
	HoyMinutos    int `json:"hoy_minutos"`
	SemanaMinutos int `json:"semana_minutos"`
	MesMinutos    int `json:"mes_minutos"`
	MetaDiaria    int `json:"meta_diaria"`
	// Dias seguidos estudiando hasta hoy; si hoy todavia no estudio, hasta ayer
	RachaDias int `json:"racha_dias"`
	// Los ultimos 182 dias, incluidos los que no estudio
	PorDia []domain.MinutosPorDia `json:"por_dia"`
	// Lo de esta semana (de lunes a hoy)
	PorMateria []domain.MinutosPorMateria `json:"por_materia"`
}

// CalcularRacha cuenta los dias seguidos con estudio que terminan hoy, o
// ayer si hoy todavia no estudio.
func CalcularRacha(dias []domain.MinutosPorDia, hoy time.Time) int {
	estudiados := make(map[string]bool, len(dias))
	for _, d := range dias {
		if d.Minutos > 0 {
			estudiados[d.Fecha] = true
		}
	}
	fecha := hoy
	if !estudiados[dia(fecha)] {
		fecha = fecha.AddDate(0, 0, -1)
	}
	racha := 0
	for estudiados[dia(fecha)] {
		racha++
		fecha = fecha.AddDate(0, 0, -1)
	}
	return racha
}

func (h *EstudioHandler) Resumen(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	hoy := h.hoy()
	lunes := lunesDe(hoy)

	dias, err := h.repo.MinutosPorDia(ctx, alumnoID, dia(hoy.AddDate(0, 0, -diasDeRacha)), dia(hoy))
	if err != nil {
		return err
	}
	porMateria, err := h.repo.MinutosPorMateria(ctx, alumnoID, dia(lunes), dia(hoy))
	if err != nil {
		return err
	}

	meta, err := h.repo.MetaDiaria(ctx, alumnoID)
	if err != nil {
		return err
	}

	minutos := make(map[string]int, len(dias))
	for _, d := range dias {
		minutos[d.Fecha] = d.Minutos
	}

	resp := ResumenEstudio{
		HoyMinutos: minutos[dia(hoy)],
		MetaDiaria: meta,
		RachaDias:  CalcularRacha(dias, hoy),
		PorDia:     make([]domain.MinutosPorDia, 0, diasDelGrafico),
		PorMateria: porMateria,
	}
	for f := lunes; !f.After(hoy); f = f.AddDate(0, 0, 1) {
		resp.SemanaMinutos += minutos[dia(f)]
	}
	primeroDelMes := time.Date(hoy.Year(), hoy.Month(), 1, 0, 0, 0, 0, time.UTC)
	for f := primeroDelMes; !f.After(hoy); f = f.AddDate(0, 0, 1) {
		resp.MesMinutos += minutos[dia(f)]
	}
	for i := diasDelGrafico - 1; i >= 0; i-- {
		f := dia(hoy.AddDate(0, 0, -i))
		resp.PorDia = append(resp.PorDia, domain.MinutosPorDia{Fecha: f, Minutos: minutos[f]})
	}
	return c.JSON(http.StatusOK, resp)
}

type RankingResponse struct {
	Participo     bool                   `json:"participo"`
	Desde         string                 `json:"desde"`
	Hasta         string                 `json:"hasta"`
	Participantes int                    `json:"participantes"`
	Puestos       []domain.PuestoRanking `json:"puestos"`
}

// ArmarRanking numera a los participantes (los empatados comparten puesto) y
// deja los primeros, mas el alumno si quedo mas abajo.
func ArmarRanking(filas []domain.FilaRanking, alumnoID int64, limite int) []domain.PuestoRanking {
	puestos := []domain.PuestoRanking{}
	posicion := 0
	for i, f := range filas {
		if i == 0 || f.Minutos != filas[i-1].Minutos {
			posicion = i + 1
		}
		soyYo := f.AlumnoID == alumnoID
		if i >= limite && !soyYo {
			continue
		}
		puestos = append(puestos, domain.PuestoRanking{
			Posicion: posicion,
			Nombre:   domain.NombreCorto(f.Nombre, f.Apellido),
			Minutos:  f.Minutos,
			SoyYo:    soyYo,
		})
	}
	return puestos
}

// Ranking es el de la semana actual, de lunes a domingo
func (h *EstudioHandler) Ranking(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	lunes := lunesDe(h.hoy())
	domingo := lunes.AddDate(0, 0, 6)

	participo, err := h.repo.EnRanking(ctx, alumnoID)
	if err != nil {
		return err
	}
	filas, err := h.repo.Ranking(ctx, dia(lunes), dia(domingo))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, RankingResponse{
		Participo:     participo,
		Desde:         dia(lunes),
		Hasta:         dia(domingo),
		Participantes: len(filas),
		Puestos:       ArmarRanking(filas, alumnoID, puestosDelRanking),
	})
}

func (h *EstudioHandler) SetParticipacion(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	var datos struct {
		Participar *bool `json:"participar"`
	}
	if err := c.Bind(&datos); err != nil {
		return InvalidJSON()
	}
	if datos.Participar == nil {
		return InvalidRequestData(map[string]string{"participar": "Tiene que ser true o false"})
	}
	if err := h.repo.SetEnRanking(c.Request().Context(), alumnoID, *datos.Participar); err != nil {
		return err
	}
	return h.Ranking(c)
}

func (h *EstudioHandler) SetMeta(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	var datos struct {
		Minutos int `json:"minutos"`
	}
	if err := c.Bind(&datos); err != nil {
		return InvalidJSON()
	}
	if datos.Minutos < domain.MetaMinutosMin || datos.Minutos > domain.MetaMinutosMax {
		return InvalidRequestData(map[string]string{"minutos": "La meta tiene que ser de 15 minutos a 12 horas"})
	}
	if err := h.repo.SetMetaDiaria(c.Request().Context(), alumnoID, datos.Minutos); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]int{"meta_diaria": datos.Minutos})
}

func (h *EstudioHandler) ListTareas(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	tareas, err := h.repo.ListTareas(c.Request().Context(), alumnoID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, tareas)
}

// leerTarea valida el cuerpo de una tarea y que la materia exista
func (h *EstudioHandler) leerTarea(c echo.Context) (domain.DatosTarea, error) {
	var datos domain.DatosTarea
	if err := c.Bind(&datos); err != nil {
		return datos, InvalidJSON()
	}
	datos.Normalizar()
	errs := datos.Validate()
	if datos.MateriaID != nil {
		if _, err := h.materiaRepo.GetByID(c.Request().Context(), *datos.MateriaID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return datos, err
			}
			errs["materia_id"] = "La materia no existe"
		}
	}
	if len(errs) > 0 {
		return datos, InvalidRequestData(errs)
	}
	return datos, nil
}

func (h *EstudioHandler) CreateTarea(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	datos, err := h.leerTarea(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	id, err := h.repo.CreateTarea(ctx, alumnoID, datos)
	if err != nil {
		return err
	}
	tarea, err := h.repo.GetTarea(ctx, alumnoID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, tarea)
}

func (h *EstudioHandler) UpdateTarea(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}
	datos, err := h.leerTarea(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	err = h.repo.UpdateTarea(ctx, alumnoID, id, datos)
	if errors.Is(err, domain.ErrTareaNoEncontrada) {
		return NewApiError(http.StatusNotFound, errors.New("Tarea no encontrada"))
	}
	if err != nil {
		return err
	}
	tarea, err := h.repo.GetTarea(ctx, alumnoID, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, tarea)
}

func (h *EstudioHandler) DeleteTarea(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	id, err := ParamID(c, "id")
	if err != nil {
		return err
	}
	err = h.repo.DeleteTarea(c.Request().Context(), alumnoID, id)
	if errors.Is(err, domain.ErrTareaNoEncontrada) {
		return NewApiError(http.StatusNotFound, errors.New("Tarea no encontrada"))
	}
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// respuestaTemporizador lleva la hora del servidor para que cada dispositivo
// corrija su reloj: los tiempos del cronometro son hora del servidor.
type respuestaTemporizador struct {
	Estado *domain.EstadoTemporizador `json:"estado"`
	Rev    int                        `json:"rev"`
	Ahora  int64                      `json:"ahora"`
}

func (h *EstudioHandler) respuestaTemporizador(t *domain.Temporizador) respuestaTemporizador {
	r := respuestaTemporizador{Ahora: h.ahora().UnixMilli()}
	if t != nil {
		r.Estado = &t.Estado
		r.Rev = t.Rev
	}
	return r
}

// GetTemporizador devuelve el cronometro en curso, para retomarlo desde
// cualquier dispositivo.
func (h *EstudioHandler) GetTemporizador(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	t, err := h.repo.GetTemporizador(c.Request().Context(), alumnoID)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, h.respuestaTemporizador(t))
}

// SetTemporizador guarda el cronometro. Si otro dispositivo lo cambio antes
// (rev distinta) responde 409 con lo que hay guardado, y no escribe nada.
func (h *EstudioHandler) SetTemporizador(c echo.Context) error {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}
	var datos struct {
		Estado domain.EstadoTemporizador `json:"estado"`
		Rev    int                       `json:"rev"`
	}
	if err := c.Bind(&datos); err != nil {
		return InvalidJSON()
	}
	if errs := datos.Estado.Validate(); len(errs) > 0 {
		return InvalidRequestData(errs)
	}
	t, guardado, err := h.repo.GuardarTemporizador(c.Request().Context(), alumnoID, datos.Estado, datos.Rev)
	if err != nil {
		return err
	}
	if !guardado {
		return c.JSON(http.StatusConflict, h.respuestaTemporizador(t))
	}
	return c.JSON(http.StatusOK, h.respuestaTemporizador(t))
}

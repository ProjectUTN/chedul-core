package handlers

import (
	"chedul-core/internals/domain"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

const (
	sesionesRecientes = 20
	diasDelGrafico    = 28
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
	// Dias seguidos estudiando hasta hoy; si hoy todavia no estudio, hasta ayer
	RachaDias int `json:"racha_dias"`
	// Los ultimos 28 dias, incluidos los que no estudio
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

	minutos := make(map[string]int, len(dias))
	for _, d := range dias {
		minutos[d.Fecha] = d.Minutos
	}

	resp := ResumenEstudio{
		HoyMinutos: minutos[dia(hoy)],
		RachaDias:  CalcularRacha(dias, hoy),
		PorDia:     make([]domain.MinutosPorDia, 0, diasDelGrafico),
		PorMateria: porMateria,
	}
	for f := lunes; !f.After(hoy); f = f.AddDate(0, 0, 1) {
		resp.SemanaMinutos += minutos[dia(f)]
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

// nombreCorto muestra el nombre de pila y la inicial del apellido: "Juan P."
func nombreCorto(nombre string) string {
	partes := strings.Fields(nombre)
	if len(partes) == 0 {
		return "Anónimo"
	}
	if len(partes) == 1 {
		return partes[0]
	}
	inicial := []rune(partes[len(partes)-1])[0]
	return partes[0] + " " + strings.ToUpper(string(inicial)) + "."
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
			Nombre:   nombreCorto(f.Nombre),
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

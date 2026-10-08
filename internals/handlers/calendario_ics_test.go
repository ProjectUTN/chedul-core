package handlers

import (
	"chedul-core/internals/domain"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArmarICS(t *testing.T) {
	ahora := time.Date(2026, 10, 8, 3, 0, 0, 0, time.UTC)
	hora := "18:30"
	segundo := "2C"
	eventos := []domain.Evento{
		{ID: 1, Titulo: "Parcial 1", Tipo: "parcial", Fecha: "2026-10-20", Hora: &hora, Materia: &domain.MateriaResumen{ID: 5, Nombre: "Gestión Gerencial"}},
		{ID: 2, Titulo: "Entregar TP; versión final, con todo", Tipo: "entrega", Fecha: "2026-11-02"},
	}
	clases := []domain.Clase{
		{ID: 7, Titulo: "Ciencia de Datos", Dia: 1, HoraInicio: "19:00", HoraFin: "22:00", Aula: "Aula 12", Tipo: "clase", Cuatrimestre: &segundo},
		{ID: 8, Titulo: "Trabajo", Dia: 3, HoraInicio: "09:00", HoraFin: "13:00", Tipo: "trabajo"},
	}
	academicas := []domain.FechaAcademica{
		{ID: 3, Titulo: "Feriado", Tipo: "feriado", Desde: "2026-10-12", Hasta: "2026-10-12"},
	}

	ics := ArmarICS(eventos, clases, academicas, ahora)

	assert.True(t, strings.HasPrefix(ics, "BEGIN:VCALENDAR\r\n"))
	assert.True(t, strings.HasSuffix(ics, "END:VCALENDAR\r\n"))
	assert.Contains(t, ics, "DTSTART;TZID=America/Argentina/Buenos_Aires:20261020T183000\r\nDURATION:PT2H")
	assert.Contains(t, ics, "SUMMARY:Parcial 1 · Gestión Gerencial")
	assert.Contains(t, ics, `SUMMARY:Entregar TP\; versión final\, con todo`)
	assert.Contains(t, ics, "DTSTART;VALUE=DATE:20261102\r\nDTEND;VALUE=DATE:20261103")

	// La del 2° cuatrimestre: lunes de agosto a diciembre, sin el feriado
	assert.Contains(t, ics, "DTSTART;TZID=America/Argentina/Buenos_Aires:20260803T190000")
	assert.Contains(t, ics, "RRULE:FREQ=WEEKLY;BYDAY=MO;UNTIL=20270101T025959Z")
	assert.Contains(t, ics, "EXDATE;TZID=America/Argentina/Buenos_Aires:20261012T190000")
	assert.Contains(t, ics, "LOCATION:Aula 12")
	// El trabajo se repite sin fin y no se saltea feriados
	assert.Contains(t, ics, "RRULE:FREQ=WEEKLY;BYDAY=WE\r\n")
	assert.Equal(t, 1, strings.Count(ics, "EXDATE"))

	assert.Contains(t, ics, "UID:facultad-3@chedul")
	for _, linea := range strings.Split(ics, "\r\n") {
		assert.LessOrEqual(t, len(linea), 75, linea)
	}
}

func TestLineaLargaICS(t *testing.T) {
	var w icsWriter
	w.linea("SUMMARY:" + strings.Repeat("ñ", 60))
	lineas := strings.Split(strings.TrimSuffix(w.b.String(), "\r\n"), "\r\n")
	assert.Greater(t, len(lineas), 1)
	for _, l := range lineas {
		assert.LessOrEqual(t, len(l), 75)
	}
	partes := strings.Split(strings.TrimSuffix(w.b.String(), "\r\n"), "\r\n ")
	assert.Equal(t, "SUMMARY:"+strings.Repeat("ñ", 60), strings.Join(partes, ""))
}

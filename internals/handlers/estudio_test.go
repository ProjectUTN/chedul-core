package handlers

import (
	"chedul-core/internals/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalcularRacha(t *testing.T) {
	hoy := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	dias := []domain.MinutosPorDia{
		{Fecha: "2026-10-03", Minutos: 30},
		{Fecha: "2026-10-05", Minutos: 25},
		{Fecha: "2026-10-06", Minutos: 50},
		{Fecha: "2026-10-07", Minutos: 10},
	}
	// Hoy todavia no estudio: la racha que viene de ayer sigue contando
	assert.Equal(t, 3, CalcularRacha(dias, hoy))
	assert.Equal(t, 4, CalcularRacha(append(dias, domain.MinutosPorDia{Fecha: "2026-10-08", Minutos: 5}), hoy))
	assert.Equal(t, 0, CalcularRacha(dias, hoy.AddDate(0, 0, 2)))
	assert.Equal(t, 0, CalcularRacha(nil, hoy))
}

func TestArmarRanking(t *testing.T) {
	filas := []domain.FilaRanking{
		{AlumnoID: 1, Nombre: "Ana María López", Minutos: 300},
		{AlumnoID: 2, Nombre: "beto", Minutos: 200},
		{AlumnoID: 3, Nombre: "Caro Díaz", Minutos: 200},
		{AlumnoID: 4, Nombre: "Dani Ruiz", Minutos: 0},
	}

	puestos := ArmarRanking(filas, 4, 2)
	// Los empatados comparten puesto y el alumno aparece aunque quede afuera
	assert.Equal(t, []domain.PuestoRanking{
		{Posicion: 1, Nombre: "Ana L.", Minutos: 300},
		{Posicion: 2, Nombre: "beto", Minutos: 200},
		{Posicion: 4, Nombre: "Dani R.", Minutos: 0, SoyYo: true},
	}, puestos)

	assert.Len(t, ArmarRanking(filas, 99, 20), 4)
	assert.Equal(t, 2, ArmarRanking(filas, 99, 20)[2].Posicion)
	assert.Empty(t, ArmarRanking(nil, 1, 20))
}

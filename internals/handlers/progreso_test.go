package handlers

import (
	"chedul-core/internals/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcularOrdenanza531(t *testing.T) {
	materias := []domain.Materia{
		{ID: 1, Nombre: "Sistemas Operativos", Nivel: 4, Tipo: "Obligatoria", CargaHoraria: 4, Horas: 96},
		{ID: 2, Nombre: "Inteligencia Artificial", Nivel: 5, Tipo: "Obligatoria", CargaHoraria: 6, Horas: 72},
		{ID: 3, Nombre: "Proyecto Final", Nivel: 5, Tipo: "Obligatoria", CargaHoraria: 6, Horas: 144},
		{ID: 4, Nombre: "Agilidad Avanzada", Nivel: 4, Tipo: "Electiva", CargaHoraria: 3, Horas: 72},
		{ID: 5, Nombre: "Práctica Supervisada", Nivel: 5, Tipo: "Obligatoria", CargaHoraria: 0, Horas: 96},
	}
	aprobada := func(id int64) domain.CondicionPorAlumno {
		return domain.CondicionPorAlumno{MateriaID: id, Condicion: "Aprobada"}
	}

	// Le faltan todas: 312 hs contra 216 del ultimo nivel
	r := CalcularOrdenanza531(materias, map[int64]domain.CondicionPorAlumno{})
	assert.False(t, r.Puede)
	assert.Equal(t, 216.0, r.HorasLimite)
	assert.Equal(t, 312.0, r.HorasFaltantes)

	// Con Sistemas Operativos aprobada le falta justo el ultimo nivel; la
	// electiva y la practica supervisada no cuentan
	r = CalcularOrdenanza531(materias, map[int64]domain.CondicionPorAlumno{
		1: aprobada(1),
		3: {MateriaID: 3, Condicion: "Regularizada"},
	})
	assert.True(t, r.Puede)
	assert.Equal(t, 216.0, r.HorasFaltantes)
	assert.Len(t, r.Faltantes, 2)
	assert.Equal(t, "Pendiente", r.Faltantes[0].EstadoActual)
	assert.Equal(t, "Regularizada", r.Faltantes[1].EstadoActual)

	// Recibido: no tiene nada que pedir
	r = CalcularOrdenanza531(materias, map[int64]domain.CondicionPorAlumno{1: aprobada(1), 2: aprobada(2), 3: aprobada(3)})
	assert.False(t, r.Puede)
	assert.Empty(t, r.Faltantes)
}

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
		{ID: 4, Nombre: "Agilidad Avanzada", Nivel: 4, Tipo: "Electiva", CargaHoraria: 10, Horas: 72},
		{ID: 6, Nombre: "Ciberseguridad", Nivel: 5, Tipo: "Electiva", CargaHoraria: 8, Horas: 72},
		{ID: 5, Nombre: "Práctica Supervisada", Nivel: 5, Tipo: "Obligatoria", CargaHoraria: 0, Horas: 96},
	}
	con := func(condicion string, ids ...int64) map[int64]domain.CondicionPorAlumno {
		estados := map[int64]domain.CondicionPorAlumno{}
		for _, id := range ids {
			estados[id] = domain.CondicionPorAlumno{MateriaID: id, Condicion: condicion}
		}
		return estados
	}

	// Le falta todo: 16 hs de obligatorias y 20 de electivas contra 12
	r := CalcularOrdenanza531(materias, map[int64]domain.CondicionPorAlumno{})
	assert.False(t, r.Puede)
	assert.Equal(t, 12.0, r.HorasLimite)
	assert.Equal(t, 20.0, r.HorasElectivasFaltantes)
	assert.Equal(t, 36.0, r.HorasFaltantes)

	// Con 18 hs de electivas le faltan 2: 12 + 2 supera el limite
	r = CalcularOrdenanza531(materias, con("Aprobada", 1, 4, 6))
	assert.Equal(t, 2.0, r.HorasElectivasFaltantes)
	assert.Equal(t, 14.0, r.HorasFaltantes)
	assert.False(t, r.Puede)

	// Con las electivas completas le falta justo el ultimo nivel; la
	// practica supervisada no cuenta
	estados := con("Aprobada", 1, 4, 6)
	estados[2] = domain.CondicionPorAlumno{MateriaID: 2, Condicion: "Cursando"}
	r = CalcularOrdenanza531(materias, estados)
	assert.True(t, r.Puede)
	assert.Equal(t, 2.0, r.HorasElectivasFaltantes)
	assert.Equal(t, 8.0, r.HorasFaltantes)
	assert.Len(t, r.Faltantes, 1)
	assert.Equal(t, "Pendiente", r.Faltantes[0].EstadoActual)

	// La electiva regularizada o en curso tambien cuenta como hecha
	estados = con("Regularizada", 1, 4)
	estados[6] = domain.CondicionPorAlumno{MateriaID: 6, Condicion: "Cursando"}
	r = CalcularOrdenanza531(materias, estados)
	assert.Equal(t, 2.0, r.HorasElectivasFaltantes)
	assert.Equal(t, 14.0, r.HorasFaltantes)

	// Recibido: no tiene nada que pedir
	r = CalcularOrdenanza531(append(materias, domain.Materia{ID: 7, Nombre: "Otra", Nivel: 4, Tipo: "Electiva", CargaHoraria: 2}), con("Aprobada", 1, 2, 3, 4, 6, 7))
	assert.False(t, r.Puede)
	assert.Empty(t, r.Faltantes)
	assert.Zero(t, r.HorasElectivasFaltantes)
}

func TestCalcularProgresoSinElectivasDescartadas(t *testing.T) {
	materias := []domain.Materia{
		{ID: 1, Nombre: "Agilidad Avanzada", Nivel: 4, Tipo: "Electiva"},
		{ID: 2, Nombre: "Ciberseguridad", Nivel: 4, Tipo: "Electiva"},
	}
	r := CalcularProgreso(1, materias, []domain.CondicionPorAlumno{{MateriaID: 1, Condicion: "No me interesa"}})
	assert.Len(t, r.MateriasPendientesDisponibles, 1)
	assert.Equal(t, int64(2), r.MateriasPendientesDisponibles[0].ID)
	assert.Empty(t, r.MateriasPendientesNoDisponibles)
}

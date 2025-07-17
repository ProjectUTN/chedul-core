package repositories

import (
	"chedul-core/internals/domain"
	"context"

	"github.com/uptrace/bun"
)

type MateriaModel struct {
	bun.BaseModel  `bun:"table:materia"`
	ID             int64   `json:"id" bun:"id,pk,autoincrement"`
	Nombre         string  `json:"nombre" bun:"nombre,notnull"`
	CargaHoraria   int     `json:"carga_horaria" bun:"carga_horaria,notnull"`
	CorrelativaID  int     `json:"correlativa_id,omitempty" bun:"correlativa_id"`
	Nivel          int     `json:"nivel" bun:"nivel,notnull"`
	Area           string  `json:"area" bun:"area,notnull"`
	Tipo           string  `json:"tipo" bun:"tipo,notnull"`
	CuatrimestreID int     `json:"cuatrimestre_id" bun:"cuatrimestre_id,notnull"`
	Horas          float64 `json:"horas" bun:"horas"`
	Bloque         string  `json:"bloque" bun:"bloque,notnull"`
	Programa       string  `json:"programa" bun:"programa,notnull"`
}

type materiaRepository struct {
	db *bun.DB
}

func (r *materiaRepository) toDomain(model MateriaModel) domain.Materia {
	return domain.Materia{
		ID:             model.ID,
		Nombre:         model.Nombre,
		CargaHoraria:   model.CargaHoraria,
		CorrelativaID:  model.CorrelativaID,
		Nivel:          model.Nivel,
		Area:           model.Area,
		Tipo:           model.Tipo,
		CuatrimestreID: model.CuatrimestreID,
		Horas:          model.Horas,
		Bloque:         model.Bloque,
		Programa:       model.Programa,
	}
}

func NewMateriaRepository(db *bun.DB) domain.MateriaRepository {
	return &materiaRepository{db: db}
}

func (r *materiaRepository) GetAll(ctx context.Context) ([]domain.Materia, error) {
	var model []MateriaModel
	err := r.db.NewSelect().Model(&model).Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Materia, len(model))
	for i, model := range model {
		result[i] = r.toDomain(model)
	}
	return result, nil
}

func (r *materiaRepository) GetByID(ctx context.Context, id int64) (*domain.Materia, error) {
	var model MateriaModel
	err := r.db.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	materia := r.toDomain(model)
	return &materia, nil
}

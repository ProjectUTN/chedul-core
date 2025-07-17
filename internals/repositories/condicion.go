package repositories

import (
	"chedul-core/internals/domain"
	"context"

	"github.com/uptrace/bun"
)

type CondicionModel struct {
	bun.BaseModel `bun:"table:carrera"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull,unique"`
}

type condicionRepository struct {
	db *bun.DB
}

func NewCondicionRepository(db *bun.DB) domain.CondicionRepository {
	return &condicionRepository{db: db}
}

func (r *condicionRepository) GetAll(ctx context.Context) ([]domain.Condicion, error) {
	var model []CondicionModel
	err := r.db.NewSelect().Model(&model).Order("nombre ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Condicion, len(model))
	for i, model := range model {
		result[i] = r.toDomain(model)
	}

	return result, nil
}

func (r *condicionRepository) toDomain(model CondicionModel) domain.Condicion {
	return domain.Condicion{
		ID:        model.ID,
		Condicion: model.Nombre,
	}
}

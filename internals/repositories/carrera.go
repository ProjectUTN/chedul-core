package repositories

import (
	"chedul-core/internals/domain"
	"context"

	"github.com/uptrace/bun"
)

type CarreraModel struct {
	bun.BaseModel `bun:"table:carrera"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull,unique"`
}

type carreraRepository struct {
	db *bun.DB
}

func NewCarreraRepository(db *bun.DB) domain.CarreraRepository {
	return &carreraRepository{db: db}
}

func (r *carreraRepository) GetAll(ctx context.Context) ([]domain.Carrera, error) {
	var model []CarreraModel
	err := r.db.NewSelect().Model(&model).Order("nombre ASC").Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Carrera, len(model))
	for i, model := range model {
		result[i] = r.toDomain(model)
	}

	return result, nil
}

func (r *carreraRepository) GetByID(ctx context.Context, id int64) (*domain.Carrera, error) {
	var model CarreraModel
	err := r.db.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	carrera := r.toDomain(model)
	return &carrera, nil
}

func (r *carreraRepository) GetByName(ctx context.Context, nombre string) (*domain.Carrera, error) {
	var model CarreraModel
	err := r.db.NewSelect().Model(&model).Where("nombre = ?", nombre).Scan(ctx)
	if err != nil {
		return nil, err
	}

	carrera := r.toDomain(model)
	return &carrera, nil
}

func (r *carreraRepository) toDomain(model CarreraModel) domain.Carrera {
	return domain.Carrera{
		ID:     model.ID,
		Nombre: model.Nombre,
	}
}

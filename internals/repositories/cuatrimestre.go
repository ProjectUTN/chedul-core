package repositories

import (
	"chedul-core/internals/domain"

	"github.com/uptrace/bun"
	"golang.org/x/net/context"
)

type CuatrimestreModel struct {
	bun.BaseModel `bun:"table:cuatrimestre"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull,unique"`
}

type cuatrimestreRepository struct {
	db *bun.DB
}

func (r *cuatrimestreRepository) toDomain(model CuatrimestreModel) domain.Cuatrimestre {
	return domain.Cuatrimestre{
		ID:     model.ID,
		Nombre: model.Nombre,
	}
}

func NewCuatrimestreRepository(db *bun.DB) domain.CuatrimestreRepository {
	return &cuatrimestreRepository{db: db}
}

func (r *cuatrimestreRepository) GetAll(ctx context.Context) ([]domain.Cuatrimestre, error) {
	var model []CuatrimestreModel
	err := r.db.NewSelect().Model(&model).Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Cuatrimestre, len(model))
	for i, model := range model {
		result[i] = r.toDomain(model)
	}

	return result, nil
}

func (r *cuatrimestreRepository) GetByID(ctx context.Context, id int64) (*domain.Cuatrimestre, error) {
	var model CuatrimestreModel
	err := r.db.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	cuatrimestre := r.toDomain(model)
	return &cuatrimestre, nil
}

func (r *cuatrimestreRepository) GetByNombre(ctx context.Context, nombre string) (*domain.Cuatrimestre, error) {
	var model CuatrimestreModel
	err := r.db.NewSelect().Model(&model).Where("email = ?", nombre).Scan(ctx)
	if err != nil {
		return nil, err
	}

	cuatrimestre := r.toDomain(model)
	return &cuatrimestre, nil
}

package repositories

import (
	"chedul-core/internals/domain"
	"context"

	"github.com/uptrace/bun"
)

type AlumnoModel struct {
	bun.BaseModel `bun:"table:alumno"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull"`
	Email         string `json:"email" bun:"email,unique"`
	Carrera       int64  `json:"carrera" bun:"carrera_id,notnull"`
}

type alumnoRepository struct {
	db *bun.DB
}

func (r *alumnoRepository) toDomain(model AlumnoModel) domain.Alumno {
	return domain.Alumno{
		ID:      model.ID,
		Nombre:  model.Nombre,
		Email:   model.Email,
		Carrera: model.Carrera,
	}
}

func (r *alumnoRepository) toModel(alumno domain.Alumno) AlumnoModel {
	return AlumnoModel{
		ID:      alumno.ID,
		Nombre:  alumno.Nombre,
		Email:   alumno.Email,
		Carrera: alumno.Carrera,
	}
}
func NewAlumnoRepository(db *bun.DB) domain.AlumnoRepository {
	return &alumnoRepository{db: db}
}

func (r *alumnoRepository) GetAll(ctx context.Context) ([]domain.Alumno, error) {
	var model []AlumnoModel
	err := r.db.NewSelect().Model(&model).Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Alumno, len(model))
	for i, model := range model {
		result[i] = r.toDomain(model)
	}

	return result, nil
}

func (r *alumnoRepository) GetByID(ctx context.Context, id int64) (*domain.Alumno, error) {
	var model AlumnoModel
	err := r.db.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	alumno := r.toDomain(model)
	return &alumno, nil
}

func (r *alumnoRepository) GetByEmail(ctx context.Context, email string) (*domain.Alumno, error) {
	var model AlumnoModel
	err := r.db.NewSelect().Model(&model).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}

	alumno := r.toDomain(model)
	return &alumno, nil
}

func (r *alumnoRepository) Create(ctx context.Context, alumno *domain.Alumno) error {
	model := r.toModel(*alumno)
	_, err := r.db.NewInsert().Model(&model).Exec(ctx)
	if err != nil {
		return err
	}

	alumno.ID = model.ID
	return nil
}

func (r *alumnoRepository) Update(ctx context.Context, alumno *domain.Alumno) error {
	model := r.toModel(*alumno)
	_, err := r.db.NewUpdate().
		Model(&model).
		Where("id = ?", alumno.ID).
		Exec(ctx)
	return err
}

func (r *alumnoRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().
		Model((*AlumnoModel)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

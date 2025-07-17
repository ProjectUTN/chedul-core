package repositories

import (
	"chedul-core/internals/domain"
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

type CondicionAlumnoModel struct {
	bun.BaseModel `bun:"table:condicion_alumno"`
	ID            int64 `json:"id" bun:"id,pk,autoincrement"`
	CondicionID   int64 `json:"condicion_id" bun:"condicion_id,notnull"`
	Nota          int   `json:"nota,omitempty" bun:"nota"`
	AlumnoID      int64 `json:"alumno_id" bun:"alumno_id,notnull"`
	MateriaID     int64 `json:"materia_id" bun:"materia_id,notnull"`
}

type condicionAlumnoRepository struct {
	db *bun.DB
}

func (r *condicionAlumnoRepository) GetCondicionPorAlumno(ctx context.Context, id int64) ([]domain.CondicionPorAlumno, error) {
	var condiciones []domain.CondicionPorAlumno
	err := r.db.NewSelect().
		Model(&condiciones).
		Relation("Materia").
		Relation("Condicion").
		Where("alumno_id = ?", id).
		Scan(ctx)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return condiciones, nil
}

func (r *condicionAlumnoRepository) Create(ctx context.Context, condicionAlumno *domain.CondicionAlumno) error {
	model := r.toModel(*condicionAlumno)

	_, err := r.db.NewInsert().
		Model(&model).
		On("conflict(alumno_id, materia_id) do update").
		Set("condicion_id = excluded.condicion_id, nota = excluded.nota").
		Exec(ctx)

	if err != nil {
		return err
	}

	condicionAlumno.ID = model.ID
	return nil
}

func (r *condicionAlumnoRepository) GetAll(ctx context.Context) ([]domain.CondicionAlumno, error) {
	var model []CondicionAlumnoModel
	err := r.db.NewSelect().Model(&model).Scan(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.CondicionAlumno, len(model))
	for i, model := range model {
		result[i] = r.toDomain(model)
	}
	return result, nil
}

func (r *condicionAlumnoRepository) GetByID(ctx context.Context, id int64) (*domain.CondicionAlumno, error) {
	var model CondicionAlumnoModel
	err := r.db.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	condicion := r.toDomain(model)
	return &condicion, nil
}

func (r *condicionAlumnoRepository) Update(ctx context.Context, condicionAlumno *domain.CondicionAlumno) error {
	panic("unimplemented")
}

func NewCondicionAlumnoRepository(db *bun.DB) domain.CondicionAlumnoRepository {
	return &condicionAlumnoRepository{db: db}
}

func (r *condicionAlumnoRepository) toDomain(model CondicionAlumnoModel) domain.CondicionAlumno {
	return domain.CondicionAlumno{
		ID:          model.ID,
		CondicionID: model.CondicionID,
		Nota:        model.Nota,
		AlumnoID:    model.AlumnoID,
		MateriaID:   model.MateriaID,
	}
}

func (r *condicionAlumnoRepository) toModel(condicion domain.CondicionAlumno) CondicionAlumnoModel {
	return CondicionAlumnoModel{
		ID:          condicion.ID,
		CondicionID: condicion.CondicionID,
		Nota:        condicion.Nota,
		AlumnoID:    condicion.AlumnoID,
		MateriaID:   condicion.MateriaID,
	}
}

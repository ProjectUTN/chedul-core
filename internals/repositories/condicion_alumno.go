package repositories

import (
	"chedul-core/internals/domain"
	"context"

	"github.com/uptrace/bun"
)

type CondicionAlumnoModel struct {
	bun.BaseModel `bun:"table:condicion_alumno"`
	ID            int64 `bun:"id,pk,autoincrement"`
	CondicionID   int64 `bun:"condicion_id,notnull"`
	Nota          *int  `bun:"nota"`
	AlumnoID      int64 `bun:"alumno_id,notnull"`
	MateriaID     int64 `bun:"materia_id,notnull"`
}

type condicionAlumnoRepository struct {
	db *bun.DB
}

func NewCondicionAlumnoRepository(db *bun.DB) domain.CondicionAlumnoRepository {
	return &condicionAlumnoRepository{db: db}
}

func (r *condicionAlumnoRepository) GetCondicionPorAlumno(ctx context.Context, alumnoID int64) ([]domain.CondicionPorAlumno, error) {
	condiciones := []domain.CondicionPorAlumno{}
	err := r.db.NewSelect().
		TableExpr("condicion_alumno AS ca").
		ColumnExpr("ca.materia_id, m.nombre AS materia, ca.condicion_id, c.condicion, ca.nota").
		Join("JOIN materia AS m ON m.id = ca.materia_id").
		Join("JOIN condicion AS c ON c.id = ca.condicion_id").
		Where("ca.alumno_id = ?", alumnoID).
		OrderExpr("m.nivel, m.nombre").
		Scan(ctx, &condiciones)
	if err != nil {
		return nil, err
	}

	return condiciones, nil
}

func (r *condicionAlumnoRepository) Set(ctx context.Context, condicionAlumno *domain.CondicionAlumno) error {
	model := CondicionAlumnoModel{
		CondicionID: condicionAlumno.CondicionID,
		Nota:        condicionAlumno.Nota,
		AlumnoID:    condicionAlumno.AlumnoID,
		MateriaID:   condicionAlumno.MateriaID,
	}

	_, err := r.db.NewInsert().
		Model(&model).
		On("CONFLICT (alumno_id, materia_id) DO UPDATE").
		Set("condicion_id = EXCLUDED.condicion_id, nota = EXCLUDED.nota").
		Returning("id").
		Exec(ctx)
	if err != nil {
		return err
	}

	condicionAlumno.ID = model.ID
	return nil
}

func (r *condicionAlumnoRepository) Delete(ctx context.Context, alumnoID, materiaID int64) error {
	_, err := r.db.NewDelete().
		Model((*CondicionAlumnoModel)(nil)).
		Where("alumno_id = ? AND materia_id = ?", alumnoID, materiaID).
		Exec(ctx)
	return err
}

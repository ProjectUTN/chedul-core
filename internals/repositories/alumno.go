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
	Password      string `json:"-" bun:"password"`
}

type alumnoRepository struct {
	db *bun.DB
}

func (r *alumnoRepository) toDomain(model AlumnoModel) (domain.Alumno, error) {
	email, err := domain.ParseEmail(model.Email)
	if err != nil {
		return domain.Alumno{}, err
	}

	userName, err := domain.NewUsername(model.Nombre)

	if err != nil {
		return domain.Alumno{}, err
	}

	return domain.Alumno{
		ID:      model.ID,
		Nombre:  userName,
		Email:   email,
		Carrera: model.Carrera,
		// TODO: Todavia no se si es peligroso asumir que la clave en la base
		// de datos ya esta encriptada
		Password: domain.ParsePasswordFromEncrypted(model.Password),
	}, nil
}

func (r *alumnoRepository) toModel(alumno domain.Alumno) AlumnoModel {
	return AlumnoModel{
		ID:       alumno.ID,
		Nombre:   alumno.Nombre.String(),
		Email:    alumno.Email.String(),
		Carrera:  alumno.Carrera,
		Password: alumno.Password.String(),
	}
}

func NewAlumnoRepository(db *bun.DB) domain.AlumnoRepository {
	return &alumnoRepository{db: db}
}

func (r *alumnoRepository) GetByID(ctx context.Context, id int64) (*domain.Alumno, error) {
	var model AlumnoModel
	err := r.db.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}

	alumno, err := r.toDomain(model)
	if err != nil {
		return nil, err
	}
	return &alumno, nil
}

func (r *alumnoRepository) GetByEmail(ctx context.Context, email string) (*domain.Alumno, error) {
	var model AlumnoModel
	err := r.db.NewSelect().Model(&model).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}

	alumno, err := r.toDomain(model)
	if err != nil {
		return nil, err
	}

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

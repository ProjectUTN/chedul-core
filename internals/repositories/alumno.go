package repositories

import (
	"chedul-core/internals/domain"
	"context"
	"fmt"

	"github.com/uptrace/bun"
)

type AlumnoModel struct {
	bun.BaseModel `bun:"table:alumno"`
	ID            int64  `json:"id" bun:"id,pk,autoincrement"`
	Nombre        string `json:"nombre" bun:"nombre,notnull"`
	Email         string `json:"email" bun:"email,unique"`
	Carrera       int64  `json:"carrera" bun:"carrera_id,notnull"`
	Password      string `json:"-" bun:"password"`
	// Update no los toca: se cambian con sus propios metodos
	EsAdmin         bool `bun:"es_admin,notnull"`
	GoogleVinculado bool `bun:"google_vinculado,notnull"`
	VersionSesion   int  `bun:"version_sesion,notnull"`
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
		Password:        domain.ParsePasswordFromEncrypted(model.Password),
		EsAdmin:         model.EsAdmin,
		GoogleVinculado: model.GoogleVinculado,
		VersionSesion:   model.VersionSesion,
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
		Column("nombre", "email", "carrera_id", "password").
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

func (r *alumnoRepository) RegistrarAcceso(ctx context.Context, id int64, metodo string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.ExecContext(ctx, "insert into acceso (alumno_id, metodo) values (?, ?)", id, metodo); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "update alumno set ultimo_acceso = now() where id = ?", id)
		return err
	})
}

// MarcarActivo escribe como mucho una vez cada 5 minutos por alumno
func (r *alumnoRepository) MarcarActivo(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `update alumno set ultimo_acceso = now()
		where id = ? and (ultimo_acceso is null or ultimo_acceso < now() - interval '5 minutes')`, id)
	return err
}

func (r *alumnoRepository) VincularGoogle(ctx context.Context, id int64, nuevaClave *domain.Password) error {
	if nuevaClave == nil {
		_, err := r.db.ExecContext(ctx, "update alumno set google_vinculado = true where id = ?", id)
		return err
	}
	_, err := r.db.ExecContext(ctx, `update alumno set google_vinculado = true, password = ?,
		version_sesion = version_sesion + 1 where id = ?`, nuevaClave.String(), id)
	return err
}

func (r *alumnoRepository) SubirVersionSesion(ctx context.Context, id int64) (int, error) {
	var version int
	err := r.db.QueryRowContext(ctx,
		"update alumno set version_sesion = version_sesion + 1 where id = ? returning version_sesion", id).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("subiendo version de sesion: %w", err)
	}
	return version, nil
}

func (r *alumnoRepository) Contar(ctx context.Context) (int, error) {
	return r.db.NewSelect().Model((*AlumnoModel)(nil)).Count(ctx)
}

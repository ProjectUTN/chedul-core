package repositories

import (
	"chedul-core/internals/domain"
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/uptrace/bun"
)

type SesionEstudioModel struct {
	bun.BaseModel `bun:"table:sesion_estudio"`
	ID            int64  `bun:"id,pk,autoincrement"`
	AlumnoID      int64  `bun:"alumno_id,notnull"`
	MateriaID     *int64 `bun:"materia_id"`
	Modo          string `bun:"modo,notnull"`
	Minutos       int    `bun:"minutos,notnull"`
}

type sesionRow struct {
	ID            int64          `bun:"id"`
	Modo          string         `bun:"modo"`
	Minutos       int            `bun:"minutos"`
	Fin           string         `bun:"fin"`
	MateriaID     sql.NullInt64  `bun:"materia_id"`
	MateriaNombre sql.NullString `bun:"materia_nombre"`
}

func (row sesionRow) toDomain() domain.SesionEstudio {
	return domain.SesionEstudio{
		ID:      row.ID,
		Modo:    row.Modo,
		Minutos: row.Minutos,
		Fin:     row.Fin,
		Materia: materiaResumen(row.MateriaID, row.MateriaNombre),
	}
}

// diaLocal es el dia (en Argentina) en que termino la sesion
const diaLocal = "(s.fin AT TIME ZONE '" + domain.ZonaHoraria + "')::date"

type estudioRepository struct {
	db *bun.DB
}

func NewEstudioRepository(db *bun.DB) domain.EstudioRepository {
	return &estudioRepository{db: db}
}

func (r *estudioRepository) selectSesiones(alumnoID int64) *bun.SelectQuery {
	return r.db.NewSelect().
		TableExpr("sesion_estudio AS s").
		ColumnExpr("s.id, s.modo, s.minutos").
		ColumnExpr(`to_char(s.fin AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS fin`).
		ColumnExpr("m.id AS materia_id, m.nombre AS materia_nombre").
		Join("LEFT JOIN materia AS m ON m.id = s.materia_id").
		Where("s.alumno_id = ?", alumnoID)
}

func (r *estudioRepository) CreateSesion(ctx context.Context, alumnoID int64, datos domain.DatosSesion) (int64, error) {
	model := SesionEstudioModel{
		AlumnoID:  alumnoID,
		MateriaID: datos.MateriaID,
		Modo:      datos.Modo,
		Minutos:   datos.Minutos,
	}
	if _, err := r.db.NewInsert().Model(&model).Returning("id").Exec(ctx); err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *estudioRepository) GetSesion(ctx context.Context, alumnoID, id int64) (*domain.SesionEstudio, error) {
	var rows []sesionRow
	if err := r.selectSesiones(alumnoID).Where("s.id = ?", id).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrSesionNoEncontrada
	}
	sesion := rows[0].toDomain()
	return &sesion, nil
}

func (r *estudioRepository) ListSesiones(ctx context.Context, alumnoID int64, limite int) ([]domain.SesionEstudio, error) {
	var rows []sesionRow
	err := r.selectSesiones(alumnoID).
		OrderExpr("s.fin DESC, s.id DESC").
		Limit(limite).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	sesiones := make([]domain.SesionEstudio, len(rows))
	for i, row := range rows {
		sesiones[i] = row.toDomain()
	}
	return sesiones, nil
}

func (r *estudioRepository) DeleteSesion(ctx context.Context, alumnoID, id int64) error {
	res, err := r.db.NewDelete().
		Model((*SesionEstudioModel)(nil)).
		Where("id = ? AND alumno_id = ?", id, alumnoID).
		Exec(ctx)
	return filasAfectadas(res, err, domain.ErrSesionNoEncontrada)
}

func (r *estudioRepository) MinutosPorDia(ctx context.Context, alumnoID int64, desde, hasta string) ([]domain.MinutosPorDia, error) {
	dias := []domain.MinutosPorDia{}
	err := r.db.NewSelect().
		TableExpr("sesion_estudio AS s").
		ColumnExpr("to_char("+diaLocal+", 'YYYY-MM-DD') AS fecha").
		ColumnExpr("sum(s.minutos)::int AS minutos").
		Where("s.alumno_id = ?", alumnoID).
		Where(diaLocal+" BETWEEN ?::date AND ?::date", desde, hasta).
		GroupExpr("fecha").
		OrderExpr("fecha ASC").
		Scan(ctx, &dias)
	return dias, err
}

func (r *estudioRepository) MinutosPorMateria(ctx context.Context, alumnoID int64, desde, hasta string) ([]domain.MinutosPorMateria, error) {
	materias := []domain.MinutosPorMateria{}
	err := r.db.NewSelect().
		TableExpr("sesion_estudio AS s").
		ColumnExpr("m.id AS materia_id, coalesce(m.nombre, '') AS nombre").
		ColumnExpr("sum(s.minutos)::int AS minutos").
		Join("LEFT JOIN materia AS m ON m.id = s.materia_id").
		Where("s.alumno_id = ?", alumnoID).
		Where(diaLocal+" BETWEEN ?::date AND ?::date", desde, hasta).
		GroupExpr("m.id, m.nombre").
		OrderExpr("minutos DESC, nombre ASC").
		Scan(ctx, &materias)
	return materias, err
}

func (r *estudioRepository) EnRanking(ctx context.Context, alumnoID int64) (bool, error) {
	var participa bool
	err := r.db.NewSelect().
		TableExpr("alumno").
		Column("en_ranking").
		Where("id = ?", alumnoID).
		Scan(ctx, &participa)
	return participa, err
}

func (r *estudioRepository) SetEnRanking(ctx context.Context, alumnoID int64, participar bool) error {
	_, err := r.db.NewUpdate().
		TableExpr("alumno").
		Set("en_ranking = ?", participar).
		Where("id = ?", alumnoID).
		Exec(ctx)
	return err
}

func (r *estudioRepository) Ranking(ctx context.Context, desde, hasta string) ([]domain.FilaRanking, error) {
	filas := []domain.FilaRanking{}
	err := r.db.NewSelect().
		TableExpr("alumno AS a").
		ColumnExpr("a.id AS alumno_id, a.nombre, a.apellido").
		ColumnExpr("coalesce(sum(s.minutos), 0)::int AS minutos").
		Join("LEFT JOIN sesion_estudio AS s ON s.alumno_id = a.id AND "+diaLocal+" BETWEEN ?::date AND ?::date", desde, hasta).
		Where("a.en_ranking").
		GroupExpr("a.id, a.nombre, a.apellido").
		OrderExpr("minutos DESC, a.id ASC").
		Scan(ctx, &filas)
	return filas, err
}

func (r *estudioRepository) MetaDiaria(ctx context.Context, alumnoID int64) (int, error) {
	var minutos int
	err := r.db.NewSelect().
		TableExpr("alumno").
		Column("meta_diaria_minutos").
		Where("id = ?", alumnoID).
		Scan(ctx, &minutos)
	return minutos, err
}

func (r *estudioRepository) SetMetaDiaria(ctx context.Context, alumnoID int64, minutos int) error {
	_, err := r.db.NewUpdate().
		TableExpr("alumno").
		Set("meta_diaria_minutos = ?", minutos).
		Where("id = ?", alumnoID).
		Exec(ctx)
	return err
}

type TareaEstudioModel struct {
	bun.BaseModel `bun:"table:tarea_estudio"`
	ID            int64  `bun:"id,pk,autoincrement"`
	AlumnoID      int64  `bun:"alumno_id,notnull"`
	MateriaID     *int64 `bun:"materia_id"`
	Titulo        string `bun:"titulo,notnull"`
	Hecha         bool   `bun:"hecha,notnull"`
}

type tareaRow struct {
	ID            int64          `bun:"id"`
	Titulo        string         `bun:"titulo"`
	Hecha         bool           `bun:"hecha"`
	Creada        string         `bun:"creada"`
	MateriaID     sql.NullInt64  `bun:"materia_id"`
	MateriaNombre sql.NullString `bun:"materia_nombre"`
}

func (row tareaRow) toDomain() domain.TareaEstudio {
	return domain.TareaEstudio{
		ID:      row.ID,
		Titulo:  row.Titulo,
		Hecha:   row.Hecha,
		Creada:  row.Creada,
		Materia: materiaResumen(row.MateriaID, row.MateriaNombre),
	}
}

func (r *estudioRepository) selectTareas(alumnoID int64) *bun.SelectQuery {
	return r.db.NewSelect().
		TableExpr("tarea_estudio AS t").
		ColumnExpr("t.id, t.titulo, t.hecha").
		ColumnExpr(`to_char(t.creada AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS creada`).
		ColumnExpr("m.id AS materia_id, m.nombre AS materia_nombre").
		Join("LEFT JOIN materia AS m ON m.id = t.materia_id").
		Where("t.alumno_id = ?", alumnoID)
}

// ListTareas devuelve primero las pendientes y despues las hechas, las mas nuevas arriba
func (r *estudioRepository) ListTareas(ctx context.Context, alumnoID int64) ([]domain.TareaEstudio, error) {
	var rows []tareaRow
	if err := r.selectTareas(alumnoID).OrderExpr("t.hecha ASC, t.creada DESC, t.id DESC").Scan(ctx, &rows); err != nil {
		return nil, err
	}
	tareas := make([]domain.TareaEstudio, len(rows))
	for i, row := range rows {
		tareas[i] = row.toDomain()
	}
	return tareas, nil
}

func (r *estudioRepository) GetTarea(ctx context.Context, alumnoID, id int64) (*domain.TareaEstudio, error) {
	var rows []tareaRow
	if err := r.selectTareas(alumnoID).Where("t.id = ?", id).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrTareaNoEncontrada
	}
	tarea := rows[0].toDomain()
	return &tarea, nil
}

func (r *estudioRepository) CreateTarea(ctx context.Context, alumnoID int64, datos domain.DatosTarea) (int64, error) {
	model := TareaEstudioModel{
		AlumnoID:  alumnoID,
		MateriaID: datos.MateriaID,
		Titulo:    datos.Titulo,
		Hecha:     datos.Hecha,
	}
	if _, err := r.db.NewInsert().Model(&model).Returning("id").Exec(ctx); err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *estudioRepository) UpdateTarea(ctx context.Context, alumnoID, id int64, datos domain.DatosTarea) error {
	res, err := r.db.NewUpdate().
		Model((*TareaEstudioModel)(nil)).
		Set("titulo = ?", datos.Titulo).
		Set("materia_id = ?", datos.MateriaID).
		Set("hecha = ?", datos.Hecha).
		Where("id = ? AND alumno_id = ?", id, alumnoID).
		Exec(ctx)
	return filasAfectadas(res, err, domain.ErrTareaNoEncontrada)
}

func (r *estudioRepository) DeleteTarea(ctx context.Context, alumnoID, id int64) error {
	res, err := r.db.NewDelete().
		Model((*TareaEstudioModel)(nil)).
		Where("id = ? AND alumno_id = ?", id, alumnoID).
		Exec(ctx)
	return filasAfectadas(res, err, domain.ErrTareaNoEncontrada)
}

func (r *estudioRepository) GetTemporizador(ctx context.Context, alumnoID int64) (*domain.Temporizador, error) {
	var crudo []byte
	var rev int
	err := r.db.QueryRowContext(ctx, "select estado, rev from temporizador where alumno_id = ?", alumnoID).Scan(&crudo, &rev)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t := domain.Temporizador{Rev: rev}
	if err := json.Unmarshal(crudo, &t.Estado); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *estudioRepository) GuardarTemporizador(ctx context.Context, alumnoID int64, estado domain.EstadoTemporizador, revBase int) (*domain.Temporizador, bool, error) {
	crudo, err := json.Marshal(estado)
	if err != nil {
		return nil, false, err
	}
	var rev int
	err = r.db.QueryRowContext(ctx, `insert into temporizador (alumno_id, estado) values (?, ?::jsonb)
		on conflict (alumno_id) do update set estado = excluded.estado, rev = temporizador.rev + 1, actualizado = now()
		where temporizador.rev = ? returning rev`, alumnoID, string(crudo), revBase).Scan(&rev)
	if errors.Is(err, sql.ErrNoRows) {
		actual, err := r.GetTemporizador(ctx, alumnoID)
		return actual, false, err
	}
	if err != nil {
		return nil, false, err
	}
	return &domain.Temporizador{Estado: estado, Rev: rev}, true, nil
}

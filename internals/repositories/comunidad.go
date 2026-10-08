package repositories

import (
	"chedul-core/internals/domain"
	"context"
	"database/sql"
	"strings"

	"github.com/uptrace/bun"
)

type ComunidadModel struct {
	bun.BaseModel `bun:"table:comunidad"`
	ID            int64  `bun:"id,pk,autoincrement"`
	AlumnoID      int64  `bun:"alumno_id"`
	MateriaID     *int64 `bun:"materia_id"`
	Nombre        string `bun:"nombre,notnull"`
	Descripcion   string `bun:"descripcion,notnull"`
	Plataforma    string `bun:"plataforma,notnull"`
	Link          string `bun:"link,notnull"`
}

type comunidadRow struct {
	ID            int64          `bun:"id"`
	Nombre        string         `bun:"nombre"`
	Descripcion   string         `bun:"descripcion"`
	Plataforma    string         `bun:"plataforma"`
	Link          string         `bun:"link"`
	Creada        string         `bun:"creada"`
	MateriaID     sql.NullInt64  `bun:"materia_id"`
	MateriaNombre sql.NullString `bun:"materia_nombre"`
	EsMia         bool           `bun:"es_mia"`
	Reportada     bool           `bun:"reportada"`
}

func (row comunidadRow) toDomain() domain.Comunidad {
	return domain.Comunidad{
		ID:          row.ID,
		Nombre:      row.Nombre,
		Descripcion: row.Descripcion,
		Plataforma:  row.Plataforma,
		Link:        row.Link,
		Creada:      row.Creada,
		Materia:     materiaResumen(row.MateriaID, row.MateriaNombre),
		EsMia:       row.EsMia,
		Reportada:   row.Reportada,
	}
}

type comunidadRepository struct {
	db *bun.DB
}

func NewComunidadRepository(db *bun.DB) domain.ComunidadRepository {
	return &comunidadRepository{db: db}
}

func (r *comunidadRepository) selectComunidades(viewerID int64) *bun.SelectQuery {
	return r.db.NewSelect().
		TableExpr("comunidad AS c").
		ColumnExpr("c.id, c.nombre, c.descripcion, c.plataforma, c.link").
		ColumnExpr(`to_char(c.creada AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS creada`).
		ColumnExpr("m.id AS materia_id, m.nombre AS materia_nombre").
		ColumnExpr("coalesce(c.alumno_id = ?, false) AS es_mia", viewerID).
		ColumnExpr("exists (select 1 from comunidad_reporte r where r.comunidad_id = c.id and r.alumno_id = ?) AS reportada", viewerID).
		Join("LEFT JOIN materia AS m ON m.id = c.materia_id").
		Where("(select count(*) from comunidad_reporte r where r.comunidad_id = c.id) < ?", domain.ComunidadReportesMax)
}

// List pone primero las generales de la carrera y despues las de cada materia
func (r *comunidadRepository) List(ctx context.Context, viewerID int64) ([]domain.Comunidad, error) {
	var rows []comunidadRow
	err := r.selectComunidades(viewerID).
		OrderExpr("m.nivel ASC NULLS FIRST, m.nombre ASC NULLS FIRST, c.nombre ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	comunidades := make([]domain.Comunidad, len(rows))
	for i, row := range rows {
		comunidades[i] = row.toDomain()
	}
	return comunidades, nil
}

func (r *comunidadRepository) Get(ctx context.Context, id, viewerID int64) (*domain.Comunidad, error) {
	var rows []comunidadRow
	if err := r.selectComunidades(viewerID).Where("c.id = ?", id).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrComunidadNoEncontrada
	}
	c := rows[0].toDomain()
	return &c, nil
}

func (r *comunidadRepository) Create(ctx context.Context, alumnoID int64, datos domain.DatosComunidad) (int64, error) {
	model := ComunidadModel{
		AlumnoID:    alumnoID,
		MateriaID:   datos.MateriaID,
		Nombre:      datos.Nombre,
		Descripcion: datos.Descripcion,
		Plataforma:  datos.Plataforma,
		Link:        datos.Link,
	}
	if _, err := r.db.NewInsert().Model(&model).Returning("id").Exec(ctx); err != nil {
		if strings.Contains(err.Error(), "comunidad_link_key") {
			return 0, domain.ErrComunidadRepetida
		}
		return 0, err
	}
	return model.ID, nil
}

func (r *comunidadRepository) Delete(ctx context.Context, id, alumnoID int64) error {
	res, err := r.db.NewDelete().
		Model((*ComunidadModel)(nil)).
		Where("id = ? AND alumno_id = ?", id, alumnoID).
		Exec(ctx)
	return filasAfectadas(res, err, domain.ErrComunidadNoEncontrada)
}

func (r *comunidadRepository) Reportar(ctx context.Context, id, alumnoID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO comunidad_reporte (comunidad_id, alumno_id) VALUES (?, ?) ON CONFLICT DO NOTHING", id, alumnoID)
	if err != nil && strings.Contains(err.Error(), "foreign key") {
		return domain.ErrComunidadNoEncontrada
	}
	return err
}

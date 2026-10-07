package repositories

import (
	"chedul-core/internals/domain"
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

type EventoModel struct {
	bun.BaseModel `bun:"table:evento"`
	ID            int64   `bun:"id,pk,autoincrement"`
	AlumnoID      int64   `bun:"alumno_id,notnull"`
	MateriaID     *int64  `bun:"materia_id"`
	Titulo        string  `bun:"titulo,notnull"`
	Tipo          string  `bun:"tipo,notnull"`
	Fecha         string  `bun:"fecha,notnull"`
	Hora          *string `bun:"hora"`
	Descripcion   string  `bun:"descripcion,notnull"`
}

type ClaseModel struct {
	bun.BaseModel `bun:"table:clase"`
	ID            int64  `bun:"id,pk,autoincrement"`
	AlumnoID      int64  `bun:"alumno_id,notnull"`
	MateriaID     *int64 `bun:"materia_id"`
	ComisionID    *int64 `bun:"comision_id"`
	Titulo        string `bun:"titulo,notnull"`
	Dia           int    `bun:"dia,notnull"`
	HoraInicio    string `bun:"hora_inicio,notnull"`
	HoraFin       string `bun:"hora_fin,notnull"`
	Aula          string `bun:"aula,notnull"`
}

type eventoRow struct {
	ID            int64          `bun:"id"`
	Titulo        string         `bun:"titulo"`
	Tipo          string         `bun:"tipo"`
	Fecha         string         `bun:"fecha"`
	Hora          sql.NullString `bun:"hora"`
	Descripcion   string         `bun:"descripcion"`
	MateriaID     sql.NullInt64  `bun:"materia_id"`
	MateriaNombre sql.NullString `bun:"materia_nombre"`
}

func (row eventoRow) toDomain() domain.Evento {
	evento := domain.Evento{
		ID:          row.ID,
		Titulo:      row.Titulo,
		Tipo:        row.Tipo,
		Fecha:       row.Fecha,
		Descripcion: row.Descripcion,
		Materia:     materiaResumen(row.MateriaID, row.MateriaNombre),
	}
	if row.Hora.Valid {
		evento.Hora = &row.Hora.String
	}
	return evento
}

type claseRow struct {
	ID            int64          `bun:"id"`
	Titulo        string         `bun:"titulo"`
	Dia           int            `bun:"dia"`
	HoraInicio    string         `bun:"hora_inicio"`
	HoraFin       string         `bun:"hora_fin"`
	Aula          string         `bun:"aula"`
	MateriaID     sql.NullInt64  `bun:"materia_id"`
	MateriaNombre sql.NullString `bun:"materia_nombre"`
	ComisionID    sql.NullInt64  `bun:"comision_id"`
}

func (row claseRow) toDomain() domain.Clase {
	var comisionID *int64
	if row.ComisionID.Valid {
		comisionID = &row.ComisionID.Int64
	}
	return domain.Clase{
		ComisionID: comisionID,
		ID:         row.ID,
		Titulo:     row.Titulo,
		Dia:        row.Dia,
		HoraInicio: row.HoraInicio,
		HoraFin:    row.HoraFin,
		Aula:       row.Aula,
		Materia:    materiaResumen(row.MateriaID, row.MateriaNombre),
	}
}

func materiaResumen(id sql.NullInt64, nombre sql.NullString) *domain.MateriaResumen {
	if !id.Valid {
		return nil
	}
	return &domain.MateriaResumen{ID: id.Int64, Nombre: nombre.String}
}

type calendarioRepository struct {
	db *bun.DB
}

func NewCalendarioRepository(db *bun.DB) domain.CalendarioRepository {
	return &calendarioRepository{db: db}
}

// Las horas se devuelven como HH:MM y las fechas como AAAA-MM-DD, que es lo
// que usa el frontend.
func (r *calendarioRepository) selectEventos(alumnoID int64) *bun.SelectQuery {
	return r.db.NewSelect().
		TableExpr("evento AS e").
		ColumnExpr("e.id, e.titulo, e.tipo, e.descripcion").
		ColumnExpr("to_char(e.fecha, 'YYYY-MM-DD') AS fecha").
		ColumnExpr("to_char(e.hora, 'HH24:MI') AS hora").
		ColumnExpr("m.id AS materia_id, m.nombre AS materia_nombre").
		Join("LEFT JOIN materia AS m ON m.id = e.materia_id").
		Where("e.alumno_id = ?", alumnoID)
}

func (r *calendarioRepository) selectClases(alumnoID int64) *bun.SelectQuery {
	return r.db.NewSelect().
		TableExpr("clase AS c").
		ColumnExpr("c.id, c.titulo, c.dia, c.aula, c.comision_id").
		ColumnExpr("to_char(c.hora_inicio, 'HH24:MI') AS hora_inicio").
		ColumnExpr("to_char(c.hora_fin, 'HH24:MI') AS hora_fin").
		ColumnExpr("m.id AS materia_id, m.nombre AS materia_nombre").
		Join("LEFT JOIN materia AS m ON m.id = c.materia_id").
		Where("c.alumno_id = ?", alumnoID)
}

func (r *calendarioRepository) ListEventos(ctx context.Context, alumnoID int64, desde, hasta string) ([]domain.Evento, error) {
	var rows []eventoRow
	err := r.selectEventos(alumnoID).
		Where("e.fecha BETWEEN ?::date AND ?::date", desde, hasta).
		OrderExpr("e.fecha ASC, e.hora ASC NULLS FIRST, e.id ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	eventos := make([]domain.Evento, len(rows))
	for i, row := range rows {
		eventos[i] = row.toDomain()
	}
	return eventos, nil
}

func (r *calendarioRepository) GetEvento(ctx context.Context, alumnoID, id int64) (*domain.Evento, error) {
	var rows []eventoRow
	if err := r.selectEventos(alumnoID).Where("e.id = ?", id).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrEventoNoEncontrado
	}
	evento := rows[0].toDomain()
	return &evento, nil
}

func (r *calendarioRepository) CreateEvento(ctx context.Context, alumnoID int64, datos domain.DatosEvento) (int64, error) {
	model := EventoModel{
		AlumnoID:    alumnoID,
		MateriaID:   datos.MateriaID,
		Titulo:      datos.Titulo,
		Tipo:        datos.Tipo,
		Fecha:       datos.Fecha,
		Hora:        datos.Hora,
		Descripcion: datos.Descripcion,
	}
	if _, err := r.db.NewInsert().Model(&model).Returning("id").Exec(ctx); err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *calendarioRepository) UpdateEvento(ctx context.Context, alumnoID, id int64, datos domain.DatosEvento) error {
	res, err := r.db.NewUpdate().
		Model((*EventoModel)(nil)).
		Set("titulo = ?", datos.Titulo).
		Set("tipo = ?", datos.Tipo).
		Set("fecha = ?", datos.Fecha).
		Set("hora = ?", datos.Hora).
		Set("descripcion = ?", datos.Descripcion).
		Set("materia_id = ?", datos.MateriaID).
		Where("id = ? AND alumno_id = ?", id, alumnoID).
		Exec(ctx)
	return filasAfectadas(res, err, domain.ErrEventoNoEncontrado)
}

func (r *calendarioRepository) DeleteEvento(ctx context.Context, alumnoID, id int64) error {
	res, err := r.db.NewDelete().
		Model((*EventoModel)(nil)).
		Where("id = ? AND alumno_id = ?", id, alumnoID).
		Exec(ctx)
	return filasAfectadas(res, err, domain.ErrEventoNoEncontrado)
}

func (r *calendarioRepository) ListClases(ctx context.Context, alumnoID int64) ([]domain.Clase, error) {
	var rows []claseRow
	err := r.selectClases(alumnoID).
		OrderExpr("c.dia ASC, c.hora_inicio ASC, c.id ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	clases := make([]domain.Clase, len(rows))
	for i, row := range rows {
		clases[i] = row.toDomain()
	}
	return clases, nil
}

func (r *calendarioRepository) GetClase(ctx context.Context, alumnoID, id int64) (*domain.Clase, error) {
	var rows []claseRow
	if err := r.selectClases(alumnoID).Where("c.id = ?", id).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrClaseNoEncontrada
	}
	clase := rows[0].toDomain()
	return &clase, nil
}

func (r *calendarioRepository) CreateClase(ctx context.Context, alumnoID int64, datos domain.DatosClase) (int64, error) {
	model := ClaseModel{
		AlumnoID:   alumnoID,
		MateriaID:  datos.MateriaID,
		ComisionID: datos.ComisionID,
		Titulo:     datos.Titulo,
		Dia:        datos.Dia,
		HoraInicio: datos.HoraInicio,
		HoraFin:    datos.HoraFin,
		Aula:       datos.Aula,
	}
	if _, err := r.db.NewInsert().Model(&model).Returning("id").Exec(ctx); err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *calendarioRepository) UpdateClase(ctx context.Context, alumnoID, id int64, datos domain.DatosClase) error {
	res, err := r.db.NewUpdate().
		Model((*ClaseModel)(nil)).
		Set("titulo = ?", datos.Titulo).
		Set("dia = ?", datos.Dia).
		Set("hora_inicio = ?", datos.HoraInicio).
		Set("hora_fin = ?", datos.HoraFin).
		Set("aula = ?", datos.Aula).
		Set("materia_id = ?", datos.MateriaID).
		Where("id = ? AND alumno_id = ?", id, alumnoID).
		Exec(ctx)
	return filasAfectadas(res, err, domain.ErrClaseNoEncontrada)
}

func (r *calendarioRepository) DeleteClase(ctx context.Context, alumnoID, id int64) error {
	res, err := r.db.NewDelete().
		Model((*ClaseModel)(nil)).
		Where("id = ? AND alumno_id = ?", id, alumnoID).
		Exec(ctx)
	return filasAfectadas(res, err, domain.ErrClaseNoEncontrada)
}

// filasAfectadas devuelve noEncontrado si la consulta no toco ninguna fila:
// el registro no existe o es de otro alumno.
func filasAfectadas(res sql.Result, err error, noEncontrado error) error {
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return noEncontrado
	}
	return nil
}

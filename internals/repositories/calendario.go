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
	ID            int64   `bun:"id,pk,autoincrement"`
	AlumnoID      int64   `bun:"alumno_id,notnull"`
	MateriaID     *int64  `bun:"materia_id"`
	ComisionID    *int64  `bun:"comision_id"`
	Titulo        string  `bun:"titulo,notnull"`
	Dia           int     `bun:"dia,notnull"`
	HoraInicio    string  `bun:"hora_inicio,notnull"`
	HoraFin       string  `bun:"hora_fin,notnull"`
	Aula          string  `bun:"aula,notnull"`
	Tipo          string  `bun:"tipo,notnull"`
	Hasta         *string `bun:"hasta,type:date"`
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
	Cuatrimestre  sql.NullString `bun:"cuatrimestre"`
	Tipo          string         `bun:"tipo"`
	Hasta         sql.NullString `bun:"hasta"`
}

func (row claseRow) toDomain() domain.Clase {
	var comisionID *int64
	if row.ComisionID.Valid {
		comisionID = &row.ComisionID.Int64
	}
	var cuatrimestre *string
	if row.Cuatrimestre.Valid {
		cuatrimestre = &row.Cuatrimestre.String
	}
	var hasta *string
	if row.Hasta.Valid {
		hasta = &row.Hasta.String
	}
	return domain.Clase{
		Tipo:         row.Tipo,
		Hasta:        hasta,
		ComisionID:   comisionID,
		Cuatrimestre: cuatrimestre,
		ID:           row.ID,
		Titulo:       row.Titulo,
		Dia:          row.Dia,
		HoraInicio:   row.HoraInicio,
		HoraFin:      row.HoraFin,
		Aula:         row.Aula,
		Materia:      materiaResumen(row.MateriaID, row.MateriaNombre),
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
		ColumnExpr("c.id, c.titulo, c.dia, c.aula, c.comision_id, c.tipo").
		ColumnExpr("to_char(c.hasta, 'YYYY-MM-DD') AS hasta").
		ColumnExpr("to_char(c.hora_inicio, 'HH24:MI') AS hora_inicio").
		ColumnExpr("to_char(c.hora_fin, 'HH24:MI') AS hora_fin").
		ColumnExpr("m.id AS materia_id, m.nombre AS materia_nombre").
		ColumnExpr("co.cuatrimestre").
		Join("LEFT JOIN materia AS m ON m.id = c.materia_id").
		Join("LEFT JOIN comision AS co ON co.id = c.comision_id").
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

func (r *calendarioRepository) ListFechasAcademicas(ctx context.Context, desde, hasta string) ([]domain.FechaAcademica, error) {
	fechas := []domain.FechaAcademica{}
	err := r.db.NewSelect().
		TableExpr("fecha_academica").
		ColumnExpr("id, titulo, tipo").
		ColumnExpr("to_char(desde, 'YYYY-MM-DD') AS desde").
		ColumnExpr("to_char(hasta, 'YYYY-MM-DD') AS hasta").
		Where("desde <= ?::date AND hasta >= ?::date", hasta, desde).
		OrderExpr("desde ASC, id ASC").
		Scan(ctx, &fechas)
	return fechas, err
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
		Tipo:       datos.Tipo,
		Hasta:      datos.Hasta,
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
		Set("tipo = ?", datos.Tipo).
		Set("hasta = ?", datos.Hasta).
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

func (r *calendarioRepository) CalendarioToken(ctx context.Context, alumnoID int64) (string, error) {
	var token sql.NullString
	err := r.db.NewSelect().
		TableExpr("alumno").
		ColumnExpr("calendario_token").
		Where("id = ?", alumnoID).
		Scan(ctx, &token)
	return token.String, err
}

func (r *calendarioRepository) SetCalendarioToken(ctx context.Context, alumnoID int64, token string) error {
	_, err := r.db.NewUpdate().
		TableExpr("alumno").
		Set("calendario_token = ?", token).
		Where("id = ?", alumnoID).
		Exec(ctx)
	return err
}

func (r *calendarioRepository) AlumnoDelCalendario(ctx context.Context, token string) (int64, error) {
	var ids []int64
	err := r.db.NewSelect().
		TableExpr("alumno").
		ColumnExpr("id").
		Where("calendario_token = ?", token).
		Scan(ctx, &ids)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, domain.ErrCalendarioNoEncontrado
	}
	return ids[0], nil
}

// queryEventosConfirmados arma los eventos confirmados por compañeros.
// Un alumno cuenta para una comision si tiene en su horario una clase de esa
// comision; para los finales (comision 0), si tiene la materia regular o la
// esta cursando. Solo votan y ven los eventos los que cuentan.
const queryEventosConfirmados = `
with yo as (select ?::int as id),
comision_de as (
    select distinct alumno_id, materia_id, comision_id from clase where comision_id is not null
    union
    select ca.alumno_id, ca.materia_id, 0
    from condicion_alumno ca join condicion co on co.id = ca.condicion_id
    where co.condicion in ('Regularizada', 'Cursando')
),
votos as (
    select distinct on (e.alumno_id, e.materia_id, e.tipo, e.fecha, cd.comision_id)
        e.alumno_id, e.materia_id, e.tipo, e.fecha, cd.comision_id, e.hora
    from evento e
    join comision_de cd on cd.alumno_id = e.alumno_id and cd.materia_id = e.materia_id
        and (e.tipo = 'final') = (cd.comision_id = 0)
    where e.tipo in ('parcial', 'entrega', 'final') and e.fecha >= ?::date
    order by e.alumno_id, e.materia_id, e.tipo, e.fecha, cd.comision_id, e.hora nulls last
),
grupos as (
    select materia_id, tipo, fecha, comision_id, count(*) as confirmaciones,
        mode() within group (order by hora) as hora,
        bool_or(alumno_id = (select id from yo)) as mio
    from votos
    group by materia_id, tipo, fecha, comision_id
),
desmentidos as (
    select d.materia_id, d.tipo, d.fecha, d.comision_id, count(*) as n,
        bool_or(d.alumno_id = (select id from yo)) as mio
    from evento_desmentido d
    join comision_de cd on cd.alumno_id = d.alumno_id and cd.materia_id = d.materia_id and cd.comision_id = d.comision_id
    group by d.materia_id, d.tipo, d.fecha, d.comision_id
)
select g.materia_id, m.nombre as materia_nombre, g.comision_id, g.tipo, g.confirmaciones,
    to_char(g.fecha, 'YYYY-MM-DD') as fecha, to_char(g.hora, 'HH24:MI') as hora
from grupos g
join materia m on m.id = g.materia_id
join comision_de mia on mia.alumno_id = (select id from yo) and mia.materia_id = g.materia_id and mia.comision_id = g.comision_id
left join desmentidos d on d.materia_id = g.materia_id and d.tipo = g.tipo and d.fecha = g.fecha and d.comision_id = g.comision_id
where g.confirmaciones >= ? and not g.mio and not coalesce(d.mio, false)
    and coalesce(d.n, 0) < g.confirmaciones
order by g.fecha, g.hora nulls first, m.nombre`

func (r *calendarioRepository) ListEventosConfirmados(ctx context.Context, alumnoID int64, desde string, minimo int) ([]domain.EventoConfirmado, error) {
	var rows []struct {
		MateriaID      int64          `bun:"materia_id"`
		MateriaNombre  string         `bun:"materia_nombre"`
		ComisionID     int64          `bun:"comision_id"`
		Tipo           string         `bun:"tipo"`
		Confirmaciones int            `bun:"confirmaciones"`
		Fecha          string         `bun:"fecha"`
		Hora           sql.NullString `bun:"hora"`
	}
	if err := r.db.NewRaw(queryEventosConfirmados, alumnoID, desde, minimo).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	eventos := make([]domain.EventoConfirmado, len(rows))
	for i, row := range rows {
		eventos[i] = domain.EventoConfirmado{
			Materia:        domain.MateriaResumen{ID: row.MateriaID, Nombre: row.MateriaNombre},
			ComisionID:     row.ComisionID,
			Tipo:           row.Tipo,
			Fecha:          row.Fecha,
			Confirmaciones: row.Confirmaciones,
		}
		if row.Hora.Valid {
			eventos[i].Hora = &row.Hora.String
		}
	}
	return eventos, nil
}

func (r *calendarioRepository) DesmentirEvento(ctx context.Context, alumnoID int64, d domain.DatosDesmentido) error {
	_, err := r.db.ExecContext(ctx, `
		insert into evento_desmentido (alumno_id, materia_id, comision_id, tipo, fecha)
		values (?, ?, ?, ?, ?)
		on conflict do nothing`, alumnoID, d.MateriaID, d.ComisionID, d.Tipo, d.Fecha)
	return err
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

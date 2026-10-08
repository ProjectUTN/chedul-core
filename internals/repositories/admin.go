package repositories

import (
	"chedul-core/internals/domain"
	"context"
	"strings"

	"github.com/uptrace/bun"
)

type adminRepository struct {
	db *bun.DB
}

func NewAdminRepository(db *bun.DB) domain.AdminRepository {
	return &adminRepository{db: db}
}

// Los dias se cuentan en hora de Argentina
const zonaAdmin = "America/Argentina/Buenos_Aires"

func (r *adminRepository) Resumen(ctx context.Context) (*domain.ResumenAdmin, error) {
	var res domain.ResumenAdmin
	err := r.db.QueryRowContext(ctx, `
select
    (select count(*) from alumno),
    (select count(*) from alumno where creado >= now() - interval '7 days'),
    (select count(*) from alumno where creado >= now() - interval '30 days'),
    (select count(*) from alumno where ultimo_acceso >= now() - interval '24 hours'),
    (select count(*) from alumno where ultimo_acceso >= now() - interval '7 days'),
    (select count(*) from alumno where google_vinculado),
    (select count(distinct alumno_id) from condicion_alumno),
    (select count(*) from aporte),
    (select count(*) from comunidad),
    (select count(*) from acceso where (fecha at time zone ?)::date = (now() at time zone ?)::date)`,
		zonaAdmin, zonaAdmin).Scan(
		&res.Alumnos, &res.Nuevos7, &res.Nuevos30, &res.Activos24h, &res.Activos7,
		&res.ConGoogle, &res.ConEstado, &res.Aportes, &res.Comunidades, &res.SesionesHoy)
	if err != nil {
		return nil, err
	}

	if res.RegistrosPorDia, err = r.porDia(ctx, "alumno", "creado"); err != nil {
		return nil, err
	}
	if res.AccesosPorDia, err = r.porDia(ctx, "acceso", "fecha"); err != nil {
		return nil, err
	}
	return &res, nil
}

// porDia cuenta filas por dia en los ultimos 30 dias, incluidos los dias en 0.
// tabla y columna son constantes de este archivo, nunca datos del usuario.
func (r *adminRepository) porDia(ctx context.Context, tabla, columna string) ([]domain.CantidadPorDia, error) {
	var filas []domain.CantidadPorDia
	err := r.db.NewRaw(`
select to_char(d.dia, 'YYYY-MM-DD') as dia, count(t.*) as cantidad
from generate_series((now() at time zone ?)::date - 29, (now() at time zone ?)::date, interval '1 day') as d(dia)
left join `+tabla+` t on (t.`+columna+` at time zone ?)::date = d.dia
group by d.dia
order by d.dia`, zonaAdmin, zonaAdmin, zonaAdmin).Scan(ctx, &filas)
	return filas, err
}

func (r *adminRepository) Alumnos(ctx context.Context, buscar string, pagina int) ([]domain.AlumnoAdmin, int, error) {
	if pagina < 1 {
		pagina = 1
	}
	var filas []domain.AlumnoAdmin
	q := r.db.NewSelect().
		TableExpr("alumno AS a").
		ColumnExpr("a.id, a.nombre, a.email, a.google_vinculado AS google").
		ColumnExpr(`to_char(a.creado AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS creado`).
		ColumnExpr(`to_char(a.ultimo_acceso AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS ultimo_acceso`).
		ColumnExpr("(select count(*) from condicion_alumno ca where ca.alumno_id = a.id) AS materias").
		ColumnExpr("(select count(*) from aporte ap where ap.alumno_id = a.id) AS aportes")
	if texto := strings.TrimSpace(buscar); texto != "" {
		patron := "%" + escaparLike(texto) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("a.nombre ILIKE ?", patron).WhereOr("a.email ILIKE ?", patron)
		})
	}
	total, err := q.
		OrderExpr("a.creado DESC NULLS LAST, a.id DESC").
		Limit(domain.AlumnosAdminPorPagina).
		Offset((pagina-1)*domain.AlumnosAdminPorPagina).
		ScanAndCount(ctx, &filas)
	if err != nil {
		return nil, 0, err
	}
	if filas == nil {
		filas = []domain.AlumnoAdmin{}
	}
	return filas, total, nil
}

func (r *adminRepository) Accesos(ctx context.Context, limite int) ([]domain.AccesoAdmin, error) {
	var filas []domain.AccesoAdmin
	err := r.db.NewRaw(`
select to_char(ac.fecha AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as fecha, ac.metodo,
    a.id as alumno_id, a.nombre, a.email
from acceso ac join alumno a on a.id = ac.alumno_id
order by ac.fecha desc
limit ?`, limite).Scan(ctx, &filas)
	if filas == nil {
		filas = []domain.AccesoAdmin{}
	}
	return filas, err
}

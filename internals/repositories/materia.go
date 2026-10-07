package repositories

import (
	"chedul-core/internals/domain"
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

type MateriaModel struct {
	bun.BaseModel  `bun:"table:materia,alias:materia"`
	ID             int64           `bun:"id,pk,autoincrement"`
	Codigo         sql.NullString  `bun:"codigo"`
	Nombre         string          `bun:"nombre,notnull"`
	CargaHoraria   int64           `bun:"carga_horaria,notnull"`
	Nivel          int64           `bun:"nivel,notnull"`
	Area           string          `bun:"area,notnull"`
	Tipo           string          `bun:"tipo,notnull"`
	CuatrimestreID sql.NullInt64   `bun:"cuatrimestre_id"`
	Cuatrimestre   sql.NullString  `bun:"cuatrimestre,scanonly"`
	Horas          sql.NullFloat64 `bun:"horas"`
	Bloque         string          `bun:"bloque,notnull"`
	Programa       string          `bun:"programa,notnull"`
}

type correlativaRow struct {
	MateriaID  int64  `bun:"materia_id"`
	RequiereID int64  `bun:"requiere_id"`
	Nombre     string `bun:"nombre"`
	Tipo       string `bun:"tipo"`
}

type materiaRepository struct {
	db *bun.DB
}

func (r *materiaRepository) toDomain(model MateriaModel) domain.Materia {
	return domain.Materia{
		ID:           model.ID,
		Codigo:       model.Codigo.String,
		Nombre:       model.Nombre,
		CargaHoraria: model.CargaHoraria,
		Nivel:        model.Nivel,
		Area:         model.Area,
		Tipo:         model.Tipo,
		Cuatrimestre: model.Cuatrimestre.String,
		Horas:        model.Horas.Float64,
		Bloque:       model.Bloque,
		Programa:     model.Programa,
		Correlativas: []domain.Correlativa{},
	}
}

func NewMateriaRepository(db *bun.DB) domain.MateriaRepository {
	return &materiaRepository{db: db}
}

func (r *materiaRepository) baseQuery(models *[]MateriaModel) *bun.SelectQuery {
	return r.db.NewSelect().
		Model(models).
		ColumnExpr("materia.*").
		ColumnExpr("cu.nombre AS cuatrimestre").
		Join("LEFT JOIN cuatrimestre AS cu ON cu.id = materia.cuatrimestre_id").
		OrderExpr("materia.nivel, materia.nombre")
}

// conCorrelativas carga en una sola consulta las correlativas de todas las materias.
func (r *materiaRepository) conCorrelativas(ctx context.Context, models []MateriaModel) ([]domain.Materia, error) {
	result := make([]domain.Materia, len(models))
	if len(models) == 0 {
		return result, nil
	}

	ids := make([]int64, len(models))
	index := make(map[int64]int, len(models))
	for i, m := range models {
		result[i] = r.toDomain(m)
		ids[i] = m.ID
		index[m.ID] = i
	}

	var rows []correlativaRow
	err := r.db.NewSelect().
		TableExpr("correlativa AS c").
		ColumnExpr("c.materia_id, c.requiere_id, c.tipo, m.nombre").
		Join("JOIN materia AS m ON m.id = c.requiere_id").
		Where("c.materia_id IN (?)", bun.In(ids)).
		OrderExpr("m.nivel, m.nombre").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		i := index[row.MateriaID]
		result[i].Correlativas = append(result[i].Correlativas, domain.Correlativa{
			MateriaID: row.RequiereID,
			Nombre:    row.Nombre,
			Tipo:      row.Tipo,
		})
	}

	return result, nil
}

func (r *materiaRepository) GetAll(ctx context.Context) ([]domain.Materia, error) {
	var models []MateriaModel
	if err := r.baseQuery(&models).Scan(ctx); err != nil {
		return nil, err
	}

	return r.conCorrelativas(ctx, models)
}

func (r *materiaRepository) GetByCarrera(ctx context.Context, carreraID int64) ([]domain.Materia, error) {
	var models []MateriaModel
	err := r.baseQuery(&models).
		Join("JOIN materiasPorcarrera AS mpc ON mpc.materia_id = materia.id").
		Where("mpc.carrera_id = ?", carreraID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return r.conCorrelativas(ctx, models)
}

func (r *materiaRepository) GetByID(ctx context.Context, id int64) (*domain.Materia, error) {
	var models []MateriaModel
	if err := r.baseQuery(&models).Where("materia.id = ?", id).Scan(ctx); err != nil {
		return nil, err
	}
	if len(models) == 0 {
		return nil, sql.ErrNoRows
	}

	materias, err := r.conCorrelativas(ctx, models)
	if err != nil {
		return nil, err
	}

	return &materias[0], nil
}

type comisionRow struct {
	ID           int64          `bun:"id"`
	Codigo       string         `bun:"codigo"`
	Cuatrimestre string         `bun:"cuatrimestre"`
	Dia          sql.NullInt64  `bun:"dia"`
	HoraInicio   sql.NullString `bun:"hora_inicio"`
	HoraFin      sql.NullString `bun:"hora_fin"`
	Aula         sql.NullString `bun:"aula"`
}

// GetComisiones devuelve las comisiones de la materia con sus horarios. El dia
// se calcula por nombre (1 = lunes) para no depender de los ids de la tabla dias.
func (r *materiaRepository) GetComisiones(ctx context.Context, materiaID int64) ([]domain.Comision, error) {
	var rows []comisionRow
	err := r.db.NewSelect().
		TableExpr("comision AS co").
		ColumnExpr("co.id, coalesce(co.codigo, '') AS codigo, co.cuatrimestre").
		ColumnExpr("array_position(array['Lunes','Martes','Miércoles','Jueves','Viernes','Sábado','Domingo'], d.nombre) AS dia").
		ColumnExpr("to_char(h.hora_inicio, 'HH24:MI') AS hora_inicio").
		ColumnExpr("to_char(h.hora_fin, 'HH24:MI') AS hora_fin").
		ColumnExpr("h.aula").
		Join("LEFT JOIN horario AS h ON h.comision_id = co.id").
		Join("LEFT JOIN dias AS d ON d.id = h.dia_id").
		Where("co.materia_id = ?", materiaID).
		OrderExpr("co.codigo, co.cuatrimestre, dia, h.hora_inicio").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	result := []domain.Comision{}
	index := map[int64]int{}
	for _, row := range rows {
		i, ok := index[row.ID]
		if !ok {
			i = len(result)
			index[row.ID] = i
			result = append(result, domain.Comision{
				ID:           row.ID,
				Codigo:       row.Codigo,
				Cuatrimestre: row.Cuatrimestre,
				Horarios:     []domain.HorarioComision{},
			})
		}
		if row.Dia.Valid {
			result[i].Horarios = append(result[i].Horarios, domain.HorarioComision{
				Dia:        int(row.Dia.Int64),
				HoraInicio: row.HoraInicio.String,
				HoraFin:    row.HoraFin.String,
				Aula:       row.Aula.String,
			})
		}
	}

	return result, nil
}

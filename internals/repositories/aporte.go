package repositories

import (
	"chedul-core/internals/domain"
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/uptrace/bun"
)

type AporteModel struct {
	bun.BaseModel `bun:"table:aporte"`
	ID            int64          `bun:"id,pk,autoincrement"`
	MateriaID     int64          `bun:"materia_id,notnull"`
	AlumnoID      int64          `bun:"alumno_id,notnull"`
	Titulo        string         `bun:"titulo,notnull"`
	Descripcion   sql.NullString `bun:"descripcion"`
	TagID         int64          `bun:"tag_id,notnull"`
	Link          *string        `bun:"link"`
	ArchivoKey    *string        `bun:"archivo_key"`
	ArchivoNombre *string        `bun:"archivo_nombre"`
	ArchivoTipo   *string        `bun:"archivo_tipo"`
	ArchivoTamano *int64         `bun:"archivo_tamano"`
}

type AporteTagModel struct {
	bun.BaseModel `bun:"table:aporte_tag"`
	ID            int64  `bun:"id,pk,autoincrement"`
	Nombre        string `bun:"nombre,notnull"`
}

type AporteFavoritoModel struct {
	bun.BaseModel `bun:"table:aportes_favoritos"`
	ID            int64 `bun:"id,pk,autoincrement"`
	AlumnoID      int64 `bun:"alumno_id,notnull"`
	AporteID      int64 `bun:"aporte_id,notnull"`
}

// aporteRow es una fila del listado, con los joins ya resueltos.
type aporteRow struct {
	ID            int64          `bun:"id"`
	Titulo        string         `bun:"titulo"`
	Descripcion   sql.NullString `bun:"descripcion"`
	Link          *string        `bun:"link"`
	CreadoEn      time.Time      `bun:"creado_en"`
	ArchivoKey    *string        `bun:"archivo_key"`
	ArchivoNombre *string        `bun:"archivo_nombre"`
	ArchivoTipo   *string        `bun:"archivo_tipo"`
	ArchivoTamano *int64         `bun:"archivo_tamano"`
	MateriaID     int64          `bun:"materia_id"`
	MateriaNombre string         `bun:"materia_nombre"`
	MateriaNivel  int64          `bun:"materia_nivel"`
	TagID         int64          `bun:"tag_id"`
	TagNombre     string         `bun:"tag_nombre"`
	AutorID       int64          `bun:"autor_id"`
	AutorNombre   string         `bun:"autor_nombre"`
	Favoritos     int            `bun:"favoritos"`
	EsFavorito    bool           `bun:"es_favorito"`
	Total         int            `bun:"total"`
}

func (row aporteRow) toDomain(viewerID int64) domain.Aporte {
	aporte := domain.Aporte{
		ID:          row.ID,
		Titulo:      row.Titulo,
		Descripcion: row.Descripcion.String,
		Link:        row.Link,
		CreadoEn:    row.CreadoEn,
		Materia:     domain.AporteMateria{ID: row.MateriaID, Nombre: row.MateriaNombre, Nivel: row.MateriaNivel},
		Tag:         domain.AporteTag{ID: row.TagID, Nombre: row.TagNombre},
		Autor:       domain.AporteAutor{ID: row.AutorID, Nombre: row.AutorNombre},
		Favoritos:   row.Favoritos,
		EsFavorito:  row.EsFavorito,
		EsMio:       row.AutorID == viewerID,
	}

	if row.ArchivoKey != nil {
		aporte.Archivo = &domain.ArchivoAporte{Key: *row.ArchivoKey}
		if row.ArchivoNombre != nil {
			aporte.Archivo.Nombre = *row.ArchivoNombre
		}
		if row.ArchivoTipo != nil {
			aporte.Archivo.Tipo = *row.ArchivoTipo
		}
		if row.ArchivoTamano != nil {
			aporte.Archivo.Tamano = *row.ArchivoTamano
		}
	}

	return aporte
}

type aporteRepository struct {
	db *bun.DB
}

func NewAporteRepository(db *bun.DB) domain.AporteRepository {
	return &aporteRepository{db: db}
}

func (r *aporteRepository) selectAportes(viewerID int64) *bun.SelectQuery {
	return r.db.NewSelect().
		TableExpr("aporte AS a").
		ColumnExpr("a.id, a.titulo, a.descripcion, a.link, a.creado_en").
		ColumnExpr("a.archivo_key, a.archivo_nombre, a.archivo_tipo, a.archivo_tamano").
		ColumnExpr("m.id AS materia_id, m.nombre AS materia_nombre, m.nivel AS materia_nivel").
		ColumnExpr("t.id AS tag_id, t.nombre AS tag_nombre").
		ColumnExpr("al.id AS autor_id, al.nombre AS autor_nombre").
		ColumnExpr("(SELECT count(*) FROM aportes_favoritos f WHERE f.aporte_id = a.id) AS favoritos").
		ColumnExpr("EXISTS (SELECT 1 FROM aportes_favoritos f WHERE f.aporte_id = a.id AND f.alumno_id = ?) AS es_favorito", viewerID).
		Join("JOIN materia AS m ON m.id = a.materia_id").
		Join("JOIN aporte_tag AS t ON t.id = a.tag_id").
		Join("JOIN alumno AS al ON al.id = a.alumno_id")
}

// escaparLike evita que % y _ escritos por el usuario funcionen como comodines.
func escaparLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (r *aporteRepository) List(ctx context.Context, filtro domain.AporteFiltro) ([]domain.Aporte, int, error) {
	q := r.selectAportes(filtro.ViewerID).
		ColumnExpr("count(*) OVER () AS total")

	if filtro.MateriaID > 0 {
		q = q.Where("a.materia_id = ?", filtro.MateriaID)
	}
	if filtro.TagID > 0 {
		q = q.Where("a.tag_id = ?", filtro.TagID)
	}
	if filtro.AutorID > 0 {
		q = q.Where("a.alumno_id = ?", filtro.AutorID)
	}
	if filtro.SoloFavoritosDe > 0 {
		q = q.Where("EXISTS (SELECT 1 FROM aportes_favoritos f WHERE f.aporte_id = a.id AND f.alumno_id = ?)", filtro.SoloFavoritosDe)
	}
	if texto := strings.TrimSpace(filtro.Texto); texto != "" {
		patron := "%" + escaparLike(texto) + "%"
		q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("a.titulo ILIKE ?", patron).
				WhereOr("a.descripcion ILIKE ?", patron).
				WhereOr("m.nombre ILIKE ?", patron)
		})
	}

	if filtro.Orden == domain.OrdenPopulares {
		q = q.OrderExpr("favoritos DESC, a.creado_en DESC, a.id DESC")
	} else {
		q = q.OrderExpr("a.creado_en DESC, a.id DESC")
	}

	var rows []aporteRow
	if err := q.Limit(filtro.Limit).Offset(filtro.Offset).Scan(ctx, &rows); err != nil {
		return nil, 0, err
	}

	aportes := make([]domain.Aporte, len(rows))
	total := 0
	for i, row := range rows {
		aportes[i] = row.toDomain(filtro.ViewerID)
		total = row.Total
	}

	// Si la pagina pedida esta vacia no tenemos el total desde la ventana
	if len(rows) == 0 && filtro.Offset > 0 {
		filtro.Offset = 0
		filtro.Limit = 1
		_, total, err := r.List(ctx, filtro)
		return aportes, total, err
	}

	return aportes, total, nil
}

func (r *aporteRepository) GetByID(ctx context.Context, id, viewerID int64) (*domain.Aporte, error) {
	var rows []aporteRow
	if err := r.selectAportes(viewerID).Where("a.id = ?", id).Scan(ctx, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, domain.ErrAporteNoEncontrado
	}

	aporte := rows[0].toDomain(viewerID)
	return &aporte, nil
}

func (r *aporteRepository) Create(ctx context.Context, autorID int64, datos domain.DatosAporte, archivo *domain.ArchivoAporte) (int64, error) {
	model := AporteModel{
		MateriaID:   datos.MateriaID,
		AlumnoID:    autorID,
		Titulo:      datos.Titulo,
		Descripcion: sql.NullString{String: datos.Descripcion, Valid: datos.Descripcion != ""},
		TagID:       datos.TagID,
		Link:        datos.Link,
	}

	if archivo != nil {
		model.ArchivoKey = &archivo.Key
		model.ArchivoNombre = &archivo.Nombre
		model.ArchivoTipo = &archivo.Tipo
		model.ArchivoTamano = &archivo.Tamano
	}

	_, err := r.db.NewInsert().Model(&model).Returning("id").Exec(ctx)
	if err != nil {
		return 0, err
	}

	return model.ID, nil
}

func (r *aporteRepository) Update(ctx context.Context, id int64, datos domain.DatosAporte) error {
	res, err := r.db.NewUpdate().
		Model((*AporteModel)(nil)).
		Set("titulo = ?", datos.Titulo).
		Set("descripcion = ?", sql.NullString{String: datos.Descripcion, Valid: datos.Descripcion != ""}).
		Set("link = ?", datos.Link).
		Set("materia_id = ?", datos.MateriaID).
		Set("tag_id = ?", datos.TagID).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrAporteNoEncontrado
	}
	return nil
}

func (r *aporteRepository) Delete(ctx context.Context, id int64) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewDelete().Model((*AporteFavoritoModel)(nil)).Where("aporte_id = ?", id).Exec(ctx); err != nil {
			return err
		}

		res, err := tx.NewDelete().Model((*AporteModel)(nil)).Where("id = ?", id).Exec(ctx)
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return domain.ErrAporteNoEncontrado
		}
		return nil
	})
}

func (r *aporteRepository) SetFavorito(ctx context.Context, aporteID, alumnoID int64, favorito bool) error {
	if !favorito {
		_, err := r.db.NewDelete().
			Model((*AporteFavoritoModel)(nil)).
			Where("aporte_id = ? AND alumno_id = ?", aporteID, alumnoID).
			Exec(ctx)
		return err
	}

	model := AporteFavoritoModel{AporteID: aporteID, AlumnoID: alumnoID}
	_, err := r.db.NewInsert().
		Model(&model).
		ExcludeColumn("id").
		On("CONFLICT (alumno_id, aporte_id) DO NOTHING").
		Exec(ctx)
	return err
}

func (r *aporteRepository) Tags(ctx context.Context) ([]domain.AporteTag, error) {
	var models []AporteTagModel
	if err := r.db.NewSelect().Model(&models).Order("id ASC").Scan(ctx); err != nil {
		return nil, err
	}

	tags := make([]domain.AporteTag, len(models))
	for i, m := range models {
		tags[i] = domain.AporteTag{ID: m.ID, Nombre: m.Nombre}
	}
	return tags, nil
}

func (r *aporteRepository) TagExiste(ctx context.Context, id int64) (bool, error) {
	return r.db.NewSelect().Model((*AporteTagModel)(nil)).Where("id = ?", id).Exists(ctx)
}

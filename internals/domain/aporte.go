package domain

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/rivo/uniseg"
)

var ErrAporteNoEncontrado = errors.New("aporte no encontrado")

type AporteTag struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

type AporteMateria struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
	Nivel  int64  `json:"nivel"`
}

type AporteAutor struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
}

type ArchivoAporte struct {
	Key    string `json:"-"`
	Nombre string `json:"nombre"`
	Tipo   string `json:"tipo"`
	Tamano int64  `json:"tamano"`
}

// Aporte es un material compartido por un alumno para una materia (o para
// toda la carrera): un resumen, un parcial resuelto, un link a un video, etc.
type Aporte struct {
	ID          int64     `json:"id"`
	Titulo      string    `json:"titulo"`
	Descripcion string    `json:"descripcion"`
	Link        *string   `json:"link"`
	CreadoEn    time.Time `json:"creado_en"`
	// nil en los aportes de toda la carrera
	Materia    *AporteMateria `json:"materia"`
	Tag        AporteTag      `json:"tag"`
	Autor      AporteAutor    `json:"autor"`
	Archivo    *ArchivoAporte `json:"archivo"`
	Favoritos  int            `json:"favoritos"`
	EsFavorito bool           `json:"es_favorito"`
	EsMio      bool           `json:"es_mio"`
}

const (
	OrdenRecientes = "recientes"
	OrdenPopulares = "populares"
)

type AporteFiltro struct {
	// Alumno que hace la consulta, para calcular es_favorito y es_mio
	ViewerID  int64
	MateriaID int64
	// Solo los de toda la carrera (sin materia)
	SoloCarrera     bool
	TagID           int64
	Texto           string
	AutorID         int64
	SoloFavoritosDe int64
	Orden           string
	Limit           int
	Offset          int
}

// DatosAporte son los campos que el alumno completa al subir o editar un aporte.
type DatosAporte struct {
	Titulo      string  `json:"titulo"`
	Descripcion string  `json:"descripcion"`
	Link        *string `json:"link"`
	// 0 es un aporte de toda la carrera
	MateriaID int64 `json:"materia_id"`
	TagID     int64 `json:"tag_id"`
}

const (
	TituloMaxLargo      = 100
	DescripcionMaxLargo = 2000
)

// Normalizar recorta espacios y convierte un link vacio en nil.
func (d *DatosAporte) Normalizar() {
	if d.MateriaID < 0 {
		d.MateriaID = 0
	}
	d.Titulo = strings.TrimSpace(d.Titulo)
	d.Descripcion = strings.TrimSpace(d.Descripcion)
	if d.Link != nil {
		link := strings.TrimSpace(*d.Link)
		if link == "" {
			d.Link = nil
		} else {
			d.Link = &link
		}
	}
}

// Validate valida los datos. tieneArchivo indica si el aporte tiene (o va a
// tener) un archivo subido: sin archivo, el link es obligatorio.
func (d *DatosAporte) Validate(tieneArchivo bool) map[string]string {
	errs := make(map[string]string)

	largoTitulo := uniseg.GraphemeClusterCount(d.Titulo)
	if largoTitulo == 0 {
		errs["titulo"] = "El título es requerido"
	} else if largoTitulo > TituloMaxLargo {
		errs["titulo"] = "El título no puede tener más de 100 caracteres"
	}

	if uniseg.GraphemeClusterCount(d.Descripcion) > DescripcionMaxLargo {
		errs["descripcion"] = "La descripción no puede tener más de 2000 caracteres"
	}

	if d.TagID <= 0 {
		errs["tag_id"] = "El tipo de aporte es requerido"
	}

	if d.Link != nil {
		u, err := url.ParseRequestURI(*d.Link)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			errs["link"] = "El link tiene que empezar con http:// o https://"
		}
	} else if !tieneArchivo {
		errs["link"] = "Subí un archivo o agregá un link"
	}

	return errs
}

type AporteRepository interface {
	List(ctx context.Context, filtro AporteFiltro) ([]Aporte, int, error)
	GetByID(ctx context.Context, id, viewerID int64) (*Aporte, error)
	Create(ctx context.Context, autorID int64, datos DatosAporte, archivo *ArchivoAporte) (int64, error)
	Update(ctx context.Context, id int64, datos DatosAporte) error
	Delete(ctx context.Context, id int64) error
	SetFavorito(ctx context.Context, aporteID, alumnoID int64, favorito bool) error
	Tags(ctx context.Context) ([]AporteTag, error)
	TagExiste(ctx context.Context, id int64) (bool, error)
}

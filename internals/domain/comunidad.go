package domain

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/rivo/uniseg"
)

var (
	ErrComunidadNoEncontrada = errors.New("comunidad no encontrada")
	ErrComunidadRepetida     = errors.New("ya hay una comunidad con ese link")
)

const (
	ComunidadNombreMax      = 80
	ComunidadDescripcionMax = 300
	ComunidadLinkMax        = 500
	// Con esta cantidad de reportes de distintos alumnos la comunidad se oculta
	ComunidadReportesMax = 3
)

// Comunidad es un grupo de la carrera (WhatsApp, Discord...) que cargo un alumno.
type Comunidad struct {
	ID          int64           `json:"id"`
	Nombre      string          `json:"nombre"`
	Descripcion string          `json:"descripcion"`
	Plataforma  string          `json:"plataforma"`
	Link        string          `json:"link"`
	Creada      string          `json:"creada"`
	Materia     *MateriaResumen `json:"materia"`
	EsMia       bool            `json:"es_mia"`
	Reportada   bool            `json:"reportada"`
}

type DatosComunidad struct {
	Nombre      string `json:"nombre"`
	Descripcion string `json:"descripcion"`
	Link        string `json:"link"`
	MateriaID   *int64 `json:"materia_id"`
	// La plataforma sale del link, no la elige el alumno
	Plataforma string `json:"-"`
}

// plataformas por dominio del link
var plataformas = map[string]string{
	"chat.whatsapp.com": "whatsapp",
	"whatsapp.com":      "whatsapp",
	"wa.me":             "whatsapp",
	"discord.gg":        "discord",
	"discord.com":       "discord",
	"t.me":              "telegram",
	"telegram.me":       "telegram",
	"instagram.com":     "instagram",
}

// PlataformaDeLink reconoce de que app es el link; lo que no conoce es "otra"
func PlataformaDeLink(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return "otra"
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	if p, ok := plataformas[host]; ok {
		return p
	}
	return "otra"
}

func (d *DatosComunidad) Normalizar() {
	d.Nombre = strings.TrimSpace(d.Nombre)
	d.Descripcion = strings.TrimSpace(d.Descripcion)
	d.Link = strings.TrimSpace(d.Link)
	d.MateriaID = normalizarMateria(d.MateriaID)
	d.Plataforma = PlataformaDeLink(d.Link)
}

func (d *DatosComunidad) Validate() map[string]string {
	errs := make(map[string]string)
	largo := uniseg.GraphemeClusterCount(d.Nombre)
	if largo == 0 {
		errs["nombre"] = "Poné un nombre"
	} else if largo > ComunidadNombreMax {
		errs["nombre"] = "Máximo 80 caracteres"
	}
	if uniseg.GraphemeClusterCount(d.Descripcion) > ComunidadDescripcionMax {
		errs["descripcion"] = "Máximo 300 caracteres"
	}
	u, err := url.ParseRequestURI(d.Link)
	if d.Link == "" {
		errs["link"] = "Pegá el link para unirse"
	} else if err != nil || u.Scheme != "https" || u.Host == "" {
		errs["link"] = "El link tiene que empezar con https://"
	} else if len(d.Link) > ComunidadLinkMax {
		errs["link"] = "El link es demasiado largo"
	}
	return errs
}

type ComunidadRepository interface {
	// List devuelve las que tienen menos de ComunidadReportesMax reportes
	List(ctx context.Context, viewerID int64) ([]Comunidad, error)
	Get(ctx context.Context, id, viewerID int64) (*Comunidad, error)
	Create(ctx context.Context, alumnoID int64, datos DatosComunidad) (int64, error)
	// Delete borra solo si es del alumno
	Delete(ctx context.Context, id, alumnoID int64) error
	Reportar(ctx context.Context, id, alumnoID int64) error
}

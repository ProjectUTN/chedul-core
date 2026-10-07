package handlers

import (
	"chedul-core/internals/domain"
	"chedul-core/pkg/storage"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Tipos de archivo que se pueden subir como aporte, por extension.
var tiposPermitidos = map[string]string{
	".pdf":  "application/pdf",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".webp": "image/webp",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".txt":  "text/plain",
	".md":   "text/markdown",
	".zip":  "application/zip",
}

const (
	aportesPorPaginaDefault = 20
	aportesPorPaginaMax     = 50
)

type AporteHandler struct {
	repo           domain.AporteRepository
	materiaRepo    domain.MateriaRepository
	storage        storage.Storage
	maxBytes       int64
	subidaArchivos bool
	logger         *zap.Logger
}

func NewAporteHandler(repo domain.AporteRepository, materiaRepo domain.MateriaRepository, storage storage.Storage, maxSizeMB int64, subidaArchivos bool, logger *zap.Logger) *AporteHandler {
	return &AporteHandler{
		repo:           repo,
		materiaRepo:    materiaRepo,
		storage:        storage,
		maxBytes:       maxSizeMB * 1024 * 1024,
		subidaArchivos: subidaArchivos,
		logger:         logger,
	}
}

type ConfigAportesResponse struct {
	SubidaArchivos bool  `json:"subida_archivos"`
	MaxMB          int64 `json:"max_mb"`
}

// Config le dice al frontend si se pueden subir archivos o solo links.
func (h *AporteHandler) Config(c echo.Context) error {
	return c.JSON(http.StatusOK, ConfigAportesResponse{
		SubidaArchivos: h.subidaArchivos,
		MaxMB:          h.maxBytes / 1024 / 1024,
	})
}

type ListaAportesResponse struct {
	Items  []domain.Aporte `json:"items"`
	Total  int             `json:"total"`
	Pagina int             `json:"pagina"`
	Limite int             `json:"limite"`
}

func queryInt(c echo.Context, name string) (int64, error) {
	v := c.QueryParam(name)
	if v == "" {
		return 0, nil
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		return 0, InvalidRequestData(map[string]string{name: "Debe ser un número positivo"})
	}
	return n, nil
}

// List devuelve los aportes paginados. Filtros por query string:
// materia_id, tag_id, q (texto), mios=1, favoritos=1, orden=recientes|populares,
// pagina (desde 1) y limite.
func (h *AporteHandler) List(c echo.Context) error {
	ctx := c.Request().Context()

	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	materiaID, err := queryInt(c, "materia_id")
	if err != nil {
		return err
	}
	tagID, err := queryInt(c, "tag_id")
	if err != nil {
		return err
	}
	pagina, err := queryInt(c, "pagina")
	if err != nil {
		return err
	}
	limite, err := queryInt(c, "limite")
	if err != nil {
		return err
	}

	if pagina < 1 {
		pagina = 1
	}
	if limite < 1 {
		limite = aportesPorPaginaDefault
	}
	if limite > aportesPorPaginaMax {
		limite = aportesPorPaginaMax
	}

	filtro := domain.AporteFiltro{
		ViewerID:  alumnoID,
		MateriaID: materiaID,
		TagID:     tagID,
		Texto:     c.QueryParam("q"),
		Orden:     c.QueryParam("orden"),
		Limit:     int(limite),
		Offset:    int((pagina - 1) * limite),
	}
	if c.QueryParam("mios") == "1" {
		filtro.AutorID = alumnoID
	}
	if c.QueryParam("favoritos") == "1" {
		filtro.SoloFavoritosDe = alumnoID
	}

	aportes, total, err := h.repo.List(ctx, filtro)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ListaAportesResponse{
		Items:  aportes,
		Total:  total,
		Pagina: int(pagina),
		Limite: int(limite),
	})
}

func (h *AporteHandler) obtener(c echo.Context) (*domain.Aporte, int64, error) {
	alumnoID, err := AlumnoID(c)
	if err != nil {
		return nil, 0, err
	}

	id, err := ParamID(c, "id")
	if err != nil {
		return nil, 0, err
	}

	aporte, err := h.repo.GetByID(c.Request().Context(), id, alumnoID)
	if errors.Is(err, domain.ErrAporteNoEncontrado) {
		return nil, 0, NotFound("Aporte")
	}
	if err != nil {
		return nil, 0, err
	}

	return aporte, alumnoID, nil
}

func (h *AporteHandler) Get(c echo.Context) error {
	aporte, _, err := h.obtener(c)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, aporte)
}

// validarReferencias revisa que la materia y el tipo de aporte existan.
func (h *AporteHandler) validarReferencias(c echo.Context, datos domain.DatosAporte) error {
	ctx := c.Request().Context()
	errs := make(map[string]string)

	if _, err := h.materiaRepo.GetByID(ctx, datos.MateriaID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		errs["materia_id"] = "La materia no existe"
	}

	existe, err := h.repo.TagExiste(ctx, datos.TagID)
	if err != nil {
		return err
	}
	if !existe {
		errs["tag_id"] = "El tipo de aporte no existe"
	}

	if len(errs) > 0 {
		return InvalidRequestData(errs)
	}
	return nil
}

func formInt(c echo.Context, name string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(c.FormValue(name)), 10, 64)
	return n
}

// Create recibe un formulario multipart con titulo, descripcion, materia_id,
// tag_id, link (opcional) y archivo (opcional). Hace falta el link o el archivo.
func (h *AporteHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	alumnoID, err := AlumnoID(c)
	if err != nil {
		return err
	}

	datos := domain.DatosAporte{
		Titulo:      c.FormValue("titulo"),
		Descripcion: c.FormValue("descripcion"),
		MateriaID:   formInt(c, "materia_id"),
		TagID:       formInt(c, "tag_id"),
	}
	if link := c.FormValue("link"); link != "" {
		datos.Link = &link
	}
	datos.Normalizar()

	fileHeader, err := c.FormFile("archivo")
	if err != nil && !errors.Is(err, http.ErrMissingFile) {
		return InvalidRequestData(map[string]string{"archivo": "No se pudo leer el archivo"})
	}
	tieneArchivo := fileHeader != nil

	if tieneArchivo && !h.subidaArchivos {
		return InvalidRequestData(map[string]string{"archivo": "La subida de archivos está desactivada. Compartí un link (Drive, YouTube, etc.)"})
	}

	if errs := datos.Validate(tieneArchivo); len(errs) > 0 {
		return InvalidRequestData(errs)
	}

	var archivo *domain.ArchivoAporte
	if tieneArchivo {
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		tipo, ok := tiposPermitidos[ext]
		if !ok {
			return InvalidRequestData(map[string]string{"archivo": "Tipo de archivo no permitido. Se aceptan PDF, imágenes, Word, Excel, PowerPoint, texto y ZIP"})
		}
		if fileHeader.Size > h.maxBytes {
			return InvalidRequestData(map[string]string{"archivo": fmt.Sprintf("El archivo no puede pesar más de %d MB", h.maxBytes/1024/1024)})
		}
		if fileHeader.Size == 0 {
			return InvalidRequestData(map[string]string{"archivo": "El archivo está vacío"})
		}

		archivo = &domain.ArchivoAporte{
			Nombre: nombreArchivoSeguro(fileHeader.Filename),
			Tipo:   tipo,
			Tamano: fileHeader.Size,
		}
	}

	if err := h.validarReferencias(c, datos); err != nil {
		return err
	}

	if archivo != nil {
		src, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		key, err := h.storage.Save(ctx, io.LimitReader(src, h.maxBytes), strings.ToLower(filepath.Ext(fileHeader.Filename)))
		if err != nil {
			h.logger.Error("No se pudo guardar el archivo del aporte", zap.Error(err))
			return err
		}
		archivo.Key = key
	}

	id, err := h.repo.Create(ctx, alumnoID, datos, archivo)
	if err != nil {
		if archivo != nil {
			_ = h.storage.Delete(ctx, archivo.Key)
		}
		return err
	}

	aporte, err := h.repo.GetByID(ctx, id, alumnoID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, aporte)
}

// Update modifica los datos de un aporte propio. El archivo no se puede cambiar:
// para eso se borra el aporte y se sube de nuevo.
func (h *AporteHandler) Update(c echo.Context) error {
	ctx := c.Request().Context()

	aporte, alumnoID, err := h.obtener(c)
	if err != nil {
		return err
	}
	if !aporte.EsMio {
		return Forbidden()
	}

	var datos domain.DatosAporte
	if err := c.Bind(&datos); err != nil {
		return InvalidJSON()
	}
	datos.Normalizar()

	if errs := datos.Validate(aporte.Archivo != nil); len(errs) > 0 {
		return InvalidRequestData(errs)
	}

	if err := h.validarReferencias(c, datos); err != nil {
		return err
	}

	if err := h.repo.Update(ctx, aporte.ID, datos); err != nil {
		return err
	}

	actualizado, err := h.repo.GetByID(ctx, aporte.ID, alumnoID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, actualizado)
}

func (h *AporteHandler) Delete(c echo.Context) error {
	ctx := c.Request().Context()

	aporte, _, err := h.obtener(c)
	if err != nil {
		return err
	}
	if !aporte.EsMio {
		return Forbidden()
	}

	if err := h.repo.Delete(ctx, aporte.ID); err != nil {
		return err
	}

	if aporte.Archivo != nil {
		if err := h.storage.Delete(ctx, aporte.Archivo.Key); err != nil {
			h.logger.Warn("No se pudo borrar el archivo del aporte", zap.Error(err), zap.Int64("aporte_id", aporte.ID))
		}
	}

	return c.NoContent(http.StatusNoContent)
}

// Descargar devuelve el archivo del aporte.
func (h *AporteHandler) Descargar(c echo.Context) error {
	ctx := c.Request().Context()

	aporte, _, err := h.obtener(c)
	if err != nil {
		return err
	}
	if aporte.Archivo == nil {
		return NotFound("Archivo")
	}

	f, err := h.storage.Open(ctx, aporte.Archivo.Key)
	if errors.Is(err, storage.ErrNotFound) {
		return NotFound("Archivo")
	}
	if err != nil {
		return err
	}
	defer f.Close()

	c.Response().Header().Set(echo.HeaderContentDisposition,
		mime.FormatMediaType("attachment", map[string]string{"filename": aporte.Archivo.Nombre}))
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(aporte.Archivo.Tamano, 10))
	c.Response().Header().Set("X-Content-Type-Options", "nosniff")

	return c.Stream(http.StatusOK, aporte.Archivo.Tipo, f)
}

type FavoritoResponse struct {
	Favoritos  int  `json:"favoritos"`
	EsFavorito bool `json:"es_favorito"`
}

func (h *AporteHandler) setFavorito(c echo.Context, favorito bool) error {
	ctx := c.Request().Context()

	aporte, alumnoID, err := h.obtener(c)
	if err != nil {
		return err
	}

	if err := h.repo.SetFavorito(ctx, aporte.ID, alumnoID, favorito); err != nil {
		return err
	}

	actualizado, err := h.repo.GetByID(ctx, aporte.ID, alumnoID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, FavoritoResponse{
		Favoritos:  actualizado.Favoritos,
		EsFavorito: actualizado.EsFavorito,
	})
}

func (h *AporteHandler) AgregarFavorito(c echo.Context) error {
	return h.setFavorito(c, true)
}

func (h *AporteHandler) QuitarFavorito(c echo.Context) error {
	return h.setFavorito(c, false)
}

func (h *AporteHandler) Tags(c echo.Context) error {
	tags, err := h.repo.Tags(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, tags)
}

// nombreArchivoSeguro deja solo el nombre (sin carpetas) y recorta nombres muy largos.
func nombreArchivoSeguro(nombre string) string {
	nombre = filepath.Base(strings.ReplaceAll(nombre, "\\", "/"))
	nombre = strings.Map(func(r rune) rune {
		if r < 32 || r == '"' {
			return -1
		}
		return r
	}, nombre)

	if len([]rune(nombre)) > 150 {
		ext := filepath.Ext(nombre)
		runes := []rune(strings.TrimSuffix(nombre, ext))
		nombre = string(runes[:150-len([]rune(ext))]) + ext
	}

	if nombre == "" || nombre == "." {
		return "archivo"
	}
	return nombre
}

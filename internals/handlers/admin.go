package handlers

import (
	"chedul-core/internals/domain"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// AdminHandler sirve el panel de administracion: solo lectura, solo admins.
type AdminHandler struct {
	repo       domain.AdminRepository
	alumnoRepo domain.AlumnoRepository
}

func NewAdminHandler(repo domain.AdminRepository, alumnoRepo domain.AlumnoRepository) *AdminHandler {
	return &AdminHandler{repo: repo, alumnoRepo: alumnoRepo}
}

// SoloAdmin deja pasar solo a los alumnos marcados como admin. A los demas les
// responde 404, como si la ruta no existiera.
func (h *AdminHandler) SoloAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := AlumnoID(c)
		if err != nil {
			return err
		}
		alumno, err := h.alumnoRepo.GetByID(c.Request().Context(), id)
		if err != nil || !alumno.EsAdmin {
			return echo.ErrNotFound
		}
		return next(c)
	}
}

func (h *AdminHandler) Resumen(c echo.Context) error {
	res, err := h.repo.Resumen(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) Alumnos(c echo.Context) error {
	pagina, _ := strconv.Atoi(c.QueryParam("pagina"))
	alumnos, total, err := h.repo.Alumnos(c.Request().Context(), c.QueryParam("buscar"), pagina)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"alumnos":    alumnos,
		"total":      total,
		"por_pagina": domain.AlumnosAdminPorPagina,
	})
}

func (h *AdminHandler) Accesos(c echo.Context) error {
	accesos, err := h.repo.Accesos(c.Request().Context(), 100)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, accesos)
}

// Estadisticas es publica: la landing muestra cuantos alumnos usan Chedul.
func (h *AdminHandler) Estadisticas(c echo.Context) error {
	total, err := h.alumnoRepo.Contar(c.Request().Context())
	if err != nil {
		return err
	}
	c.Response().Header().Set("Cache-Control", "public, max-age=300")
	return c.JSON(http.StatusOK, map[string]int{"alumnos": total})
}

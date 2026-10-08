package domain

import "context"

// ResumenAdmin son los numeros generales del panel de administracion.
type ResumenAdmin struct {
	Alumnos     int `json:"alumnos"`
	Nuevos7     int `json:"nuevos_7_dias"`
	Nuevos30    int `json:"nuevos_30_dias"`
	Activos24h  int `json:"activos_24_horas"`
	Activos7    int `json:"activos_7_dias"`
	ConGoogle   int `json:"con_google"`
	ConEstado   int `json:"con_estado_cargado"`
	Aportes     int `json:"aportes"`
	Comunidades int `json:"comunidades"`
	SesionesHoy int `json:"accesos_hoy"`
	// Ultimos 30 dias, de mas viejo a mas nuevo
	RegistrosPorDia []CantidadPorDia `json:"registros_por_dia"`
	AccesosPorDia   []CantidadPorDia `json:"accesos_por_dia"`
}

type CantidadPorDia struct {
	Dia      string `json:"dia"`
	Cantidad int    `json:"cantidad"`
}

// AlumnoAdmin es una fila de la lista de alumnos del panel.
type AlumnoAdmin struct {
	ID           int64   `json:"id"`
	Nombre       string  `json:"nombre"`
	Email        string  `json:"email"`
	Creado       *string `json:"creado"`
	UltimoAcceso *string `json:"ultimo_acceso"`
	Google       bool    `json:"google"`
	EsAdmin      bool    `json:"es_admin"`
	Materias     int     `json:"materias"`
	Aportes      int     `json:"aportes"`
}

// AccesoAdmin es un inicio de sesion.
type AccesoAdmin struct {
	Fecha    string `json:"fecha"`
	Metodo   string `json:"metodo"`
	AlumnoID int64  `json:"alumno_id"`
	Nombre   string `json:"nombre"`
	Email    string `json:"email"`
}

const AlumnosAdminPorPagina = 50

type AdminRepository interface {
	Resumen(ctx context.Context) (*ResumenAdmin, error)
	// Alumnos lista de a AlumnosAdminPorPagina, los mas nuevos primero.
	// buscar filtra por nombre o correo.
	Alumnos(ctx context.Context, buscar string, pagina int) ([]AlumnoAdmin, int, error)
	Accesos(ctx context.Context, limite int) ([]AccesoAdmin, error)
}

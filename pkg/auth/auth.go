package auth

import "github.com/labstack/echo/v4"

// ContextKey define una clave para almacenar datos en el contexto de Echo.
// Es una buena práctica usar un tipo para evitar colisiones con otras claves.
type ContextKey string

const AlumnoIDContextKey ContextKey = "alumnoID" // Clave para almacenar el ID del alumno en el contexto de la petición

// GetAlumnoIDFromContext es una función de utilidad para obtener el AlumnoID de la petición.
func GetAlumnoIDFromContext(c echo.Context) (int64, bool) {
	id, ok := c.Get(string(AlumnoIDContextKey)).(int64)
	return id, ok
}

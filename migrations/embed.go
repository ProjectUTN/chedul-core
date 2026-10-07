// Package migrations embebe los archivos SQL de goose en el binario, asi el
// servidor puede aplicar las migraciones al iniciar sin necesitar el CLI de goose
// (la imagen de Docker es "scratch" y no trae nada mas que el binario).
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

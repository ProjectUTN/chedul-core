// Package storage guarda los archivos que suben los alumnos.
//
// Por ahora hay una sola implementacion, en disco local. En produccion el
// directorio tiene que estar en un volumen persistente (por ejemplo un volume
// de Fly.io o Railway). Si mas adelante se quiere usar S3 o Cloudflare R2
// alcanza con otra implementacion de Storage.
package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

var ErrNotFound = errors.New("archivo no encontrado")

type Storage interface {
	// Save guarda el contenido y devuelve la clave con la que se puede recuperar.
	Save(ctx context.Context, r io.Reader, ext string) (string, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

// Las claves son generadas por nosotros: 32 caracteres hex + extension.
// Validarlas evita que una clave manipulada lea fuera del directorio.
var keyPattern = regexp.MustCompile(`^[a-f0-9]{32}(\.[a-z0-9]{1,8})?$`)

type Local struct {
	dir string
}

func NewLocal(dir string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("no se pudo crear el directorio de archivos %q: %w", dir, err)
	}
	return &Local{dir: dir}, nil
}

func (l *Local) path(key string) (string, error) {
	if !keyPattern.MatchString(key) {
		return "", ErrNotFound
	}
	return filepath.Join(l.dir, key), nil
}

func (l *Local) Save(_ context.Context, r io.Reader, ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	key := hex.EncodeToString(buf) + ext
	path, err := l.path(key)
	if err != nil {
		return "", fmt.Errorf("extension inválida %q", ext)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(path)
		return "", err
	}

	if err := f.Close(); err != nil {
		os.Remove(path)
		return "", err
	}

	return key, nil
}

func (l *Local) Open(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := l.path(key)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	return f, err
}

func (l *Local) Delete(_ context.Context, key string) error {
	path, err := l.path(key)
	if err != nil {
		return nil
	}

	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

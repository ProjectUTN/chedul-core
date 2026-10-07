package storage

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLocalGuardaLeeYBorra(t *testing.T) {
	ctx := context.Background()
	s, err := NewLocal(t.TempDir())
	require.NoError(t, err)

	key, err := s.Save(ctx, strings.NewReader("resumen de AM1"), ".pdf")
	require.NoError(t, err)
	require.True(t, strings.HasSuffix(key, ".pdf"))

	r, err := s.Open(ctx, key)
	require.NoError(t, err)
	contenido, err := io.ReadAll(r)
	r.Close()
	require.NoError(t, err)
	require.Equal(t, "resumen de AM1", string(contenido))

	require.NoError(t, s.Delete(ctx, key))
	_, err = s.Open(ctx, key)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestLocalRechazaClavesManipuladas(t *testing.T) {
	s, err := NewLocal(t.TempDir())
	require.NoError(t, err)

	for _, key := range []string{"../config.yml", "/etc/passwd", "abc", ""} {
		_, err := s.Open(context.Background(), key)
		require.ErrorIs(t, err, ErrNotFound, key)
	}
}

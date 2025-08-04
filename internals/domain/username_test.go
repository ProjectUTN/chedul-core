package domain

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestUsernamePropTest(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Nombres de usuario validos son aceptados", prop.ForAll(
		func(validUsername string) bool {
			email, err := NewUsername(validUsername)
			if err != nil {
				return false
			}
			return email.String() != ""
		},
		genValidUserNames(),
	))

	properties.TestingRun(t)
}

func genValidUserNames() gopter.Gen {
	return gen.Const("").Map(func(_ string) string {
		return gofakeit.Username()
	})
}

func TestUsername(t *testing.T) {
	testCases := []struct {
		name     string
		username string
		rejected bool
	}{
		{"nombre de 256 caracteres es valido", strings.Repeat("a", 256), false},
		{"nombre mas largo que 256 caracteres es rechazado", strings.Repeat("a", 257), true},
		{"nombres que solo contienen espacios son rechazados", " ", true},
		{"string vacio es rechazada", "", true},
		{"nombre valido es procesado exitosamente", "Lautaro Acosta Quintana", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewUsername(tc.username)
			if tc.rejected && err == nil {
				t.Errorf("Se esperaba error para el nombre de usuario %q, pero no se obtuvo ninguno", tc.username)
			}
			if !tc.rejected && err != nil {
				t.Errorf("No se esperaba error para el nombre de usuario %q, pero se obtuvo: %v", tc.username, err)
			}
		})
	}
	invalidChars := []rune{'/', '(', ')', '"', '<', '>', '\\', '{', '}'}
	for _, char := range invalidChars {
		t.Run("caracter invalido "+string(char)+" es rechazado", func(t *testing.T) {
			_, err := NewUsername(string(char))
			if err == nil {
				t.Errorf("Se esperaba error para el nombre de usuario que contiene %q, pero no se obtuvo ninguno", char)
			}
		})
	}
}

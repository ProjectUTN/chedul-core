package domain

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestPasswordPropTest(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 1000
	properties := gopter.NewProperties(parameters)

	properties.Property("claves válidas son aceptadas", prop.ForAll(
		func(validPassword string) bool {
			password, err := ParsePassword(validPassword)
			if err != nil {
				t.Logf("Clave rechazada: %q\nError: %v", validPassword, err)
				return false
			}
			return password.String() != ""
		},
		genValidPasswords(),
	))

	properties.TestingRun(t)
}

func genValidPasswords() gopter.Gen {
	return gen.Const("").Map(func(_ string) string {
		length := gofakeit.Number(8, 128)

		var password []rune
		password = append(password, []rune(gofakeit.LetterN(1))...)
		password = append(password, []rune(strings.ToUpper(gofakeit.LetterN(1)))...)
		password = append(password, []rune(gofakeit.DigitN(1))...)
		password = append(password, rune("!@#$%&*_-+="[gofakeit.Number(0, 10)]))

		remaining := length - len(password)

		all := "abcdefghijklmnopqrstuvwxyz" +
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"0123456789" +
			"!@#$%&*_-+="

		for range remaining {
			password = append(password, rune(all[gofakeit.Number(0, len(all)-1)]))
		}

		gofakeit.ShuffleAnySlice(password)
		return string(password)
	})
}

func TestInvalidPasswords(t *testing.T) {
	testCases := []struct {
		name     string
		password string
	}{
		{"password con espacios", "My Secure P@ssw0rd 123"},
		{"clave de +128 caracteres", strings.Repeat("a", 256)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParsePassword(tc.password)
			if err == nil {
				t.Errorf("Password inválida %q fue aceptada: %v", tc.password, err)
			}
		})
	}
}

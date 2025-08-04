package domain

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestEmailValidationProperty(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("emails validos son aceptados", prop.ForAll(
		func(validEmail string) bool {
			email, err := ParseEmail(validEmail)
			if err != nil {
				return false
			}
			return email.String() != ""
		},
		generateValidEmails(),
	))

	properties.TestingRun(t)
}

func generateValidEmails() gopter.Gen {
	return gen.Const("").Map(func(_ string) string {
		return gofakeit.Email()
	})
}

func TestEmailValidationSpecificCases(t *testing.T) {
	testCases := []struct {
		name     string
		email    string
		rejected bool
	}{
		{"string vacio", "", true},
		{"falta el simbolo @ es rechazado", "example.com", true},
		{"falta el asunto es rechazado", "@domain", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseEmail(tc.email)
			if tc.rejected && err == nil {
				t.Errorf("Se esperaba error para el email %q, pero no se obtuvo ninguno", tc.email)
			}
			if !tc.rejected && err != nil {
				t.Errorf("No se esperaba error para el email %q, pero se obtuvo: %v", tc.email, err)
			}
		})
	}
}

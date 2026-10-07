package db

import "testing"

func TestLimpiarDSN(t *testing.T) {
	casos := map[string]string{
		// URL tal cual la muestra Neon
		"postgresql://u:p@ep-x.sa-east-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require": "postgresql://u:p@ep-x.sa-east-1.aws.neon.tech/neondb?sslmode=require",
		"postgres://u:p@localhost:5432/db?sslmode=disable":                                             "postgres://u:p@localhost:5432/db?sslmode=disable",
		"postgres://u:p@localhost:5432/db":                                                             "postgres://u:p@localhost:5432/db",
		"postgres://u:p@localhost/db?gssencmode=disable":                                               "postgres://u:p@localhost/db",
	}

	for entrada, esperado := range casos {
		if got := limpiarDSN(entrada); got != esperado {
			t.Errorf("limpiarDSN(%q) = %q, se esperaba %q", entrada, got, esperado)
		}
	}
}

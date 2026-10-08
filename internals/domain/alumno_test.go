package domain

import "testing"

func TestNombreCorto(t *testing.T) {
	casos := []struct{ nombre, apellido, esperado string }{
		{"Juan", "Pérez", "Juan P."},
		{"Juan Pablo", "pérez gómez", "Juan P."},
		{"Ana María López", "", "Ana L."},
		{"beto", "", "beto"},
		{"Caro", "  ", "Caro"},
		{"", "Ruiz", "Anónimo"},
		{"Lu", "123", "Lu"},
	}
	for _, c := range casos {
		if got := NombreCorto(c.nombre, c.apellido); got != c.esperado {
			t.Errorf("NombreCorto(%q, %q) = %q, quiero %q", c.nombre, c.apellido, got, c.esperado)
		}
	}
}

func TestParseApellido(t *testing.T) {
	if got, err := ParseApellido("  Pérez "); err != nil || got != "Pérez" {
		t.Errorf("recorta espacios: %q %v", got, err)
	}
	if _, err := ParseApellido("<script>"); err == nil {
		t.Error("tiene que rechazar caracteres raros")
	}
	if got, err := ParseApellido(""); err != nil || got != "" {
		t.Error("vacio es valido")
	}
}

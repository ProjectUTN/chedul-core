package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlataformaDeLink(t *testing.T) {
	casos := map[string]string{
		"https://chat.whatsapp.com/AbCdEf123": "whatsapp",
		"https://discord.gg/abc":              "discord",
		"https://www.instagram.com/isi.frre/": "instagram",
		"https://t.me/isi_utn":                "telegram",
		"https://www.reddit.com/r/utn":        "otra",
		"no es un link":                       "otra",
	}
	for link, esperada := range casos {
		assert.Equal(t, esperada, PlataformaDeLink(link), link)
	}
}

func TestDatosComunidadValidate(t *testing.T) {
	d := DatosComunidad{Nombre: "  ISI FRRe  ", Link: " https://chat.whatsapp.com/x "}
	d.Normalizar()
	assert.Empty(t, d.Validate())
	assert.Equal(t, "ISI FRRe", d.Nombre)
	assert.Equal(t, "whatsapp", d.Plataforma)

	for _, link := range []string{"", "http://chat.whatsapp.com/x", "javascript:alert(1)", "chat.whatsapp.com/x"} {
		d := DatosComunidad{Nombre: "Grupo", Link: link}
		d.Normalizar()
		assert.Contains(t, d.Validate(), "link", link)
	}
}

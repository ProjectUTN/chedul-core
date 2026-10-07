package domain

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"strings"
)

type Email struct {
	value string
}

func ParseEmail(s string) (Email, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(s))

	if err != nil {
		return Email{}, fmt.Errorf("formato invalido: %v", err)
	}

	// Se guarda en minusculas para que el login no dependa de como se escribio
	return Email{value: strings.ToLower(addr.Address)}, nil
}

func (e *Email) String() string {
	return e.value
}

func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

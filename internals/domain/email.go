package domain

import (
	"encoding/json"
	"fmt"
	"net/mail"
)

type Email struct {
	value string
}

func ParseEmail(s string) (Email, error) {
	addr, err := mail.ParseAddress(s)

	if err != nil {
		return Email{}, fmt.Errorf("formato invalido: %v", err)
	}

	return Email{value: addr.Address}, nil
}

func (e *Email) String() string {
	return e.value
}

func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

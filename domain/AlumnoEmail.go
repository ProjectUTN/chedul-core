package domain

import (
	"fmt"
	"net/mail"
)

type AlumnoEmail string

func NewAlumnoEmail(s string) (AlumnoEmail, error) {
	addr, err := mail.ParseAddress(s)

	if err != nil {
		return "", fmt.Errorf("formato invalido: %v", err)
	}

	return AlumnoEmail(addr.Address), nil
}

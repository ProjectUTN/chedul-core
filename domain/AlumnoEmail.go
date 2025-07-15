package domain

import (
	"encoding/json"
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

func (self AlumnoEmail) String() string {
	return string(self)
}

func (self AlumnoEmail) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(self))
}

func (self *AlumnoEmail) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	addr, err := NewAlumnoEmail(s)
	if err != nil {
		return err
	}

	*self = addr

	return nil
}

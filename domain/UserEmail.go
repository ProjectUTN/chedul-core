package domain

import (
	"encoding/json"
	"fmt"
	"net/mail"
)

type UserEmail string

func NewUserEmail(s string) (UserEmail, error) {
	addr, err := mail.ParseAddress(s)

	if err != nil {
		return "", fmt.Errorf("formato invalido: %v", err)
	}

	return UserEmail(addr.Address), nil
}

func (self UserEmail) String() string {
	return string(self)
}

func (self UserEmail) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(self))
}

func (self *UserEmail) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	addr, err := NewUserEmail(s)
	if err != nil {
		return err
	}

	*self = addr

	return nil
}

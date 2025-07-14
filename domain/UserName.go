package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
)

type UserName string

func NewUserName(s string) (UserName, error) {
	emptyOrWhitespace := strings.TrimSpace(s) == ""
	isTooLong := uniseg.GraphemeClusterCount(s) > 256
	forbiddenChars := []rune{'/', '(', ')', '"', '<', '>', '\\', '{', '}'}
	hasInvalidChar := false
	for _, char := range s {
		for _, forbidden := range forbiddenChars {
			if char == forbidden {
				hasInvalidChar = true
			}
		}
	}

	if hasInvalidChar || emptyOrWhitespace || isTooLong {
		return "", fmt.Errorf("%s es un nombre invalido", s)
	}

	return UserName(s), nil
}

func (self UserName) String() string {
	return string(self)
}

func (self UserName) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(self))
}

func (self *UserName) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	name, err := NewUserName(s)
	if err != nil {
		return err
	}

	*self = name

	return nil
}

package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
)

type Username struct {
	value string
}

func (u *Username) String() string {
	return u.value
}

func (u Username) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func NewUsername(s string) (Username, error) {
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
		return Username{}, fmt.Errorf("''%s' es un nombre invalido", s)
	}

	return Username{value: s}, nil
}

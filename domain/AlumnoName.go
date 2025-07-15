package domain

import (
	"fmt"
	"strings"

	"github.com/rivo/uniseg"
)

type AlumnoName string

func NewAlumnoName(s string) (AlumnoName, error) {
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
		return "", fmt.Errorf("''%s' es un nombre invalido", s)
	}

	return AlumnoName(s), nil
}

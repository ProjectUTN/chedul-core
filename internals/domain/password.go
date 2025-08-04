package domain

import (
	"fmt"
	"strings"
	"unicode"

	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"

	"golang.org/x/crypto/argon2"
)

const (
	memory      = 19 * 1024
	iterations  = 2
	parallelism = 1
	saltLength  = 16
	keyLength   = 32
)

type Password struct {
	inner string
}

func ParsePassword(s string) (Password, error) {
	if strings.ContainsAny(s, " \t\n\r") {
		return Password{}, fmt.Errorf("La clave no debe tener espacios vacios")
	}
	if len(s) < 8 {
		return Password{}, fmt.Errorf("La clave debe tener minimo 8 caracteres")
	}

	if len(s) > 128 {
		return Password{}, fmt.Errorf("la clave no puede exceder 128 caracteres")
	}

	if !hasCharacterDiversity(s) {
		return Password{}, fmt.Errorf("la clave debe contener al menos 3 de los siguientes tipos: mayúsculas, minúsculas, números y símbolos")
	}

	password, err := hash(s)
	if err != nil {
		return Password{}, fmt.Errorf("La clave no pudo ser encriptada")
	}

	return Password{
		inner: password,
	}, nil
}

// TODO: No me gusta tener algo asi. Usar con cuidado
func ParsePasswordFromEncrypted(s string) Password {
	return Password{
		inner: s,
	}
}

func (p *Password) String() string {
	return p.inner
}

func hasCharacterDiversity(password string) bool {
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	count := 0

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSymbol {
		count++
	}

	return count >= 3
}

func hash(s string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("no se pudo generar la sal: %w", err)
	}

	hash := argon2.IDKey([]byte(s), salt, iterations, memory, parallelism, keyLength)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, iterations, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func CheckPassword(hashed, password string) (bool, error) {
	parts := strings.Split(hashed, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("formato hash inválido: expected 6 parts, got %d", len(parts))
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("formato hash inválido: expected argon2id, got %s", parts[1])
	}

	if parts[2] != "v=19" {
		return false, fmt.Errorf("formato hash inválido: expected v=19, got %s", parts[2])
	}

	var mem, it uint32
	var par uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &it, &par)
	if err != nil {
		return false, fmt.Errorf("formato hash inválido en parámetros: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("salt corrupta: %w", err)
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("hash corrupto: %w", err)
	}

	computed := argon2.IDKey([]byte(password), salt, it, mem, par, uint32(len(expectedHash)))

	if subtle.ConstantTimeCompare(computed, expectedHash) == 1 {
		return true, nil
	}

	return false, nil
}

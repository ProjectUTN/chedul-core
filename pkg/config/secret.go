package config

import (
	"crypto/subtle"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
)

type Secret struct {
	value string
}

func NewSecret(value string) *Secret {
	return &Secret{value: value}
}

func (s *Secret) Expose() string {
	if s == nil {
		return ""
	}
	return s.value
}

func (s *Secret) String() string {
	return "[REDACTED]"
}

func (s *Secret) GoString() string {
	return "[REDACTED]"
}

func (s *Secret) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}

func (s *Secret) ConstantTimeEquals(other *Secret) bool {
	if s == nil || other == nil {
		return s == other
	}
	return subtle.ConstantTimeCompare([]byte(s.value), []byte(other.value)) == 1
}

func SecretDecodeHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if t != reflect.TypeOf(&Secret{}) {
			return data, nil
		}

		if str, ok := data.(string); ok {
			return NewSecret(str), nil
		}

		return data, nil
	}
}

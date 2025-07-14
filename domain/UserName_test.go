package domain_test

import (
	"chedul-core/domain"
	"strings"
	"testing"
)

func TestNewUserName(t *testing.T) {
	tests := []struct {
		description string
		s           string
		want        domain.UserName
		wantErr     bool
	}{
		{
			description: "a_256_graphene_long_name_is_valid",
			s:           strings.Repeat("a", 256),
			want:        domain.UserName(strings.Repeat("a", 256)),
			wantErr:     false,
		},
		{
			description: "a_name_longer_than_256_graphemes_is_rejected",
			s:           strings.Repeat("a", 257),
			wantErr:     true,
		},
		{
			description: "whitespace_only_names_are_rejected",
			s:           "     ",
			wantErr:     true,
		},
		{
			description: "a_valid_name_is_parsed_successfully",
			s:           "Chedul Chedulero",
			want:        domain.UserName("Chedul Chedulero"),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			got, gotErr := domain.NewUserName(tt.s)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("NewUserName() falló inesperadamente: %v", gotErr)
				}
				return
			}

			if tt.wantErr {
				t.Fatal("NewUserName() fue exitoso cuando se esperaba error")
			}

			if got != tt.want {
				t.Errorf("NewUserName() = %v, se esperaba %v", got, tt.want)
			}
		})
	}
}

func TestNewUserName_with_forbidden_characters(t *testing.T) {
	forbidden := []rune{'/', '(', ')', '"', '<', '>', '\\', '{', '}'}
	for _, ch := range forbidden {
		t.Run("forbidden_char_"+string(ch), func(t *testing.T) {
			_, err := domain.NewUserName(string(ch))
			if err == nil {
				t.Errorf("se esperaba error para la entrada %q, pero no se obtuvo", ch)
			}
		})
	}
}

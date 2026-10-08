package handlers

import "testing"

func TestRefreshTokenNoSirveComoAccess(t *testing.T) {
	const secreto = "secreto-de-prueba-secreto-de-prueba"
	refresh, err := GenerateRefreshToken(7, 0, secreto)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseAccessToken(refresh, secreto); err == nil {
		t.Error("el refresh token se acepto como access token")
	}
	if _, err := ParseRefreshToken(refresh, secreto); err != nil {
		t.Errorf("el refresh token no se acepto como refresh: %v", err)
	}

	access, err := GenerateAccessToken(7, secreto)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseRefreshToken(access, secreto); err == nil {
		t.Error("el access token se acepto como refresh token")
	}
	if claims, err := ParseAccessToken(access, secreto); err != nil || claims.Sub != 7 {
		t.Errorf("el access token no se acepto: %v", err)
	}
}

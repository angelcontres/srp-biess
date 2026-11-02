package security_test

import (
	"testing"

	"srp-biess/internal/infrastructure/security"
)

func TestArgon2PasswordHashing(t *testing.T) {
	// Parámetros ligeros para velocidad en tests
	fastParams := &security.Argon2Params{
		Memory:      16 * 1024,
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	password := "ContrasenaSuperSegura2026!"

	encodedHash, err := security.HashPasswordWithParams(password, fastParams)
	if err != nil {
		t.Fatalf("error al generar hash: %v", err)
	}

	if encodedHash == "" || len(encodedHash) < 40 {
		t.Errorf("hash generado sospechosamente corto: %s", encodedHash)
	}

	// Verificación exitosa
	valido, err := security.VerifyPassword(password, encodedHash)
	if err != nil {
		t.Fatalf("error verificando password correcto: %v", err)
	}
	if !valido {
		t.Errorf("se esperaba que la contraseña coincidiera con el hash")
	}

	// Verificación fallida con password incorrecto
	invalido, err := security.VerifyPassword("ContrasenaEquivocada", encodedHash)
	if err != nil {
		t.Fatalf("error inesperado verificando password incorrecto: %v", err)
	}
	if invalido {
		t.Errorf("no debía coincidir con una contraseña errónea")
	}

	// Formato de hash corrupto
	_, err = security.VerifyPassword(password, "hash_invalido_sin_formato")
	if err == nil {
		t.Errorf("se esperaba error al validar hash corrupto")
	}
}

func TestGenerateOpaqueToken(t *testing.T) {
	token1, err := security.GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("error generando token: %v", err)
	}
	if len(token1) != 64 { // 32 bytes en hex = 64 caracteres
		t.Errorf("longitud esperada 64, obtenida: %d", len(token1))
	}

	token2, err := security.GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("error generando token 2: %v", err)
	}
	if token1 == token2 {
		t.Errorf("dos tokens generados sucesivamente no deben ser iguales")
	}
}

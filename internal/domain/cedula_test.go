package domain_test

import (
	"testing"

	"srp-biess/internal/domain"
)

func TestValidarCedulaEcuador(t *testing.T) {
	tests := []struct {
		name     string
		cedula   string
		expected bool
	}{
		{"Cédula válida Guayas", "0926687856", true},
		{"Cédula válida Pichincha", "1710034065", true},
		{"Longitud incorrecta corta", "092668785", false},
		{"Longitud incorrecta larga", "09266878561", false},
		{"Caracteres no numéricos", "092668785A", false},
		{"Provincia inválida (00)", "0026687856", false},
		{"Provincia inválida (25)", "2526687856", false},
		{"Tercer dígito >= 6", "0966687856", false},
		{"Dígito verificador erróneo", "0926687850", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.ValidarCedulaEcuador(tt.cedula)
			if got != tt.expected {
				t.Errorf("ValidarCedulaEcuador(%s) = %v; want %v", tt.cedula, got, tt.expected)
			}
		})
	}
}

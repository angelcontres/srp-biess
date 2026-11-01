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

// Pruebas del Patrón Strategy
func TestDocumentValidationStrategies(t *testing.T) {
	t.Run("Estrategia Cédula Ecuatoriana", func(t *testing.T) {
		strategy := domain.NewCedulaEcuatorianaStrategy()
		if strategy.Tipo() != "CEDULA_ECUADOR" {
			t.Errorf("tipo esperado CEDULA_ECUADOR, obtenido: %s", strategy.Tipo())
		}

		valido, err := strategy.Validate("0926687856")
		if !valido || err != nil {
			t.Errorf("esperaba cédula válida, error: %v", err)
		}

		valido, err = strategy.Validate("0926687850")
		if valido || err == nil {
			t.Errorf("esperaba rechazo de cédula inválida")
		}
	})

	t.Run("Estrategia RUC Ecuatoriano", func(t *testing.T) {
		strategy := domain.NewRucEcuatorianoStrategy()
		if strategy.Tipo() != "RUC_ECUADOR" {
			t.Errorf("tipo esperado RUC_ECUADOR, obtenido: %s", strategy.Tipo())
		}

		// RUC válido de persona natural: cédula válida + 001
		valido, err := strategy.Validate("0926687856001")
		if !valido || err != nil {
			t.Errorf("esperaba RUC válido, error: %v", err)
		}

		// RUC con sucursal inválida
		valido, err = strategy.Validate("0926687856000")
		if valido || err == nil {
			t.Errorf("esperaba rechazo de RUC sin 001")
		}
	})

	t.Run("Estrategia Pasaporte", func(t *testing.T) {
		strategy := domain.NewPasaporteStrategy()
		if strategy.Tipo() != "PASAPORTE" {
			t.Errorf("tipo esperado PASAPORTE, obtenido: %s", strategy.Tipo())
		}

		valido, err := strategy.Validate("AB1234567")
		if !valido || err != nil {
			t.Errorf("esperaba pasaporte válido")
		}

		valido, err = strategy.Validate("123")
		if valido || err == nil {
			t.Errorf("esperaba rechazo por pasaporte muy corto")
		}
	})
}

package application_test

import (
	"testing"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

func TestAuthService_ValidarIdentidad_StrategyPattern(t *testing.T) {
	t.Run("Validación con estrategia por defecto (Cédula Ecuador)", func(t *testing.T) {
		svc := application.NewAuthService(nil, nil)

		// Cédula válida Módulo 10
		if err := svc.ValidarIdentidad("0926687856"); err != nil {
			t.Errorf("esperaba éxito con cédula válida, error: %v", err)
		}

		// Cédula inválida
		if err := svc.ValidarIdentidad("0926687850"); err != domain.ErrInvalidCedula {
			t.Errorf("esperaba ErrInvalidCedula, obtenido: %v", err)
		}
	})

	t.Run("Intercambio dinámico a estrategia Pasaporte", func(t *testing.T) {
		svc := application.NewAuthService(nil, domain.NewCedulaEcuatorianaStrategy())

		// Como tiene estrategia cédula, un pasaporte alfanumérico falla
		if err := svc.ValidarIdentidad("PASAPORTE123"); err == nil {
			t.Errorf("esperaba rechazo de pasaporte con validador de cédula")
		}

		// Cambiamos la estrategia dinámicamente a Pasaporte (Patrón Strategy en acción)
		svc.SetValidator(domain.NewPasaporteStrategy())

		if err := svc.ValidarIdentidad("PASAPORTE123"); err != nil {
			t.Errorf("esperaba éxito al usar estrategia de pasaporte, error: %v", err)
		}
	})
}

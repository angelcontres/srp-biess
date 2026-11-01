package domain

import (
	"strconv"
)

// ValidarCedulaEcuador implementa el algoritmo matemático Módulo 10 según normativa ecuatoriana (ADR-013).
// Reglas:
// 1. Longitud exacta de 10 dígitos.
// 2. Dos primeros dígitos (código provincial): entre 01 y 24, o 30 (consulados).
// 3. Tercer dígito menor a 6 (personas naturales).
// 4. Coeficientes alternos [2, 1, 2, 1, 2, 1, 2, 1, 2]. Si el producto >= 10, se le resta 9.
// 5. El dígito verificador es (decenaSuperior - suma) % 10.
func ValidarCedulaEcuador(cedula string) bool {
	if len(cedula) != 10 {
		return false
	}

	// Verificar que sean todos dígitos numéricos
	for _, c := range cedula {
		if c < '0' || c > '9' {
			return false
		}
	}

	// Validar código de provincia
	provincia, err := strconv.Atoi(cedula[0:2])
	if err != nil || (provincia < 1 || (provincia > 24 && provincia != 30)) {
		return false
	}

	// Validar tercer dígito (< 6 para personas naturales)
	tercerDigito, err := strconv.Atoi(string(cedula[2]))
	if err != nil || tercerDigito >= 6 {
		return false
	}

	// Algoritmo Módulo 10
	coeficientes := [9]int{2, 1, 2, 1, 2, 1, 2, 1, 2}
	suma := 0

	for i := 0; i < 9; i++ {
		digito := int(cedula[i] - '0')
		producto := digito * coeficientes[i]
		if producto >= 10 {
			producto -= 9
		}
		suma += producto
	}

	digitoVerificadorEsperado := (10 - (suma % 10)) % 10
	digitoVerificadorReal := int(cedula[9] - '0')

	return digitoVerificadorEsperado == digitoVerificadorReal
}

// ---------------------------------------------------------------------
// Patrón Strategy: Abstracción de Validadores de Identidad (ADR-013)
// ---------------------------------------------------------------------

// DocumentValidationStrategy define el contrato abstracto del Patrón Strategy.
// Permite intercambiar algoritmos de validación (Cédula, RUC, Pasaporte) sin modificar el cliente.
type DocumentValidationStrategy interface {
	Validate(document string) (bool, error)
	Tipo() string
}

// CedulaEcuatorianaStrategy implementa la estrategia de validación de cédulas (Módulo 10).
type CedulaEcuatorianaStrategy struct{}

func NewCedulaEcuatorianaStrategy() *CedulaEcuatorianaStrategy {
	return &CedulaEcuatorianaStrategy{}
}

func (s *CedulaEcuatorianaStrategy) Tipo() string {
	return "CEDULA_ECUADOR"
}

func (s *CedulaEcuatorianaStrategy) Validate(document string) (bool, error) {
	if !ValidarCedulaEcuador(document) {
		return false, ErrInvalidCedula
	}
	return true, nil
}

// RucEcuatorianoStrategy implementa la estrategia para RUC de personas naturales (13 dígitos).
type RucEcuatorianoStrategy struct{}

func NewRucEcuatorianoStrategy() *RucEcuatorianoStrategy {
	return &RucEcuatorianoStrategy{}
}

func (s *RucEcuatorianoStrategy) Tipo() string {
	return "RUC_ECUADOR"
}

func (s *RucEcuatorianoStrategy) Validate(document string) (bool, error) {
	if len(document) != 13 {
		return false, ErrInvalidCedula
	}
	if document[10:] != "001" {
		return false, ErrInvalidCedula
	}
	if !ValidarCedulaEcuador(document[:10]) {
		return false, ErrInvalidCedula
	}
	return true, nil
}

// PasaporteStrategy implementa la validación de formato alfanumérico para pasaportes extranjeros.
type PasaporteStrategy struct{}

func NewPasaporteStrategy() *PasaporteStrategy {
	return &PasaporteStrategy{}
}

func (s *PasaporteStrategy) Tipo() string {
	return "PASAPORTE"
}

func (s *PasaporteStrategy) Validate(document string) (bool, error) {
	if len(document) < 6 || len(document) > 20 {
		return false, ErrInvalidCedula
	}
	return true, nil
}

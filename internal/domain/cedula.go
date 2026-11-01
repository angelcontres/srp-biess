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

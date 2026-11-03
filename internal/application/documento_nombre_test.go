package application_test

import (
	"strings"
	"testing"

	"srp-biess/internal/application"
)

func TestSanearNombreArchivo(t *testing.T) {
	casos := map[string]string{
		"Título Ingeniería.pdf":     "Título Ingeniería.pdf",
		"../../etc/passwd.pdf":      "passwd.pdf",
		`C:\Users\maria\titulo.pdf`: "titulo.pdf",
		"tit\x00ulo\n.pdf":          "titulo.pdf",
		"   ":                       "documento.pdf",
		"..":                        "..",
	}
	for entrada, esperado := range casos {
		if got := application.SanearNombreArchivo(entrada); got != esperado {
			t.Errorf("SanearNombreArchivo(%q) = %q, esperado %q", entrada, got, esperado)
		}
	}
	largo := application.SanearNombreArchivo(strings.Repeat("a", 300) + ".pdf")
	if len([]rune(largo)) != 255 || !strings.HasSuffix(largo, ".pdf") {
		t.Fatalf("debía recortarse a 255 caracteres conservando la extensión: %d", len([]rune(largo)))
	}
}

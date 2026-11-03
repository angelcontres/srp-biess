package security_test

import (
	"errors"
	"testing"
	"time"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/security"
)

func TestFirmadorHMAC(t *testing.T) {
	f := security.NewFirmadorHMAC([]byte("secreto-de-prueba"))
	ahora := time.Unix(1_800_000_000, 0)
	vence := ahora.Add(5 * time.Minute)
	firma := f.Firmar("doc-1", vence)

	if err := f.Verificar("doc-1", vence.Unix(), firma, ahora); err != nil {
		t.Fatalf("una firma vigente debía validar: %v", err)
	}
	casos := map[string]func() error{
		"vencida":            func() error { return f.Verificar("doc-1", vence.Unix(), firma, vence.Add(time.Second)) },
		"otro documento":     func() error { return f.Verificar("doc-2", vence.Unix(), firma, ahora) },
		"vencimiento movido": func() error { return f.Verificar("doc-1", vence.Unix()+3600, firma, ahora) },
		"firma alterada":     func() error { return f.Verificar("doc-1", vence.Unix(), "00"+firma[2:], ahora) },
		"no hexadecimal":     func() error { return f.Verificar("doc-1", vence.Unix(), "zz", ahora) },
		"otro secreto": func() error {
			return security.NewFirmadorHMAC([]byte("otro")).Verificar("doc-1", vence.Unix(), firma, ahora)
		},
	}
	for nombre, verificar := range casos {
		if err := verificar(); !errors.Is(err, domain.ErrFirmaDescargaInvalida) {
			t.Errorf("%s: esperado ErrFirmaDescargaInvalida, obtenido %v", nombre, err)
		}
	}
}

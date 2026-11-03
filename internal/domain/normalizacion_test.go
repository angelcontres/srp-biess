package domain_test

import (
	"testing"
	"time"

	"srp-biess/internal/domain"
)

func fechaISO(t *testing.T, s string) time.Time {
	t.Helper()
	f, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("fecha %q: %v", s, err)
	}
	return f
}

// Mismos casos que frontend/src/features/postulacion/utils/experiencia.test.ts (SRS 7.5).
func TestCalcularTiempoExperiencia(t *testing.T) {
	casos := []struct {
		inicio, fin string
		esperado    domain.TiempoExperiencia
	}{
		{"2020-01-15", "2023-03-10", domain.TiempoExperiencia{Anios: 3, Meses: 1, Dias: 23}},
		{"2020-02-29", "2021-02-28", domain.TiempoExperiencia{Anios: 1}},
		{"2023-01-31", "2023-02-28", domain.TiempoExperiencia{Meses: 1}},
		{"2024-05-10", "2024-05-10", domain.TiempoExperiencia{}},
		{"2024-05-10", "2024-05-25", domain.TiempoExperiencia{Dias: 15}},
	}
	for _, c := range casos {
		tiempo, ok := domain.CalcularTiempoExperiencia(fechaISO(t, c.inicio), fechaISO(t, c.fin))
		if !ok || tiempo != c.esperado {
			t.Errorf("%s -> %s: obtenido %+v ok=%v, esperado %+v", c.inicio, c.fin, tiempo, ok, c.esperado)
		}
	}
}

func TestCalcularTiempoExperiencia_FinAnteriorAInicio(t *testing.T) {
	if _, ok := domain.CalcularTiempoExperiencia(fechaISO(t, "2024-05-10"), fechaISO(t, "2024-05-09")); ok {
		t.Error("una fecha de fin anterior al inicio debe devolver ok=false")
	}
}

func TestEfectoInhabilidad(t *testing.T) {
	no := false
	casos := []struct {
		nombre     string
		efecto     string
		bloqueante *bool
		respuesta  bool
		esperado   string
	}{
		{"1.9.1 con No bloquea", domain.EfectoBloqueante, &no, false, domain.EfectoBloqueante},
		{"1.9.1 con Sí no bloquea", domain.EfectoBloqueante, &no, true, domain.EfectoInformativo},
		{"bloqueante sin respuesta configurada", domain.EfectoBloqueante, nil, false, domain.EfectoInformativo},
		{"alerta con Sí", domain.EfectoAlerta, nil, true, domain.EfectoAlerta},
		{"alerta con No", domain.EfectoAlerta, nil, false, domain.EfectoInformativo},
		{"informativa", domain.EfectoInformativo, nil, true, domain.EfectoInformativo},
	}
	for _, c := range casos {
		if obtenido := domain.EfectoInhabilidad(c.efecto, c.bloqueante, c.respuesta); obtenido != c.esperado {
			t.Errorf("%s: obtenido %s, esperado %s", c.nombre, obtenido, c.esperado)
		}
	}
}

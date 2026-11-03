package domain_test

import (
	"testing"

	"srp-biess/internal/domain"
)

func TestEsPDF(t *testing.T) {
	casos := map[string]bool{"%PDF-1.7\n": true, "%PDF-": true, "%PDF": false, "\xFF\xD8\xFF\xE0": false, "": false, " %PDF-": false}
	for cabecera, esperado := range casos {
		if domain.EsPDF([]byte(cabecera)) != esperado {
			t.Errorf("EsPDF(%q) debía ser %v", cabecera, esperado)
		}
	}
}

func TestLimitesYTiposDocumento(t *testing.T) {
	l := domain.LimitesPorDefecto
	if l.MaximoPara(domain.DocRespaldoOrganizacional) != 4*1024*1024 || l.MaximoPara(domain.DocTitulo) != 2*1024*1024 || l.TotalBytes != 50*1024*1024 {
		t.Fatalf("límites inesperados: %+v", l)
	}
	if domain.DocFormularioPostulacion.CargablePorPostulante() || domain.TipoDocumento("OTRO").CargablePorPostulante() {
		t.Fatal("FORMULARIO_POSTULACION y tipos desconocidos no son cargables")
	}
	for _, tipo := range []domain.TipoDocumento{domain.DocRespaldoOrganizacional, domain.DocTitulo, domain.DocExperiencia, domain.DocPublicacion, domain.DocCapacitacion} {
		if !tipo.CargablePorPostulante() {
			t.Errorf("%s debía ser cargable", tipo)
		}
	}
}

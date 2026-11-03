package pdf_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/pdf"
)

func formularioDePrueba(items int) domain.FormularioPDF {
	f := domain.FormularioPDF{
		CodigoUnico: "BIESS-2026-AFI-000123",
		FechaEnvio:  time.Date(2026, 11, 3, 20, 5, 0, 0, time.UTC),
		Cargo:       "Miembro Principal y Alterno en Representación del Sector de Afiliados",
		Identificacion: []domain.CampoPDF{
			{Etiqueta: "Nombres y apellidos", Valor: "María José Peñafiel Núñez"},
			{Etiqueta: "Cédula de identidad", Valor: "0926687856"},
		},
		VersionTerminos:     "2026.1",
		TextoConsentimiento: "Texto de consentimiento de prueba.",
	}
	exp := domain.BloquePDF{Titulo: "Experiencia profesional"}
	for i := 0; i < items; i++ {
		exp.Subbloques = append(exp.Subbloques, domain.BloquePDF{Titulo: "Registro", Campos: []domain.CampoPDF{
			{Etiqueta: "Institución", Valor: "Instituto Ecuatoriano de Seguridad Social, Dirección de Tecnología"},
			{Etiqueta: "Cargo", Valor: "Analista"},
		}})
	}
	f.Secciones = []domain.BloquePDF{exp}
	return f
}

func TestGeneradorGopdf_PDFValidoYLegible(t *testing.T) {
	contenido, err := pdf.NewGeneradorGopdf().Generar(formularioDePrueba(60)) // fuerza varias páginas
	if err != nil {
		t.Fatalf("Generar: %v", err)
	}
	if !domain.EsPDF(contenido) || !bytes.Contains(contenido, []byte("%%EOF")) || !bytes.Contains(contenido, []byte("xref")) {
		t.Fatal("la salida no tiene la estructura de un PDF")
	}

	// Abrirlo con un lector real cuando esté disponible (poppler-utils).
	pdftotext, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Skip("pdftotext no instalado: se validó solo la estructura del PDF")
	}
	ruta := filepath.Join(t.TempDir(), "formulario.pdf")
	if err := os.WriteFile(ruta, contenido, 0o600); err != nil {
		t.Fatal(err)
	}
	salida, err := exec.Command(pdftotext, "-enc", "UTF-8", ruta, "-").Output()
	if err != nil {
		t.Fatalf("el PDF no se pudo abrir: %v", err)
	}
	texto := string(salida)
	for _, esperado := range []string{
		"Formulario de postulación", "BIESS-2026-AFI-000123", "María José Peñafiel Núñez",
		"3 de noviembre de 2026, 15:05 (hora de Ecuador)", "Firma del postulante", "Página 2",
	} {
		if !strings.Contains(texto, esperado) {
			t.Errorf("el texto extraído no contiene %q", esperado)
		}
	}
}

package email_test

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"
	"time"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/email"
)

func TestConstruirMensaje_MIMEConAdjunto(t *testing.T) {
	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte{0, 1, 2, 250}, 200)...)
	msg, err := email.ConstruirMensaje(
		mail.Address{Name: "Concurso Directorio BIESS", Address: "notificaciones@biess.gob.ec"},
		mail.Address{Address: "maria@correo.ec"},
		domain.Correo{
			Asunto:   "Confirmación de Postulación BIESS — BIESS-2026-AFI-000123",
			Texto:    "Su postulación fue enviada correctamente.",
			HTML:     "<p>Su postulación fue enviada correctamente.</p>",
			Adjuntos: []domain.Adjunto{{Nombre: "Postulacion-BIESS-BIESS-2026-AFI-000123.pdf", Tipo: "application/pdf", Contenido: pdf}},
		}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	m, err := mail.ReadMessage(bytes.NewReader(msg))
	if err != nil {
		t.Fatalf("el mensaje no es RFC 5322 válido: %v", err)
	}
	asunto, _ := new(mime.WordDecoder).DecodeHeader(m.Header.Get("Subject"))
	if asunto != "Confirmación de Postulación BIESS — BIESS-2026-AFI-000123" {
		t.Fatalf("asunto = %q", asunto)
	}
	tipo, params, _ := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if tipo != "multipart/mixed" {
		t.Fatalf("Content-Type = %s", tipo)
	}
	partes := multipart.NewReader(m.Body, params["boundary"])
	var vioAlternativa, vioAdjunto bool
	for {
		p, err := partes.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		ct, _, _ := mime.ParseMediaType(p.Header.Get("Content-Type"))
		switch ct {
		case "multipart/alternative":
			vioAlternativa = true
		case "application/pdf":
			vioAdjunto = true
			contenido, _ := io.ReadAll(p) // NextPart decodifica quoted-printable, no base64: se decodifica abajo
			if !strings.Contains(p.Header.Get("Content-Disposition"), "attachment") {
				t.Fatal("el PDF debía ir como adjunto")
			}
			decodificado, err := decodeBase64(contenido)
			if err != nil || !bytes.Equal(decodificado, pdf) {
				t.Fatalf("el adjunto no coincide con el PDF original: %v", err)
			}
		}
	}
	if !vioAlternativa || !vioAdjunto {
		t.Fatalf("faltan partes: alternativa=%v adjunto=%v", vioAlternativa, vioAdjunto)
	}
}

func TestConstruirMensaje_RechazaInyeccionDeCabeceras(t *testing.T) {
	_, err := email.ConstruirMensaje(mail.Address{Address: "a@b.ec"}, mail.Address{Address: "c@d.ec"},
		domain.Correo{Asunto: "Hola\r\nBcc: atacante@x.com"}, time.Now())
	if err == nil {
		t.Fatal("un asunto con salto de línea debía rechazarse")
	}
}

func TestPlantillas(t *testing.T) {
	p, err := email.NewPlantillas()
	if err != nil {
		t.Fatal(err)
	}
	html, texto, err := p.Renderizar("confirmacion_postulacion.html", map[string]any{
		"nombres": "<script>MARIA</script>", "codigo_unico": "BIESS-2026-AFI-000123",
		"fecha_envio_texto": "3 de noviembre de 2026, 16:41 (hora de Ecuador)", "cargo": "Afiliados",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "<script>") || !strings.Contains(html, "&lt;script&gt;") {
		t.Fatal("html/template debía escapar los datos")
	}
	for _, esperado := range []string{"BIESS-2026-AFI-000123", "3 de noviembre de 2026, 16:41", "Afiliados", "Imprímalo, fírmelo"} {
		if !strings.Contains(html, esperado) || !strings.Contains(texto, esperado) {
			t.Errorf("falta %q en el correo", esperado)
		}
	}
	if _, _, err := p.Renderizar("no_existe.html", nil); !errors.Is(err, domain.ErrPlantillaDesconocida) {
		t.Fatalf("esperado ErrPlantillaDesconocida, obtenido %v", err)
	}
}

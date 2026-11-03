// Package email envía los correos del outbox por SMTP (ADR-011, SRS RF-EF12-23).
package email

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"strings"
	texttemplate "text/template"

	"srp-biess/internal/domain"
)

//go:embed plantillas/*.html plantillas/*.txt
var archivos embed.FS

// Plantillas implementa domain.RenderizadorCorreo con las plantillas embebidas. html/template
// escapa los datos, así un nombre o un valor del payload no inyecta HTML.
type Plantillas struct {
	html  *htmltemplate.Template
	texto *texttemplate.Template
}

var _ domain.RenderizadorCorreo = (*Plantillas)(nil)

func NewPlantillas() (*Plantillas, error) {
	h, err := htmltemplate.ParseFS(archivos, "plantillas/*.html")
	if err != nil {
		return nil, fmt.Errorf("plantillas HTML: %w", err)
	}
	t, err := texttemplate.ParseFS(archivos, "plantillas/*.txt")
	if err != nil {
		return nil, fmt.Errorf("plantillas de texto: %w", err)
	}
	return &Plantillas{html: h.Option("missingkey=zero"), texto: t.Option("missingkey=zero")}, nil
}

// Renderizar recibe el nombre registrado en la notificación ("confirmacion_postulacion.html").
func (p *Plantillas) Renderizar(plantilla string, datos map[string]any) (string, string, error) {
	base := strings.TrimSuffix(plantilla, ".html")
	h := p.html.Lookup(base + ".html")
	t := p.texto.Lookup(base + ".txt")
	if h == nil || t == nil {
		return "", "", fmt.Errorf("%w: %s", domain.ErrPlantillaDesconocida, plantilla)
	}
	var html, texto bytes.Buffer
	if err := h.Execute(&html, datos); err != nil {
		return "", "", fmt.Errorf("renderizando %s: %w", plantilla, err)
	}
	if err := t.Execute(&texto, datos); err != nil {
		return "", "", fmt.Errorf("renderizando %s: %w", plantilla, err)
	}
	return html.String(), texto.String(), nil
}

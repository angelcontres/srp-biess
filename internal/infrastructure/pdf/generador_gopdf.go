// Package pdf renderiza el formulario de postulación con gopdf (SDD 5: PdfGenerator).
package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/signintech/gopdf"

	"srp-biess/internal/domain"
)

// Liberation Sans (SIL OFL 1.1, ver fuentes/LICENSE-LiberationSans): cubre tildes y eñes.
var (
	//go:embed fuentes/LiberationSans-Regular.ttf
	fuenteRegular []byte
	//go:embed fuentes/LiberationSans-Bold.ttf
	fuenteNegrita []byte
	// Logo de la SPA (frontend/src/assets/icono-biess.svg) en PNG: gopdf no admite SVG.
	//go:embed recursos/logo-biess.png
	logoBIESS []byte
)

// Colores de DESIGN.md: la cabecera replica el navbar de la SPA.
var (
	colorPrimary     = [3]uint8{11, 92, 173}   // primary #0B5CAD: franja del header
	colorPrimaryDark = [3]uint8{8, 63, 120}    // primary-dark #083F78: títulos de sección
	colorSecondary   = [3]uint8{242, 183, 5}   // secondary #F2B705: franja de 4 pt bajo el header
	colorTexto       = [3]uint8{15, 23, 42}    // text #0F172A
	colorMuted       = [3]uint8{71, 85, 105}   // muted #475569
	colorSobrePrim   = [3]uint8{204, 222, 239} // texto secundario sobre primary (como white/80)
)

const (
	cabeceraAlto = 60.0 // franja primary
	franjaAlto   = 4.0  // franja secondary
	inicioCuerpo = cabeceraAlto + franjaAlto + 26
	familia      = "liberation"
	familiaBold  = "liberation-bold"
	margen       = 50.0
	pieAltura    = 30.0
	cuerpo       = 10.0
	interlinea   = 1.45
)

// GeneradorGopdf implementa domain.GeneradorFormulario.
type GeneradorGopdf struct {
	zona *time.Location
}

var _ domain.GeneradorFormulario = (*GeneradorGopdf)(nil)

// NewGeneradorGopdf usa la hora de Ecuador para las fechas impresas.
func NewGeneradorGopdf() *GeneradorGopdf {
	zona, err := time.LoadLocation("America/Guayaquil")
	if err != nil {
		zona = time.FixedZone("ECT", -5*3600)
	}
	return &GeneradorGopdf{zona: zona}
}

type documento struct {
	logo   gopdf.ImageHolder
	pdf    *gopdf.GoPdf
	ancho  float64
	alto   float64
	pagina int
	codigo string
}

func (g *GeneradorGopdf) Generar(f domain.FormularioPDF) ([]byte, error) {
	d := &documento{pdf: &gopdf.GoPdf{}, ancho: gopdf.PageSizeA4.W, alto: gopdf.PageSizeA4.H, codigo: f.CodigoUnico}
	d.pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	d.pdf.SetInfo(gopdf.PdfInfo{
		Title:        domain.NombreFormularioPDF(f.CodigoUnico),
		Author:       "Banco del Instituto Ecuatoriano de Seguridad Social (BIESS)",
		Subject:      "Formulario de postulación al Concurso Público de Méritos y Oposición",
		Creator:      "SRP-BIESS",
		CreationDate: f.FechaEnvio,
	})
	if err := d.pdf.AddTTFFontData(familia, fuenteRegular); err != nil {
		return nil, fmt.Errorf("fuente regular: %w", err)
	}
	if err := d.pdf.AddTTFFontData(familiaBold, fuenteNegrita); err != nil {
		return nil, fmt.Errorf("fuente negrita: %w", err)
	}
	if logo, err := gopdf.ImageHolderByBytes(logoBIESS); err == nil {
		d.logo = logo
	}
	d.pdf.AddHeader(d.cabecera)
	d.pdf.AddFooter(d.pie)
	d.nuevaPagina()

	d.texto("Formulario de postulación", 16, true)
	d.espacio(4)
	d.campos([]domain.CampoPDF{
		{Etiqueta: "Código de la postulación", Valor: f.CodigoUnico},
		{Etiqueta: "Fecha y hora de envío", Valor: g.fecha(f.FechaEnvio)},
		{Etiqueta: "Estado", Valor: f.Estado},
		{Etiqueta: "Cargo al que postula", Valor: f.Cargo},
	})

	d.titulo("Datos de identificación")
	d.campos(f.Identificacion)

	for _, s := range f.Secciones {
		d.bloque(s, 0)
	}

	d.titulo("Consentimiento")
	d.texto(f.TextoConsentimiento, cuerpo, false)
	if f.VersionTerminos != "" {
		d.campos([]domain.CampoPDF{{Etiqueta: "Versión de los términos aceptados", Valor: f.VersionTerminos}})
	}

	d.firma(f)

	var buf bytes.Buffer
	if err := d.pdf.Write(&buf); err != nil {
		return nil, fmt.Errorf("no se pudo escribir el PDF: %w", err)
	}
	return buf.Bytes(), nil
}

func (g *GeneradorGopdf) fecha(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	meses := []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto",
		"septiembre", "octubre", "noviembre", "diciembre"}
	l := t.In(g.zona)
	return fmt.Sprintf("%d de %s de %d, %02d:%02d (hora de Ecuador)", l.Day(), meses[l.Month()-1], l.Year(), l.Hour(), l.Minute())
}

func (d *documento) nuevaPagina() {
	d.pagina++ // antes de AddPage: la cabecera y el pie se dibujan dentro de AddPage
	d.pdf.AddPage()
	d.pdf.SetXY(margen, inicioCuerpo)
}

func (d *documento) relleno(c [3]uint8) { d.pdf.SetFillColor(c[0], c[1], c[2]) }
func (d *documento) tinta(c [3]uint8)   { d.pdf.SetTextColor(c[0], c[1], c[2]) }

// cabecera replica el navbar de la SPA: franja primary con logo y nombre del concurso, y la
// franja secondary de 4 pt debajo (DESIGN.md §1).
func (d *documento) cabecera() {
	d.relleno(colorPrimary)
	d.pdf.RectFromUpperLeftWithStyle(0, 0, d.ancho, cabeceraAlto, "F")
	d.relleno(colorSecondary)
	d.pdf.RectFromUpperLeftWithStyle(0, cabeceraAlto, d.ancho, franjaAlto, "F")
	x := margen
	if d.logo != nil {
		_ = d.pdf.ImageByHolder(d.logo, margen, 12, &gopdf.Rect{W: 36, H: 36})
		x += 46
	}
	_ = d.pdf.SetFont(familia, "", 8)
	d.tinta(colorSobrePrim)
	d.pdf.SetXY(x, 17)
	_ = d.pdf.Cell(nil, "Banco del IESS")
	_ = d.pdf.SetFont(familiaBold, "", 13)
	d.pdf.SetTextColor(255, 255, 255)
	d.pdf.SetXY(x, 29)
	_ = d.pdf.Cell(nil, "Concurso Público de Méritos y Oposición")
	d.tinta(colorTexto)
}

func (d *documento) pie() {
	_ = d.pdf.SetFont(familia, "", 8)
	d.tinta(colorMuted)
	d.pdf.SetXY(margen, d.alto-margen+10)
	_ = d.pdf.Cell(nil, fmt.Sprintf("Formulario %s. Página %d", d.codigo, d.pagina))
	d.tinta(colorTexto)
}

// asegurar pasa a una página nueva si no caben `alto` puntos antes del pie.
func (d *documento) asegurar(alto float64) {
	if d.pdf.GetY()+alto > d.alto-margen-pieAltura {
		d.nuevaPagina()
	}
}

func (d *documento) espacio(h float64) {
	d.pdf.SetXY(margen, d.pdf.GetY()+h)
}

func (d *documento) texto(t string, tam float64, negrita bool) {
	d.textoEn(t, tam, negrita, margen)
}

func (d *documento) textoEn(t string, tam float64, negrita bool, x float64) {
	f := familia
	if negrita {
		f = familiaBold
	}
	_ = d.pdf.SetFont(f, "", tam)
	alto := tam * interlinea
	for _, parrafo := range strings.Split(strings.TrimSpace(t), "\n") {
		lineas, err := d.pdf.SplitTextWithWordWrap(parrafo, d.ancho-margen-x)
		if err != nil || len(lineas) == 0 {
			lineas = []string{parrafo}
		}
		for _, l := range lineas {
			d.asegurar(alto)
			d.pdf.SetXY(x, d.pdf.GetY())
			_ = d.pdf.Cell(nil, l)
			d.pdf.SetXY(x, d.pdf.GetY()+alto)
		}
	}
}

func (d *documento) titulo(t string) {
	// Título (≈ 48 pt con el espacio y la línea) más al menos una línea de contenido: nunca huérfano.
	d.asegurar(48 + cuerpo*interlinea)
	d.espacio(10)
	d.tinta(colorPrimaryDark)
	d.texto(t, 12, true)
	d.tinta(colorTexto)
	d.pdf.SetStrokeColor(colorPrimary[0], colorPrimary[1], colorPrimary[2])
	d.pdf.SetLineWidth(0.5)
	d.pdf.Line(margen, d.pdf.GetY(), d.ancho-margen, d.pdf.GetY())
	d.pdf.SetStrokeColor(colorTexto[0], colorTexto[1], colorTexto[2])
	d.espacio(6)
}

func (d *documento) campos(cs []domain.CampoPDF) {
	d.camposEn(cs, margen)
}

func (d *documento) camposEn(cs []domain.CampoPDF, x float64) {
	for _, c := range cs {
		valor := strings.TrimSpace(c.Valor)
		if valor == "" {
			valor = "No registrado"
		}
		d.textoEn(c.Etiqueta+": "+valor, cuerpo, false, x)
	}
}

func (d *documento) bloque(b domain.BloquePDF, nivel int) {
	x := margen + float64(nivel)*14
	if nivel == 0 {
		d.titulo(b.Titulo)
	} else {
		d.espacio(4)
		d.textoEn(b.Titulo, cuerpo, true, x)
	}
	d.camposEn(b.Campos, x)
	for _, s := range b.Subbloques {
		d.bloque(s, nivel+1)
	}
}

func (d *documento) firma(f domain.FormularioPDF) {
	d.asegurar(110)
	d.espacio(40)
	y := d.pdf.GetY()
	d.pdf.SetLineWidth(0.7)
	d.pdf.Line(margen, y, margen+220, y)
	d.espacio(6)
	d.texto("Firma del postulante", cuerpo, true)
	for _, c := range f.Identificacion {
		if c.Etiqueta == "Nombres y apellidos" || c.Etiqueta == "Cédula de identidad" {
			d.campos([]domain.CampoPDF{c})
		}
	}
	d.campos([]domain.CampoPDF{{Etiqueta: "Fecha", Valor: ""}})
}

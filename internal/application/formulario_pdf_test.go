package application_test

import (
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
)

func TestEtiqueta(t *testing.T) {
	casos := map[string]string{
		"telefonoCelular": "Teléfono celular", "callePrincipal": "Calle principal",
		"institucion": "Institución", "fechaInicio": "Fecha inicio", "p1": "P 1",
	}
	for clave, esperado := range casos {
		if got := application.Etiqueta(clave); got != esperado {
			t.Errorf("Etiqueta(%q) = %q, esperado %q", clave, got, esperado)
		}
	}
}

func TestArmarFormularioPDF(t *testing.T) {
	detalle := &domain.DetalleFormulario{
		Verificados: &domain.DatosVerificados{Cedula: "0926687856", Nombres: "MARÍA", Apellidos: "PEÑAFIEL", FechaNacimiento: "1990-05-04"},
		Secciones: map[string][]byte{
			// Orden de JSONB (claves cortas primero), no el del formulario.
			"domicilio":     []byte(`{"numero":"555","paisId":"52","sector":"Centro","cantonId":"2301","esPropia":true,"provinciaId":"23","callePrincipal":"Av. Amazonas"}`),
			"inhabilidades": []byte(`{"respuestas":["SI","NO","NO","NO","NO","NO","SI","NO","NO","NO","NO","NO","NO"],"remocionCargo":"Director"}`),
			"personal":      []byte(`{"estadoCivil":"UNION_DE_HECHO","parejaNombres":"JUAN IESS","aniosServicio":12}`),
			"experiencia":   []byte(`{"items":[{"institucion":"IESS","fechaInicio":"2020-01-15","documento":{"id":"d1","name":"cert.pdf","sizeBytes":10}}]}`),
		},
	}
	ubicar := func(tipo, id, padre string) string {
		return map[string]string{"pais|52|": "Ecuador", "provincia|23|52": "Santo Domingo de los Tsáchilas", "canton|2301|23": "Santo Domingo"}[tipo+"|"+id+"|"+padre]
	}
	f := application.ArmarFormularioPDF(detalle, "BIESS-2026-AFI-000001", time.Now(), "Afiliados", "2026.1", ubicar)
	if len(f.Secciones) != len(domain.SeccionesFormulario) {
		t.Fatalf("debían incluirse todas las secciones: %d", len(f.Secciones))
	}
	buscar := func(titulo string) domain.BloquePDF {
		for _, s := range f.Secciones {
			if s.Titulo == titulo {
				return s
			}
		}
		t.Fatalf("falta la sección %s", titulo)
		return domain.BloquePDF{}
	}
	dom := buscar("Información domiciliaria y de contacto")
	// Orden del formulario, e identificadores del catálogo como nombres (nunca "País id: 52").
	esperado := []domain.CampoPDF{
		{Etiqueta: "País", Valor: "Ecuador"},
		{Etiqueta: "Provincia", Valor: "Santo Domingo de los Tsáchilas"},
		{Etiqueta: "Cantón", Valor: "Santo Domingo"},
		{Etiqueta: "Calle principal", Valor: "Av. Amazonas"},
		{Etiqueta: "Número", Valor: "555"},
		{Etiqueta: "Sector", Valor: "Centro"},
		{Etiqueta: "Es propia", Valor: "Sí"}, // clave no listada: al final
	}
	if len(dom.Campos) != len(esperado) {
		t.Fatalf("domicilio inesperado: %+v", dom.Campos)
	}
	for i := range esperado {
		if dom.Campos[i] != esperado[i] {
			t.Fatalf("domicilio[%d] = %+v, esperado %+v", i, dom.Campos[i], esperado[i])
		}
	}
	exp := buscar("Experiencia profesional")
	if len(exp.Subbloques) != 1 {
		t.Fatalf("experiencia debía tener un registro: %+v", exp)
	}
	campos := map[string]string{}
	for _, c := range exp.Subbloques[0].Campos {
		campos[c.Etiqueta] = c.Valor
	}
	if campos["Fecha inicio"] != "15/01/2020" || campos["Documento"] != "Documento adjunto: cert.pdf" || campos["Institución"] != "IESS" {
		t.Fatalf("registro de experiencia inesperado: %+v", campos)
	}
	// Orden del formulario (no alfabético), etiquetas de catálogo de la SPA y texto libre intacto.
	per := buscar("Información personal").Campos
	if len(per) != 3 || per[0] != (domain.CampoPDF{Etiqueta: "Estado civil", Valor: "Unión de hecho"}) ||
		per[1].Valor != "JUAN IESS" || per[2].Valor != "12" {
		t.Fatalf("información personal inesperada: %+v", per)
	}
	if exp.Subbloques[0].Titulo != "Items: registro 1 de 1" {
		t.Fatalf("título del registro inesperado: %s", exp.Subbloques[0].Titulo)
	}
	inh := buscar("Inhabilidades").Campos
	if len(inh) != 14 || inh[0].Etiqueta != "1.9.1 ¿Es ecuatoriano/a y está en pleno goce de los derechos de participación política?" ||
		inh[0].Valor != "Sí" || inh[6].Etiqueta != "1.9.7 ¿Removido por Organismo de Control en los últimos 5 años?" || inh[6].Valor != "Sí" ||
		inh[13] != (domain.CampoPDF{Etiqueta: "Remoción cargo", Valor: "Director"}) {
		t.Fatalf("inhabilidades inesperadas: %+v", inh)
	}
	if buscar("Familiares en el IESS o el BIESS").Campos[0].Valor != "Sin información registrada" {
		t.Fatal("una sección sin datos debía indicarlo")
	}
	if f.TextoConsentimiento != application.TextoConsentimientoPendiente || f.Identificacion[0].Valor != "MARÍA PEÑAFIEL" {
		t.Fatalf("cabecera inesperada: %+v", f.Identificacion)
	}
}

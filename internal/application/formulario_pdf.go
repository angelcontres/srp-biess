package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"srp-biess/internal/domain"
)

// Armado del formulario PDF (SRS 3.5.21): identificación, todas las secciones guardadas y el
// consentimiento. Las etiquetas se derivan de los nombres de campo del formulario; las etiquetas
// oficiales de cada campo no están definidas en el SDD.

// TextoConsentimientoPendiente: los textos de consentimiento son [V] (SDD 8.5.8).
const TextoConsentimientoPendiente = "El postulante aceptó los consentimientos vigentes al enviar su postulación. " +
	"[PENDIENTE: texto de los consentimientos según la normativa de protección de datos personales y las bases del concurso (SDD 8.5.8)]"

// TextoConsentimientos numera los consentimientos aceptados para el PDF.
func TextoConsentimientos(textos []string) string {
	lineas := make([]string, 0, len(textos)+1)
	lineas = append(lineas, "El postulante aceptó los siguientes consentimientos al enviar su postulación:")
	for i, t := range textos {
		lineas = append(lineas, fmt.Sprintf("%d. %s", i+1, t))
	}
	return strings.Join(lineas, "\n")
}

// Títulos de las secciones, en el orden del formulario (SRS 3.5.1). La clave `cargo` guarda el
// respaldo organizacional; la selección de cargo va en la cabecera.
var titulosSeccion = map[string]string{
	"personal":      "Información personal",
	"domicilio":     "Información domiciliaria y de contacto",
	"familiares":    "Familiares en el IESS o el BIESS",
	"cargo":         "Respaldo organizacional",
	"formacion":     "Formación académica",
	"experiencia":   "Experiencia profesional",
	"publicaciones": "Publicaciones e investigaciones",
	"capacitacion":  "Capacitación",
	"inhabilidades": "Inhabilidades",
}

// Campos técnicos que no aportan al formulario impreso.
var camposOmitidos = map[string]bool{"id": true, "url": true, "uploadedAt": true, "sizeBytes": true}

// Tildes de las palabras frecuentes en los nombres de campo (los nombres de campo no las llevan).
var tildes = map[string]string{
	"telefono": "teléfono", "direccion": "dirección", "numero": "número", "institucion": "institución",
	"titulo": "título", "academica": "académica", "academico": "académico", "electronico": "electrónico",
	"organizacion": "organización", "relacion": "relación", "descripcion": "descripción", "pais": "país",
	"canton": "cantón", "codigo": "código", "registro": "registro", "senescyt": "SENESCYT", "iess": "IESS",
	"biess": "BIESS", "anio": "año", "duracion": "duración", "publicacion": "publicación", "area": "área",
	"genero": "género", "funcion": "función", "remocion": "remoción", "cedula": "cédula", "razon": "razón", "via": "vía", "parentesco": "parentesco",
}

var fechaISO = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)

// etiquetasCatalogo son las etiquetas que muestra la SPA para los valores de catálogo
// (frontend/src/features/postulacion/constants.ts). Mantener sincronizado.
var etiquetasCatalogo = map[string]string{
	"BIESS":                 "BIESS",
	"CASADO":                "Casado/a",
	"CONVIVIENTE":           "Conviviente",
	"CONYUGE":               "Cónyuge",
	"CUNADO":                "Cuñado/a",
	"CURSO":                 "Curso",
	"DIRECCION":             "De dirección",
	"DIVORCIADO":            "Divorciado/a",
	"ESPECIALIDAD":          "Especialidad",
	"EXTRANJERA":            "Extranjera",
	"FEMENINO":              "Femenino",
	"GENERAL":               "General",
	"HERMANO":               "Hermano/a",
	"HIJO":                  "Hijo/a",
	"HIJO_POLITICO":         "Hijo/a político/a",
	"IESS":                  "IESS",
	"MADRE":                 "Madre",
	"MAESTRIA":              "Maestría",
	"MASCULINO":             "Masculino",
	"NACIONAL":              "Nacional",
	"NIETO":                 "Nieto/a",
	"NUERA":                 "Nuera",
	"OTRO":                  "Otro",
	"PADRE":                 "Padre",
	"PHD":                   "Doctorado (PhD)",
	"PREFIERO_NO_DECIR":     "Prefiero no decirlo",
	"PRIMO_HERMANO":         "Primo/a hermano/a",
	"PUBLICACION":           "Publicación",
	"SEMINARIO":             "Seminario",
	"SOBRINO":               "Sobrino/a",
	"SOLTERO":               "Soltero/a",
	"SUEGRO":                "Suegro/a",
	"TALLER":                "Taller",
	"TERCER_NIVEL":          "Tercer nivel",
	"TIO":                   "Tío/a",
	"TRABAJO_INVESTIGACION": "Trabajo de investigación",
	"UNION_DE_HECHO":        "Unión de hecho",
	"VIUDO":                 "Viudo/a",
	"YERNO":                 "Yerno",
}

// ordenCampos es el orden de los campos del formulario (frontend/src/features/postulacion/schemas).
// JSONB no conserva el orden de las claves, así que el PDF lo toma de aquí; las claves de cada lista
// se buscan como "<lista>[]". Las claves no listadas van al final. Mantener sincronizado.
var ordenCampos = map[string][]string{
	"personal":         {"estadoCivil", "parejaNombres", "parejaCedula"},
	"domicilio":        {"paisId", "provinciaId", "cantonId", "provinciaManual", "ciudadManual", "callePrincipal", "numero", "calleSecundaria", "sector", "referencia", "telefonoCelular", "telefonoDomicilio", "correoAlterno"},
	"familiares":       {"tieneFamiliares", "familiares"},
	"familiares[]":     {"nombresApellidos", "parentesco", "institucion", "area", "cargo"},
	"cargo":            {"tieneRespaldo", "nombreOrganizacion", "documento"},
	"formacion":        {"formaciones"},
	"formaciones[]":    {"nivel", "institucion", "tipoInstitucion", "titulo", "area", "numeroSenescyt", "fechaSenescyt", "documento"},
	"experiencia":      {"experiencias"},
	"experiencias[]":   {"tipo", "cargo", "institucion", "fechaInicio", "actualmente", "fechaFin", "actividades", "funciones", "documento"},
	"publicaciones":    {"tienePublicaciones", "publicaciones"},
	"publicaciones[]":  {"titulo", "tipo", "medio", "fecha", "relacion", "documento"},
	"capacitacion":     {"capacitaciones"},
	"capacitaciones[]": {"tipoEvento", "nombre", "institucion", "fechaInicio", "fechaFin", "horas", "documento"},
	"inhabilidades":    {"respuestas", "remocionInstitucion", "remocionCargo", "remocionFechaFin", "remocionMotivo"},
}

// preguntasInhabilidad son las 13 preguntas (SRS 7.4; frontend constants INHABILIDADES), en orden.
// La sección guarda solo las respuestas; el PDF imprime pregunta y respuesta. Mantener sincronizado.
var preguntasInhabilidad = []string{
	"1.9.1 ¿Es ecuatoriano/a y está en pleno goce de los derechos de participación política?",
	"1.9.2 ¿Inhabilitado para ejercer el comercio?",
	"1.9.3 ¿En mora con el Estado o instituciones bajo control de la Superintendencia de Bancos?",
	"1.9.4 ¿Vínculo con sistema financiero, seguros, bolsas, casas de valores o fondos?",
	"1.9.5 ¿Funcionario o empleado del IESS-BIESS?",
	"1.9.6 ¿Interés propio o representación de terceros en aseguradoras o seguridad social?",
	"1.9.7 ¿Removido por Organismo de Control en los últimos 5 años?",
	"1.9.8 ¿Obligaciones pendientes con el SRI?",
	"1.9.9 ¿Castigo de obligaciones por institución financiera en los últimos 5 años?",
	"1.9.10 ¿Litigando contra el IESS-BIESS?",
	"1.9.11 ¿Procesado, enjuiciado o condenado por corrupción o crimen organizado?",
	"1.9.12 ¿Responsabilidades en firme de la Contraloría?",
	"1.9.13 ¿Registrado en la base de la UAFE?",
}

// ordenar devuelve las claves en el orden del formulario y luego las restantes.
func ordenar(claves, orden []string) []string {
	if len(orden) == 0 {
		return claves
	}
	presentes := make(map[string]bool, len(claves))
	for _, k := range claves {
		presentes[k] = true
	}
	resultado := make([]string, 0, len(claves))
	usadas := map[string]bool{}
	for _, k := range orden {
		if presentes[k] {
			resultado = append(resultado, k)
			usadas[k] = true
		}
	}
	for _, k := range claves {
		if !usadas[k] {
			resultado = append(resultado, k)
		}
	}
	return resultado
}

// UbicadorCatalogo traduce los identificadores del catálogo geográfico a nombres. `tipo` es
// "pais", "provincia" o "canton"; `padre` es el país de la provincia o la provincia del cantón.
// Devuelve "" si no lo encuentra.
type UbicadorCatalogo func(tipo, id, padre string) string

// camposUbicacion: el formulario guarda los identificadores del catálogo; el PDF imprime el nombre.
var camposUbicacion = map[string]struct{ etiqueta, tipo, padre string }{
	"paisId":      {"País", "pais", ""},
	"provinciaId": {"Provincia", "provincia", "paisId"},
	"cantonId":    {"Cantón", "canton", "provinciaId"},
}

// ArmarFormularioPDF compone los datos del PDF a partir del detalle guardado de la postulación.
// `ubicar` puede ser nil: los identificadores geográficos se imprimen entonces como "No registrado".
func ArmarFormularioPDF(detalle *domain.DetalleFormulario, codigo string, fechaEnvio time.Time, cargo, versionTerminos string, ubicar UbicadorCatalogo) domain.FormularioPDF {
	f := domain.FormularioPDF{
		CodigoUnico: codigo, FechaEnvio: fechaEnvio, Cargo: cargo, Estado: "Enviada",
		VersionTerminos: versionTerminos, TextoConsentimiento: TextoConsentimientoPendiente,
	}
	if v := detalle.Verificados; v != nil {
		f.Identificacion = []domain.CampoPDF{
			{Etiqueta: "Nombres y apellidos", Valor: strings.TrimSpace(v.Nombres + " " + v.Apellidos)},
			{Etiqueta: "Cédula de identidad", Valor: v.Cedula},
			{Etiqueta: "Género", Valor: formatoValor(v.Genero)},
			{Etiqueta: "Nacionalidad", Valor: v.Nacionalidad},
			{Etiqueta: "Fecha de nacimiento", Valor: formatoValor(v.FechaNacimiento)},
			{Etiqueta: "Correo electrónico", Valor: v.Correo},
		}
	}
	for _, clave := range domain.SeccionesFormulario {
		b := domain.BloquePDF{Titulo: titulosSeccion[clave]}
		datos, err := decodificarOrdenado(detalle.Secciones[clave])
		if err != nil || datos == nil || len(datos.claves) == 0 {
			b.Campos = []domain.CampoPDF{{Etiqueta: "Estado", Valor: "Sin información registrada"}}
		} else {
			b.Campos, b.Subbloques = camposDe(datos, ubicar, clave)
		}
		f.Secciones = append(f.Secciones, b)
	}
	return f
}

// objeto conserva el orden de las claves tal como las guardó el formulario.
type objeto struct {
	claves  []string
	valores map[string]any // valores: *objeto, []any o primitivos
}

// decodificarOrdenado lee JSON preservando el orden de las claves (un map de Go no lo conserva).
func decodificarOrdenado(raw []byte) (*objeto, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	v, err := leerValor(dec)
	if err != nil {
		return nil, err
	}
	o, _ := v.(*objeto)
	return o, nil
}

func leerValor(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			o := &objeto{valores: map[string]any{}}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				k, _ := kt.(string)
				v, err := leerValor(dec)
				if err != nil {
					return nil, err
				}
				if _, repetida := o.valores[k]; !repetida {
					o.claves = append(o.claves, k)
				}
				o.valores[k] = v
			}
			_, err := dec.Token() // '}'
			return o, err
		case '[':
			var lista []any
			for dec.More() {
				v, err := leerValor(dec)
				if err != nil {
					return nil, err
				}
				lista = append(lista, v)
			}
			_, err := dec.Token() // ']'
			return lista, err
		}
	}
	return tok, nil
}

// camposDe convierte un objeto en campos, en el orden del formulario; las listas de objetos pasan
// a subbloques numerados.
func camposDe(datos *objeto, ubicar UbicadorCatalogo, contexto string) ([]domain.CampoPDF, []domain.BloquePDF) {
	var campos []domain.CampoPDF
	var sub []domain.BloquePDF
	for _, k := range ordenar(datos.claves, ordenCampos[contexto]) {
		if camposOmitidos[k] {
			continue
		}
		// Inhabilidades: cada respuesta con su pregunta.
		if contexto == "inhabilidades" && k == "respuestas" {
			respuestas, _ := datos.valores[k].([]any)
			for i, r := range respuestas {
				pregunta := fmt.Sprintf("Pregunta %d", i+1)
				if i < len(preguntasInhabilidad) {
					pregunta = preguntasInhabilidad[i]
				}
				campos = append(campos, domain.CampoPDF{Etiqueta: pregunta, Valor: formatoValor(r)})
			}
			continue
		}
		if u, ok := camposUbicacion[k]; ok {
			id := formatoValor(datos.valores[k])
			if id == "" {
				continue
			}
			nombre := ""
			if ubicar != nil {
				nombre = ubicar(u.tipo, id, formatoValor(datos.valores[u.padre]))
			}
			campos = append(campos, domain.CampoPDF{Etiqueta: u.etiqueta, Valor: nombre})
			continue
		}
		etiqueta := Etiqueta(k)
		switch v := datos.valores[k].(type) {
		case []any:
			if len(v) == 0 {
				campos = append(campos, domain.CampoPDF{Etiqueta: etiqueta, Valor: "Ninguno"})
				continue
			}
			if _, esObjeto := v[0].(*objeto); esObjeto {
				for i, item := range v {
					obj, _ := item.(*objeto)
					if obj == nil {
						continue
					}
					c, s := camposDe(obj, ubicar, k+"[]")
					sub = append(sub, domain.BloquePDF{Titulo: fmt.Sprintf("%s: registro %d de %d", etiqueta, i+1, len(v)), Campos: c, Subbloques: s})
				}
				continue
			}
			valores := make([]string, 0, len(v))
			for _, item := range v {
				valores = append(valores, formatoValor(item))
			}
			campos = append(campos, domain.CampoPDF{Etiqueta: etiqueta, Valor: strings.Join(valores, ", ")})
		case *objeto:
			// Documento adjunto ({ id, name, sizeBytes, ... }): se imprime su nombre.
			if nombre, ok := v.valores["name"].(string); ok {
				campos = append(campos, domain.CampoPDF{Etiqueta: etiqueta, Valor: "Documento adjunto: " + nombre})
				continue
			}
			c, s := camposDe(v, ubicar, k)
			sub = append(sub, domain.BloquePDF{Titulo: etiqueta, Campos: c, Subbloques: s})
		default:
			// Un campo vacío en una sección completa es opcional o no aplica (pareja de una persona
			// soltera, provincia manual con país Ecuador): no se imprime.
			valor := formatoValor(v)
			if valor == "" {
				continue
			}
			campos = append(campos, domain.CampoPDF{Etiqueta: etiqueta, Valor: valor})
		}
	}
	return campos, sub
}

func formatoValor(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case bool:
		if x {
			return "Sí"
		}
		return "No"
	case json.Number:
		return x.String()
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case string:
		switch strings.ToUpper(strings.TrimSpace(x)) {
		case "SI", "SÍ":
			return "Sí"
		case "NO":
			return "No"
		}
		if m := fechaISO.FindStringSubmatch(x); m != nil {
			return m[3] + "/" + m[2] + "/" + m[1]
		}
		if etiqueta, ok := etiquetasCatalogo[x]; ok {
			return etiqueta // CASADO → Casado/a; un texto libre en mayúsculas no se toca
		}
		return x
	default:
		return fmt.Sprint(x)
	}
}

// Etiqueta convierte un nombre de campo camelCase en texto con mayúscula inicial: "telefonoCelular"
// → "Teléfono celular".
func Etiqueta(clave string) string {
	var palabras []string
	var actual []rune
	for i, r := range clave {
		if i > 0 && (unicode.IsUpper(r) || (unicode.IsDigit(r) && len(actual) > 0 && !unicode.IsDigit(actual[len(actual)-1]))) {
			palabras = append(palabras, string(actual))
			actual = nil
		}
		actual = append(actual, unicode.ToLower(r))
	}
	if len(actual) > 0 {
		palabras = append(palabras, string(actual))
	}
	for i, p := range palabras {
		if t, ok := tildes[p]; ok {
			palabras[i] = t
		}
	}
	texto := strings.Join(palabras, " ")
	if texto == "" {
		return clave
	}
	r := []rune(texto)
	if unicode.IsLower(r[0]) {
		r[0] = unicode.ToUpper(r[0])
	}
	return string(r)
}

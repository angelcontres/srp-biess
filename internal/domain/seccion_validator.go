package domain

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

//go:embed secciones_reglas.json
var ReglasJSON []byte

// ErrorCampo representa un fallo de validación en un campo específico.
type ErrorCampo struct {
	Campo   string `json:"campo"`
	Mensaje string `json:"mensaje"`
}

// ErrorValidacionSeccion consolida los errores de validación de una sección.
type ErrorValidacionSeccion struct {
	Detalles []ErrorCampo `json:"detalles"`
}

func (e *ErrorValidacionSeccion) Error() string {
	return "los datos de la sección no son válidos"
}

// ContextoPostulante contiene datos de identidad para validaciones cruzadas.
type ContextoPostulante struct {
	CedulaPostulante string
	CorreoPrincipal  string
	PaisEcuadorID    string
}

var (
	reTelefono = regexp.MustCompile(`^(\+\d{8,15}|0\d{8,9})$`)
	reCedula   = regexp.MustCompile(`^\d{10}$`)
	reHoras    = regexp.MustCompile(`^[1-9]\d*$`)
	reEmail    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func parsearFecha(valor string) (time.Time, error) {
	v := strings.TrimSpace(valor)
	if v == "" {
		return time.Time{}, fmt.Errorf("fecha vacía")
	}
	if t, err := time.Parse("2006-01-02", v); err == nil {
		return t, nil
	}
	if t, err := time.Parse("02/01/2006", v); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("formato no reconocido")
}

func esFechaFutura(t time.Time) bool {
	hoy := time.Now().Format("2006-01-02")
	return t.Format("2006-01-02") > hoy
}

func esDocumentoValido(raw json.RawMessage) bool {
	if len(raw) == 0 || string(raw) == "null" {
		return false
	}
	var docObj struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &docObj); err == nil && strings.TrimSpace(docObj.ID) != "" {
		return true
	}
	var strID string
	if err := json.Unmarshal(raw, &strID); err == nil && strings.TrimSpace(strID) != "" {
		return true
	}
	return false
}

// ValidarSeccion valida los datos de una sección antes de permitir marcarla como COMPLETA (RF-EF12-24).
func ValidarSeccion(seccion string, datos []byte, ctx ContextoPostulante) error {
	var detalles []ErrorCampo

	switch seccion {
	case "personal":
		detalles = validarPersonal(datos, ctx)
	case "domicilio":
		detalles = validarDomicilio(datos, ctx)
	case "familiares":
		detalles = validarFamiliares(datos)
	case "cargo":
		detalles = validarCargo(datos)
	case "formacion":
		detalles = validarFormacion(datos)
	case "experiencia":
		detalles = validarExperiencia(datos)
	case "publicaciones":
		detalles = validarPublicaciones(datos)
	case "capacitacion":
		detalles = validarCapacitacion(datos)
	case "inhabilidades":
		detalles = validarInhabilidades(datos)
	default:
		return ErrSeccionInvalida
	}

	if len(detalles) > 0 {
		return &ErrorValidacionSeccion{Detalles: detalles}
	}
	return nil
}

func validarPersonal(datos []byte, ctx ContextoPostulante) []ErrorCampo {
	var d struct {
		EstadoCivil   string `json:"estadoCivil"`
		ParejaNombres string `json:"parejaNombres"`
		ParejaCedula  string `json:"parejaCedula"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	estadosValidos := map[string]bool{
		"SOLTERO": true, "CASADO": true, "UNION_DE_HECHO": true,
		"DIVORCIADO": true, "VIUDO": true, "OTRO": true,
	}

	ec := strings.TrimSpace(d.EstadoCivil)
	if ec == "" || !estadosValidos[ec] {
		errs = append(errs, ErrorCampo{Campo: "estadoCivil", Mensaje: "Seleccione su estado civil."})
	}

	if ec == "CASADO" || ec == "UNION_DE_HECHO" {
		pn := strings.TrimSpace(d.ParejaNombres)
		if pn == "" {
			errs = append(errs, ErrorCampo{Campo: "parejaNombres", Mensaje: "Ingrese los nombres y apellidos de su pareja."})
		} else if len(pn) > 255 {
			errs = append(errs, ErrorCampo{Campo: "parejaNombres", Mensaje: "Use como máximo 255 caracteres."})
		}

		pc := strings.TrimSpace(d.ParejaCedula)
		if pc == "" {
			errs = append(errs, ErrorCampo{Campo: "parejaCedula", Mensaje: "Ingrese la cédula de su pareja."})
		} else if !reCedula.MatchString(pc) {
			errs = append(errs, ErrorCampo{Campo: "parejaCedula", Mensaje: "La cédula debe tener 10 números. Revise que no tenga letras ni espacios."})
		} else if ctx.CedulaPostulante != "" && pc == ctx.CedulaPostulante {
			errs = append(errs, ErrorCampo{Campo: "parejaCedula", Mensaje: "La cédula de su pareja debe ser distinta de la suya."})
		}
	}
	return errs
}

func validarDomicilio(datos []byte, ctx ContextoPostulante) []ErrorCampo {
	var d struct {
		PaisID            string `json:"paisId"`
		ProvinciaID       string `json:"provinciaId"`
		CantonID          string `json:"cantonId"`
		ProvinciaManual   string `json:"provinciaManual"`
		CiudadManual      string `json:"ciudadManual"`
		TelefonoCelular   string `json:"telefonoCelular"`
		CorreoAlterno     string `json:"correoAlterno"`
		CallePrincipal    string `json:"callePrincipal"`
		Numero            string `json:"numero"`
		CalleSecundaria   string `json:"calleSecundaria"`
		Sector            string `json:"sector"`
		Referencia        string `json:"referencia"`
		TelefonoDomicilio string `json:"telefonoDomicilio"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	pais := strings.TrimSpace(d.PaisID)
	if pais == "" {
		errs = append(errs, ErrorCampo{Campo: "paisId", Mensaje: "Seleccione el país de residencia."})
	} else {
		esEcuador := pais == "EC" || pais == "1" || (ctx.PaisEcuadorID != "" && pais == ctx.PaisEcuadorID)
		if esEcuador {
			if strings.TrimSpace(d.ProvinciaID) == "" {
				errs = append(errs, ErrorCampo{Campo: "provinciaId", Mensaje: "Seleccione la provincia."})
			}
			if strings.TrimSpace(d.CantonID) == "" {
				errs = append(errs, ErrorCampo{Campo: "cantonId", Mensaje: "Seleccione el cantón."})
			}
		} else {
			if strings.TrimSpace(d.ProvinciaManual) == "" {
				errs = append(errs, ErrorCampo{Campo: "provinciaManual", Mensaje: "Ingrese la provincia o estado."})
			} else if len(d.ProvinciaManual) > 120 {
				errs = append(errs, ErrorCampo{Campo: "provinciaManual", Mensaje: "Use como máximo 120 caracteres."})
			}
			if strings.TrimSpace(d.CiudadManual) == "" {
				errs = append(errs, ErrorCampo{Campo: "ciudadManual", Mensaje: "Ingrese la ciudad."})
			} else if len(d.CiudadManual) > 120 {
				errs = append(errs, ErrorCampo{Campo: "ciudadManual", Mensaje: "Use como máximo 120 caracteres."})
			}
		}
	}

	cel := strings.TrimSpace(d.TelefonoCelular)
	if cel == "" {
		errs = append(errs, ErrorCampo{Campo: "telefonoCelular", Mensaje: "Ingrese su teléfono celular."})
	} else if !reTelefono.MatchString(cel) {
		errs = append(errs, ErrorCampo{Campo: "telefonoCelular", Mensaje: "Ingrese el teléfono con 9 o 10 números, por ejemplo 0991234567, o con código de país, por ejemplo +593991234567."})
	}

	correoAlt := strings.TrimSpace(d.CorreoAlterno)
	if correoAlt != "" {
		if !reEmail.MatchString(correoAlt) {
			errs = append(errs, ErrorCampo{Campo: "correoAlterno", Mensaje: "Ingrese un correo válido, por ejemplo nombre@dominio.com."})
		} else if ctx.CorreoPrincipal != "" && strings.EqualFold(correoAlt, ctx.CorreoPrincipal) {
			errs = append(errs, ErrorCampo{Campo: "correoAlterno", Mensaje: "El correo secundario debe ser distinto del correo principal."})
		}
	}

	cp := strings.TrimSpace(d.CallePrincipal)
	if cp == "" {
		errs = append(errs, ErrorCampo{Campo: "callePrincipal", Mensaje: "Ingrese la calle principal."})
	} else if len(cp) > 150 {
		errs = append(errs, ErrorCampo{Campo: "callePrincipal", Mensaje: "Use como máximo 150 caracteres."})
	}

	num := strings.TrimSpace(d.Numero)
	if num == "" {
		errs = append(errs, ErrorCampo{Campo: "numero", Mensaje: "Ingrese el número de casa o edificio."})
	} else if len(num) > 20 {
		errs = append(errs, ErrorCampo{Campo: "numero", Mensaje: "Use como máximo 20 caracteres."})
	}

	if len(d.CalleSecundaria) > 150 {
		errs = append(errs, ErrorCampo{Campo: "calleSecundaria", Mensaje: "Use como máximo 150 caracteres."})
	}
	if len(d.Sector) > 120 {
		errs = append(errs, ErrorCampo{Campo: "sector", Mensaje: "Use como máximo 120 caracteres."})
	}
	if len(d.Referencia) > 255 {
		errs = append(errs, ErrorCampo{Campo: "referencia", Mensaje: "Use como máximo 255 caracteres."})
	}

	telDom := strings.TrimSpace(d.TelefonoDomicilio)
	if telDom != "" && !reTelefono.MatchString(telDom) {
		errs = append(errs, ErrorCampo{Campo: "telefonoDomicilio", Mensaje: "Ingrese el teléfono con 9 o 10 números, por ejemplo 0991234567, o con código de país, por ejemplo +593991234567."})
	}

	return errs
}

func validarFamiliares(datos []byte) []ErrorCampo {
	var d struct {
		TieneFamiliares string `json:"tieneFamiliares"`
		Familiares      []struct {
			NombresApellidos string `json:"nombresApellidos"`
			Parentesco       string `json:"parentesco"`
			Institucion      string `json:"institucion"`
			Area             string `json:"area"`
			Cargo            string `json:"cargo"`
		} `json:"familiares"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	tf := strings.TrimSpace(d.TieneFamiliares)
	if tf != "SI" && tf != "NO" {
		errs = append(errs, ErrorCampo{Campo: "tieneFamiliares", Mensaje: "Indique si tiene familiares que trabajen en el IESS o el BIESS."})
	}

	if tf == "SI" {
		if len(d.Familiares) == 0 {
			errs = append(errs, ErrorCampo{Campo: "familiares", Mensaje: "Agregue al menos un familiar o responda “No”."})
		}
		parentescos := map[string]bool{
			"CONVIVIENTE": true, "CONYUGE": true, "CUNADO": true, "HERMANO": true,
			"HIJO_POLITICO": true, "HIJO": true, "MADRE": true, "NIETO": true,
			"NUERA": true, "PADRE": true, "PRIMO_HERMANO": true, "SOBRINO": true,
			"SUEGRO": true, "TIO": true, "YERNO": true,
		}
		for i, f := range d.Familiares {
			if strings.TrimSpace(f.NombresApellidos) == "" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("familiares.%d.nombresApellidos", i), Mensaje: "Ingrese los nombres y apellidos del familiar."})
			}
			p := strings.TrimSpace(f.Parentesco)
			if p == "" || !parentescos[p] {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("familiares.%d.parentesco", i), Mensaje: "Seleccione el parentesco."})
			}
			inst := strings.TrimSpace(f.Institucion)
			if inst != "IESS" && inst != "BIESS" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("familiares.%d.institucion", i), Mensaje: "Seleccione la institución."})
			}
			if strings.TrimSpace(f.Area) == "" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("familiares.%d.area", i), Mensaje: "Ingrese el área en la que trabaja."})
			}
			if strings.TrimSpace(f.Cargo) == "" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("familiares.%d.cargo", i), Mensaje: "Ingrese el cargo que ocupa."})
			}
		}
	}
	return errs
}

func validarCargo(datos []byte) []ErrorCampo {
	var d struct {
		TieneRespaldo      string          `json:"tieneRespaldo"`
		NombreOrganizacion string          `json:"nombreOrganizacion"`
		Documento          json.RawMessage `json:"documento"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	tr := strings.TrimSpace(d.TieneRespaldo)
	if tr != "SI" && tr != "NO" {
		errs = append(errs, ErrorCampo{Campo: "tieneRespaldo", Mensaje: "Indique si cuenta con respaldo organizacional."})
	}
	if tr == "SI" {
		if strings.TrimSpace(d.NombreOrganizacion) == "" {
			errs = append(errs, ErrorCampo{Campo: "nombreOrganizacion", Mensaje: "Ingrese el nombre de la organización."})
		} else if len(d.NombreOrganizacion) > 200 {
			errs = append(errs, ErrorCampo{Campo: "nombreOrganizacion", Mensaje: "Use como máximo 200 caracteres."})
		}
		if !esDocumentoValido(d.Documento) {
			errs = append(errs, ErrorCampo{Campo: "documento", Mensaje: "Adjunte el documento de respaldo en PDF."})
		}
	}
	return errs
}

func validarFormacion(datos []byte) []ErrorCampo {
	var d struct {
		Formaciones []struct {
			Nivel           string          `json:"nivel"`
			Institucion     string          `json:"institucion"`
			TipoInstitucion string          `json:"tipoInstitucion"`
			Titulo          string          `json:"titulo"`
			Area            string          `json:"area"`
			NumeroSenescyt  string          `json:"numeroSenescyt"`
			FechaSenescyt   string          `json:"fechaSenescyt"`
			Documento       json.RawMessage `json:"documento"`
		} `json:"formaciones"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	if len(d.Formaciones) == 0 {
		return []ErrorCampo{{Campo: "formaciones", Mensaje: "Agregue al menos una formación académica."}}
	}

	niveles := map[string]bool{
		"TERCER_NIVEL": true, "ESPECIALIDAD": true, "MAESTRIA": true, "PHD": true,
	}

	for i, f := range d.Formaciones {
		if !niveles[strings.TrimSpace(f.Nivel)] {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.nivel", i), Mensaje: "Seleccione el nivel de instrucción."})
		}
		if strings.TrimSpace(f.Institucion) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.institucion", i), Mensaje: "Ingrese la institución educativa."})
		}
		ti := strings.TrimSpace(f.TipoInstitucion)
		if ti != "NACIONAL" && ti != "EXTRANJERA" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.tipoInstitucion", i), Mensaje: "Seleccione si la institución es nacional o extranjera."})
		}
		if strings.TrimSpace(f.Titulo) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.titulo", i), Mensaje: "Ingrese el título obtenido."})
		}
		if strings.TrimSpace(f.Area) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.area", i), Mensaje: "Ingrese el área de estudios."})
		}
		if strings.TrimSpace(f.NumeroSenescyt) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.numeroSenescyt", i), Mensaje: "Ingrese el número de registro de la SENESCYT."})
		}

		if strings.TrimSpace(f.FechaSenescyt) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.fechaSenescyt", i), Mensaje: "Ingrese la fecha de registro de la SENESCYT."})
		} else if t, err := parsearFecha(f.FechaSenescyt); err != nil {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.fechaSenescyt", i), Mensaje: "Ingrese la fecha con el formato dd/mm/aaaa, por ejemplo 15/01/2020."})
		} else if esFechaFutura(t) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.fechaSenescyt", i), Mensaje: "La fecha no puede ser posterior a hoy."})
		}

		if !esDocumentoValido(f.Documento) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("formaciones.%d.documento", i), Mensaje: "Adjunte el título en PDF."})
		}
	}
	return errs
}

func validarExperiencia(datos []byte) []ErrorCampo {
	var d struct {
		Experiencias []struct {
			Tipo        string          `json:"tipo"`
			Cargo       string          `json:"cargo"`
			Institucion string          `json:"institucion"`
			FechaInicio string          `json:"fechaInicio"`
			Actualmente bool            `json:"actualmente"`
			FechaFin    string          `json:"fechaFin"`
			Actividades string          `json:"actividades"`
			Funciones   string          `json:"funciones"`
			Documento   json.RawMessage `json:"documento"`
		} `json:"experiencias"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	for i, exp := range d.Experiencias {
		t := strings.TrimSpace(exp.Tipo)
		if t != "GENERAL" && t != "DIRECCION" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.tipo", i), Mensaje: "Seleccione el tipo de experiencia."})
		}
		if strings.TrimSpace(exp.Cargo) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.cargo", i), Mensaje: "Ingrese el cargo o puesto."})
		}
		if strings.TrimSpace(exp.Institucion) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.institucion", i), Mensaje: "Ingrese la institución o empresa."})
		}

		var tInicio time.Time
		var errInicio error
		if strings.TrimSpace(exp.FechaInicio) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.fechaInicio", i), Mensaje: "Ingrese la fecha de inicio."})
		} else if tInicio, errInicio = parsearFecha(exp.FechaInicio); errInicio != nil {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.fechaInicio", i), Mensaje: "Ingrese la fecha con el formato dd/mm/aaaa, por ejemplo 15/01/2020."})
		} else if esFechaFutura(tInicio) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.fechaInicio", i), Mensaje: "La fecha no puede ser posterior a hoy."})
		}

		if !exp.Actualmente {
			if strings.TrimSpace(exp.FechaFin) == "" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.fechaFin", i), Mensaje: "Ingrese la fecha de fin o marque “Trabajo aquí actualmente”."})
			} else if tFin, errFin := parsearFecha(exp.FechaFin); errFin != nil {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.fechaFin", i), Mensaje: "Ingrese la fecha con el formato dd/mm/aaaa, por ejemplo 15/01/2020."})
			} else if esFechaFutura(tFin) {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.fechaFin", i), Mensaje: "La fecha no puede ser posterior a hoy."})
			} else if errInicio == nil && tFin.Before(tInicio) {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.fechaFin", i), Mensaje: "La fecha de fin debe ser posterior a la fecha de inicio."})
			}
		}

		if strings.TrimSpace(exp.Actividades) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.actividades", i), Mensaje: "Ingrese las actividades relacionadas con el cargo."})
		}
		if strings.TrimSpace(exp.Funciones) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.funciones", i), Mensaje: "Ingrese la descripción de funciones."})
		}
		if !esDocumentoValido(exp.Documento) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("experiencias.%d.documento", i), Mensaje: "Adjunte el certificado de experiencia en PDF."})
		}
	}
	return errs
}

func validarPublicaciones(datos []byte) []ErrorCampo {
	var d struct {
		TienePublicaciones string `json:"tienePublicaciones"`
		Publicaciones      []struct {
			Titulo    string          `json:"titulo"`
			Tipo      string          `json:"tipo"`
			Medio     string          `json:"medio"`
			Fecha     string          `json:"fecha"`
			Relacion  string          `json:"relacion"`
			Documento json.RawMessage `json:"documento"`
		} `json:"publicaciones"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	tp := strings.TrimSpace(d.TienePublicaciones)
	if tp != "SI" && tp != "NO" {
		errs = append(errs, ErrorCampo{Campo: "tienePublicaciones", Mensaje: "Indique si cuenta con publicaciones o investigaciones académicas."})
	}

	if tp == "SI" {
		if len(d.Publicaciones) == 0 {
			errs = append(errs, ErrorCampo{Campo: "publicaciones", Mensaje: "Agregue al menos una publicación o responda “No”."})
		}
		for i, pub := range d.Publicaciones {
			if strings.TrimSpace(pub.Titulo) == "" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("publicaciones.%d.titulo", i), Mensaje: "Ingrese el título de la publicación."})
			}
			t := strings.TrimSpace(pub.Tipo)
			if t != "PUBLICACION" && t != "TRABAJO_INVESTIGACION" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("publicaciones.%d.tipo", i), Mensaje: "Seleccione el tipo de publicación."})
			}
			if strings.TrimSpace(pub.Medio) == "" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("publicaciones.%d.medio", i), Mensaje: "Ingrese el medio de publicación."})
			}
			if strings.TrimSpace(pub.Fecha) == "" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("publicaciones.%d.fecha", i), Mensaje: "Ingrese la fecha de publicación."})
			} else if tf, err := parsearFecha(pub.Fecha); err != nil {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("publicaciones.%d.fecha", i), Mensaje: "Ingrese la fecha con el formato dd/mm/aaaa, por ejemplo 15/01/2020."})
			} else if esFechaFutura(tf) {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("publicaciones.%d.fecha", i), Mensaje: "La fecha no puede ser posterior a hoy."})
			}
			if strings.TrimSpace(pub.Relacion) == "" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("publicaciones.%d.relacion", i), Mensaje: "Indique la relación con su formación o ejercicio profesional."})
			}
			if !esDocumentoValido(pub.Documento) {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("publicaciones.%d.documento", i), Mensaje: "Adjunte la publicación en PDF."})
			}
		}
	}
	return errs
}

func validarCapacitacion(datos []byte) []ErrorCampo {
	var d struct {
		Capacitaciones []struct {
			TipoEvento  string          `json:"tipoEvento"`
			Nombre      string          `json:"nombre"`
			Institucion string          `json:"institucion"`
			FechaInicio string          `json:"fechaInicio"`
			FechaFin    string          `json:"fechaFin"`
			Horas       string          `json:"horas"`
			Documento   json.RawMessage `json:"documento"`
		} `json:"capacitaciones"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	tipos := map[string]bool{"CURSO": true, "SEMINARIO": true, "TALLER": true}

	for i, c := range d.Capacitaciones {
		if !tipos[strings.TrimSpace(c.TipoEvento)] {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.tipoEvento", i), Mensaje: "Seleccione el tipo de evento."})
		}
		if strings.TrimSpace(c.Nombre) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.nombre", i), Mensaje: "Ingrese el nombre del evento."})
		}
		if strings.TrimSpace(c.Institucion) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.institucion", i), Mensaje: "Ingrese la institución que organizó el evento."})
		}

		var tInicio time.Time
		var errInicio error
		if strings.TrimSpace(c.FechaInicio) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.fechaInicio", i), Mensaje: "Ingrese la fecha de inicio."})
		} else if tInicio, errInicio = parsearFecha(c.FechaInicio); errInicio != nil {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.fechaInicio", i), Mensaje: "Ingrese la fecha con el formato dd/mm/aaaa, por ejemplo 15/01/2020."})
		} else if esFechaFutura(tInicio) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.fechaInicio", i), Mensaje: "La fecha no puede ser posterior a hoy."})
		}

		if strings.TrimSpace(c.FechaFin) == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.fechaFin", i), Mensaje: "Ingrese la fecha de fin."})
		} else if tFin, errFin := parsearFecha(c.FechaFin); errFin != nil {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.fechaFin", i), Mensaje: "Ingrese la fecha con el formato dd/mm/aaaa, por ejemplo 15/01/2020."})
		} else if esFechaFutura(tFin) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.fechaFin", i), Mensaje: "La fecha no puede ser posterior a hoy."})
		} else if errInicio == nil && tFin.Before(tInicio) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.fechaFin", i), Mensaje: "La fecha de fin debe ser posterior a la fecha de inicio."})
		}

		h := strings.TrimSpace(c.Horas)
		if h == "" {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.horas", i), Mensaje: "Ingrese el número de horas."})
		} else if !reHoras.MatchString(h) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.horas", i), Mensaje: "Ingrese un número de horas mayor que cero, sin decimales."})
		}

		if !esDocumentoValido(c.Documento) {
			errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("capacitaciones.%d.documento", i), Mensaje: "Adjunte el certificado en PDF."})
		}
	}
	return errs
}

func validarInhabilidades(datos []byte) []ErrorCampo {
	var d struct {
		Respuestas          []string `json:"respuestas"`
		RemocionInstitucion string   `json:"remocionInstitucion"`
		RemocionCargo       string   `json:"remocionCargo"`
		RemocionFechaFin    string   `json:"remocionFechaFin"`
		RemocionMotivo      string   `json:"remocionMotivo"`
	}
	_ = json.Unmarshal(datos, &d)

	var errs []ErrorCampo
	if len(d.Respuestas) != 13 {
		errs = append(errs, ErrorCampo{Campo: "respuestas", Mensaje: "Responda todas las preguntas de inhabilidades."})
	} else {
		for i, r := range d.Respuestas {
			resp := strings.TrimSpace(r)
			if resp != "SI" && resp != "NO" {
				errs = append(errs, ErrorCampo{Campo: fmt.Sprintf("respuestas.%d", i), Mensaje: "Responda esta pregunta."})
			}
		}

		// Pregunta 7 (índice 6: remoción por organismo de control)
		if d.Respuestas[6] == "SI" {
			if strings.TrimSpace(d.RemocionInstitucion) == "" {
				errs = append(errs, ErrorCampo{Campo: "remocionInstitucion", Mensaje: "Ingrese la institución."})
			}
			if strings.TrimSpace(d.RemocionCargo) == "" {
				errs = append(errs, ErrorCampo{Campo: "remocionCargo", Mensaje: "Ingrese el cargo."})
			}

			if strings.TrimSpace(d.RemocionFechaFin) == "" {
				errs = append(errs, ErrorCampo{Campo: "remocionFechaFin", Mensaje: "Ingrese la fecha de finalización."})
			} else if tFin, err := parsearFecha(d.RemocionFechaFin); err != nil {
				errs = append(errs, ErrorCampo{Campo: "remocionFechaFin", Mensaje: "Ingrese la fecha con el formato dd/mm/aaaa, por ejemplo 15/01/2020."})
			} else if esFechaFutura(tFin) {
				errs = append(errs, ErrorCampo{Campo: "remocionFechaFin", Mensaje: "La fecha no puede ser posterior a hoy."})
			}

			if strings.TrimSpace(d.RemocionMotivo) == "" {
				errs = append(errs, ErrorCampo{Campo: "remocionMotivo", Mensaje: "Ingrese el motivo de la remoción."})
			}
		}
	}
	return errs
}

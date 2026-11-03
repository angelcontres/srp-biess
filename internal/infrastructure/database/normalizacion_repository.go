package database

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"srp-biess/internal/domain"
)

// PostgresNormalizador implementa domain.NormalizadorPostulacion (SRP-017e, decisión del #42):
// el JSONB de `postulacion_seccion` es la fuente durante el borrador y el envío lo vuelca a las
// tablas de SDD 4.6. Usa la transacción del contexto (TxManager) si existe.
//
// Reglas del volcado:
//   - Secciones repetibles: las filas reflejan exactamente la lista guardada. El id de cada
//     ítem se conserva entre envíos para no romper las referencias a documentos.
//   - Secciones de una sola fila (domicilio, pareja, respaldo): se reemplazan; si la sección
//     nunca se guardó, no se tocan.
//   - La sección `cargo` contiene el respaldo organizacional; `postulacion.cargo_id` lo escribe
//     PUT /cargo (SRP-017a). De `personal` solo se vuelca la pareja: el resto ya está en `postulante`.
type PostgresNormalizador struct {
	q   Querier
	hoy func() time.Time
}

var _ domain.NormalizadorPostulacion = (*PostgresNormalizador)(nil)

// Ecuador no tiene horario de verano: UTC-5 fijo evita depender de tzdata en Windows.
var zonaEcuador = time.FixedZone("America/Guayaquil", -5*60*60)

func NewPostgresNormalizador(q Querier) *PostgresNormalizador {
	return &PostgresNormalizador{q: q, hoy: func() time.Time { return time.Now().In(zonaEcuador) }}
}

// ErrIDItemAjeno indica que el id de un ítem ya pertenece a otra postulación.
var ErrIDItemAjeno = errors.New("el identificador del ítem pertenece a otra postulación")

// Pregunta 1.9.7: la única que pide los datos de remoción (RF-EF12-16).
const codigoPreguntaRemocion = "1.9.7"

// documentoRef es la referencia a un PDF subido (UploadedFileInfo en la SPA).
type documentoRef struct {
	ID string `json:"id"`
}

type volcado struct {
	q             Querier
	postulacionID string
	postulanteID  string
	hoy           time.Time
}

func (n *PostgresNormalizador) Normalizar(ctx context.Context, postulacionID string) error {
	if !idsValidos(postulacionID) {
		return domain.ErrNotFound
	}
	v := &volcado{q: ExtractQuerier(ctx, n.q), postulacionID: postulacionID, hoy: n.hoy()}

	err := v.q.QueryRow(ctx,
		`SELECT postulante_id::text FROM postulacion WHERE id = $1::uuid FOR UPDATE`, postulacionID,
	).Scan(&v.postulanteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("error consultando postulación: %w", err)
	}

	secciones, err := v.leerSecciones(ctx)
	if err != nil {
		return err
	}

	pasos := []struct {
		seccion string
		volcar  func(context.Context, []byte) error
	}{
		{"personal", v.pareja},
		{"domicilio", v.domicilio},
		{"familiares", v.familiares},
		{"cargo", v.respaldo},
		{"formacion", v.formacion},
		{"experiencia", v.experiencia},
		{"publicaciones", v.publicaciones},
		{"capacitacion", v.capacitacion},
		{"inhabilidades", v.inhabilidades},
	}
	for _, paso := range pasos {
		if err := paso.volcar(ctx, secciones[paso.seccion]); err != nil {
			return fmt.Errorf("normalizar sección %s: %w", paso.seccion, err)
		}
	}
	return nil
}

func (v *volcado) leerSecciones(ctx context.Context) (map[string][]byte, error) {
	rows, err := v.q.Query(ctx,
		`SELECT seccion, datos::text FROM postulacion_seccion WHERE postulacion_id = $1::uuid`, v.postulacionID)
	if err != nil {
		return nil, fmt.Errorf("error consultando secciones: %w", err)
	}
	defer rows.Close()
	secciones := map[string][]byte{}
	for rows.Next() {
		var seccion, datos string
		if err := rows.Scan(&seccion, &datos); err != nil {
			return nil, fmt.Errorf("error leyendo secciones: %w", err)
		}
		secciones[seccion] = []byte(datos)
	}
	return secciones, rows.Err()
}

// --- Secciones de una sola fila ---

func (v *volcado) pareja(ctx context.Context, datos []byte) error {
	if datos == nil {
		return nil
	}
	var s struct {
		EstadoCivil   string `json:"estadoCivil"`
		ParejaNombres string `json:"parejaNombres"`
		ParejaCedula  string `json:"parejaCedula"`
	}
	if err := json.Unmarshal(datos, &s); err != nil {
		return err
	}
	// RF-EF12-05: solo casado o unión de hecho registran pareja.
	if s.EstadoCivil != "CASADO" && s.EstadoCivil != "UNION_DE_HECHO" {
		_, err := v.q.Exec(ctx, `DELETE FROM pareja_postulante WHERE postulante_id = $1::uuid`, v.postulanteID)
		return err
	}
	_, err := v.q.Exec(ctx, `
		INSERT INTO pareja_postulante (postulante_id, nombres_apellidos, cedula)
		VALUES ($1::uuid, $2, $3)
		ON CONFLICT (postulante_id) DO UPDATE
		SET nombres_apellidos = EXCLUDED.nombres_apellidos, cedula = EXCLUDED.cedula, updated_at = now()`,
		v.postulanteID, strings.TrimSpace(s.ParejaNombres), strings.TrimSpace(s.ParejaCedula))
	return err
}

func (v *volcado) domicilio(ctx context.Context, datos []byte) error {
	if datos == nil {
		return nil
	}
	var s struct {
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
	if err := json.Unmarshal(datos, &s); err != nil {
		return err
	}
	// La validación país/provincia/cantón la hace el trigger fn_domicilio_valida_ubicacion.
	_, err := v.q.Exec(ctx, `
		INSERT INTO domicilio_postulante (
			postulante_id, pais_id, provincia_id, canton_id, provincia_manual, ciudad_manual,
			telefono_celular, correo_alterno, calle_principal, numero, calle_secundaria,
			sector, referencia, telefono_domicilio)
		VALUES ($1::uuid, NULLIF($2, '')::smallint, NULLIF($3, '')::smallint, NULLIF($4, '')::smallint,
			NULLIF($5, ''), NULLIF($6, ''), $7, NULLIF($8, ''), $9, $10, $11, $12, $13, NULLIF($14, ''))
		ON CONFLICT (postulante_id) DO UPDATE SET
			pais_id = EXCLUDED.pais_id, provincia_id = EXCLUDED.provincia_id, canton_id = EXCLUDED.canton_id,
			provincia_manual = EXCLUDED.provincia_manual, ciudad_manual = EXCLUDED.ciudad_manual,
			telefono_celular = EXCLUDED.telefono_celular, correo_alterno = EXCLUDED.correo_alterno,
			calle_principal = EXCLUDED.calle_principal, numero = EXCLUDED.numero,
			calle_secundaria = EXCLUDED.calle_secundaria, sector = EXCLUDED.sector,
			referencia = EXCLUDED.referencia, telefono_domicilio = EXCLUDED.telefono_domicilio,
			updated_at = now()`,
		v.postulanteID, s.PaisID, s.ProvinciaID, s.CantonID,
		strings.TrimSpace(s.ProvinciaManual), strings.TrimSpace(s.CiudadManual),
		strings.TrimSpace(s.TelefonoCelular), strings.TrimSpace(s.CorreoAlterno),
		strings.TrimSpace(s.CallePrincipal), strings.TrimSpace(s.Numero),
		strings.TrimSpace(s.CalleSecundaria), strings.TrimSpace(s.Sector),
		strings.TrimSpace(s.Referencia), strings.TrimSpace(s.TelefonoDomicilio))
	return err
}

func (v *volcado) respaldo(ctx context.Context, datos []byte) error {
	if datos == nil {
		return nil
	}
	var s struct {
		TieneRespaldo      string        `json:"tieneRespaldo"`
		NombreOrganizacion string        `json:"nombreOrganizacion"`
		Documento          *documentoRef `json:"documento"`
	}
	if err := json.Unmarshal(datos, &s); err != nil {
		return err
	}
	tiene, err := siNo(s.TieneRespaldo)
	if err != nil {
		return err
	}
	nombre, doc := strings.TrimSpace(s.NombreOrganizacion), s.Documento
	if !tiene {
		nombre, doc = "", nil
	}
	_, err = v.q.Exec(ctx, `
		INSERT INTO respaldo_organizacional (postulacion_id, tiene_respaldo, nombre_organizacion, documento_id)
		VALUES ($1::uuid, $2, NULLIF($3, ''), `+sqlDocumento(4, 5)+`)
		ON CONFLICT (postulacion_id) DO UPDATE SET
			tiene_respaldo = EXCLUDED.tiene_respaldo, nombre_organizacion = EXCLUDED.nombre_organizacion,
			documento_id = EXCLUDED.documento_id, updated_at = now()`,
		v.postulacionID, tiene, nombre, documentoID(doc), v.postulanteID)
	return err
}

// --- Secciones repetibles ---

func (v *volcado) familiares(ctx context.Context, datos []byte) error {
	var s struct {
		TieneFamiliares string `json:"tieneFamiliares"`
		Familiares      []struct {
			ID               string `json:"id"`
			NombresApellidos string `json:"nombresApellidos"`
			Parentesco       string `json:"parentesco"`
			Institucion      string `json:"institucion"`
			Area             string `json:"area"`
			Cargo            string `json:"cargo"`
		} `json:"familiares"`
	}
	if err := unmarshalOpcional(datos, &s); err != nil {
		return err
	}
	if s.TieneFamiliares != "SI" {
		s.Familiares = nil
	}
	if err := v.marcarRespuesta(ctx, "tiene_familiares_institucion", s.TieneFamiliares); err != nil {
		return err
	}
	// familiar_institucion cuelga del postulante (SDD 4.6), no de la postulación.
	ids := make([]string, 0, len(s.Familiares))
	for _, f := range s.Familiares {
		id := v.idItem("familiares", f.ID)
		ids = append(ids, id)
		if err := upsertPropio(ctx, v.q, `
			INSERT INTO familiar_institucion (id, postulante_id, nombres_apellidos, parentesco, institucion, area, cargo)
			VALUES ($1::uuid, $2::uuid, $3, $4::parentesco_familiar, $5::institucion_familiar, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				nombres_apellidos = EXCLUDED.nombres_apellidos, parentesco = EXCLUDED.parentesco,
				institucion = EXCLUDED.institucion, area = EXCLUDED.area, cargo = EXCLUDED.cargo, updated_at = now()
			WHERE familiar_institucion.postulante_id = EXCLUDED.postulante_id`,
			id, v.postulanteID, strings.TrimSpace(f.NombresApellidos), f.Parentesco, f.Institucion,
			strings.TrimSpace(f.Area), strings.TrimSpace(f.Cargo)); err != nil {
			return err
		}
	}
	return borrarSobrantes(ctx, v.q, "familiar_institucion", "postulante_id", v.postulanteID, ids)
}

func (v *volcado) formacion(ctx context.Context, datos []byte) error {
	var s struct {
		Formaciones []struct {
			ID              string        `json:"id"`
			Nivel           string        `json:"nivel"`
			Institucion     string        `json:"institucion"`
			TipoInstitucion string        `json:"tipoInstitucion"`
			Titulo          string        `json:"titulo"`
			Area            string        `json:"area"`
			NumeroSenescyt  string        `json:"numeroSenescyt"`
			FechaSenescyt   string        `json:"fechaSenescyt"`
			Documento       *documentoRef `json:"documento"`
		} `json:"formaciones"`
	}
	if err := unmarshalOpcional(datos, &s); err != nil {
		return err
	}
	ids := make([]string, 0, len(s.Formaciones))
	for _, f := range s.Formaciones {
		id := v.idItem("formacion", f.ID)
		ids = append(ids, id)
		if err := upsertPropio(ctx, v.q, `
			INSERT INTO formacion_academica (
				id, postulacion_id, nivel_instruccion, institucion_educativa, tipo_institucion, titulo_obtenido,
				area_estudios, numero_registro_senescyt, fecha_registro_senescyt, documento_id)
			VALUES ($1::uuid, $2::uuid, $3::nivel_instruccion, $4, $5::tipo_institucion, $6, $7,
				NULLIF($8, ''), NULLIF($9, '')::date, `+sqlDocumento(10, 11)+`)
			ON CONFLICT (id) DO UPDATE SET
				nivel_instruccion = EXCLUDED.nivel_instruccion, institucion_educativa = EXCLUDED.institucion_educativa,
				tipo_institucion = EXCLUDED.tipo_institucion, titulo_obtenido = EXCLUDED.titulo_obtenido,
				area_estudios = EXCLUDED.area_estudios, numero_registro_senescyt = EXCLUDED.numero_registro_senescyt,
				fecha_registro_senescyt = EXCLUDED.fecha_registro_senescyt, documento_id = EXCLUDED.documento_id,
				updated_at = now()
			WHERE formacion_academica.postulacion_id = EXCLUDED.postulacion_id`,
			id, v.postulacionID, f.Nivel, strings.TrimSpace(f.Institucion), f.TipoInstitucion,
			strings.TrimSpace(f.Titulo), strings.TrimSpace(f.Area), strings.TrimSpace(f.NumeroSenescyt),
			f.FechaSenescyt, documentoID(f.Documento), v.postulanteID); err != nil {
			return err
		}
	}
	return borrarSobrantes(ctx, v.q, "formacion_academica", "postulacion_id", v.postulacionID, ids)
}

func (v *volcado) experiencia(ctx context.Context, datos []byte) error {
	var s struct {
		Experiencias []struct {
			ID          string        `json:"id"`
			Tipo        string        `json:"tipo"`
			Cargo       string        `json:"cargo"`
			Institucion string        `json:"institucion"`
			FechaInicio string        `json:"fechaInicio"`
			Actualmente bool          `json:"actualmente"`
			FechaFin    string        `json:"fechaFin"`
			Actividades string        `json:"actividades"`
			Funciones   string        `json:"funciones"`
			Documento   *documentoRef `json:"documento"`
		} `json:"experiencias"`
	}
	if err := unmarshalOpcional(datos, &s); err != nil {
		return err
	}
	ids := make([]string, 0, len(s.Experiencias))
	for _, e := range s.Experiencias {
		id := v.idItem("experiencia", e.ID)
		ids = append(ids, id)

		inicio, err := time.Parse(time.DateOnly, e.FechaInicio)
		if err != nil {
			return fmt.Errorf("fecha de inicio inválida %q", e.FechaInicio)
		}
		fin, fechaFin := v.hoy, ""
		if !e.Actualmente {
			if fin, err = time.Parse(time.DateOnly, e.FechaFin); err != nil {
				return fmt.Errorf("fecha de fin inválida %q", e.FechaFin)
			}
			fechaFin = e.FechaFin
		}
		tiempo, ok := domain.CalcularTiempoExperiencia(inicio, fin)
		if !ok {
			return fmt.Errorf("la fecha de fin %s es anterior al inicio %s", fin.Format(time.DateOnly), e.FechaInicio)
		}

		if err := upsertPropio(ctx, v.q, `
			INSERT INTO experiencia_profesional (
				id, postulacion_id, tipo_experiencia, cargo_puesto, institucion_empresa, fecha_inicio, fecha_fin,
				actualmente, actividades_relacionadas, descripcion_funciones, tiempo_anios, tiempo_meses,
				tiempo_dias, documento_id)
			VALUES ($1::uuid, $2::uuid, $3::tipo_experiencia, $4, $5, $6::date, NULLIF($7, '')::date,
				$8, NULLIF($9, ''), $10, $11, $12, $13, `+sqlDocumento(14, 15)+`)
			ON CONFLICT (id) DO UPDATE SET
				tipo_experiencia = EXCLUDED.tipo_experiencia, cargo_puesto = EXCLUDED.cargo_puesto,
				institucion_empresa = EXCLUDED.institucion_empresa, fecha_inicio = EXCLUDED.fecha_inicio,
				fecha_fin = EXCLUDED.fecha_fin, actualmente = EXCLUDED.actualmente,
				actividades_relacionadas = EXCLUDED.actividades_relacionadas,
				descripcion_funciones = EXCLUDED.descripcion_funciones, tiempo_anios = EXCLUDED.tiempo_anios,
				tiempo_meses = EXCLUDED.tiempo_meses, tiempo_dias = EXCLUDED.tiempo_dias,
				documento_id = EXCLUDED.documento_id, updated_at = now()
			WHERE experiencia_profesional.postulacion_id = EXCLUDED.postulacion_id`,
			id, v.postulacionID, e.Tipo, strings.TrimSpace(e.Cargo), strings.TrimSpace(e.Institucion),
			e.FechaInicio, fechaFin, e.Actualmente, strings.TrimSpace(e.Actividades), strings.TrimSpace(e.Funciones),
			tiempo.Anios, tiempo.Meses, tiempo.Dias, documentoID(e.Documento), v.postulanteID); err != nil {
			return err
		}
	}
	return borrarSobrantes(ctx, v.q, "experiencia_profesional", "postulacion_id", v.postulacionID, ids)
}

func (v *volcado) publicaciones(ctx context.Context, datos []byte) error {
	var s struct {
		TienePublicaciones string `json:"tienePublicaciones"`
		Publicaciones      []struct {
			ID        string        `json:"id"`
			Titulo    string        `json:"titulo"`
			Tipo      string        `json:"tipo"`
			Medio     string        `json:"medio"`
			Fecha     string        `json:"fecha"`
			Relacion  string        `json:"relacion"`
			Documento *documentoRef `json:"documento"`
		} `json:"publicaciones"`
	}
	if err := unmarshalOpcional(datos, &s); err != nil {
		return err
	}
	if s.TienePublicaciones != "SI" {
		s.Publicaciones = nil
	}
	if err := v.marcarRespuesta(ctx, "tiene_publicaciones", s.TienePublicaciones); err != nil {
		return err
	}
	ids := make([]string, 0, len(s.Publicaciones))
	for _, p := range s.Publicaciones {
		id := v.idItem("publicaciones", p.ID)
		ids = append(ids, id)
		// La SPA pide un texto de relación y SDD 4.6 guarda un booleano: se registra si hay relación declarada.
		if err := upsertPropio(ctx, v.q, `
			INSERT INTO publicacion (
				id, postulacion_id, titulo, tipo, medio_publicacion, fecha_publicacion, relacion_formacion, documento_id)
			VALUES ($1::uuid, $2::uuid, $3, $4::tipo_publicacion, $5, $6::date, $7, `+sqlDocumento(8, 9)+`)
			ON CONFLICT (id) DO UPDATE SET
				titulo = EXCLUDED.titulo, tipo = EXCLUDED.tipo, medio_publicacion = EXCLUDED.medio_publicacion,
				fecha_publicacion = EXCLUDED.fecha_publicacion, relacion_formacion = EXCLUDED.relacion_formacion,
				documento_id = EXCLUDED.documento_id, updated_at = now()
			WHERE publicacion.postulacion_id = EXCLUDED.postulacion_id`,
			id, v.postulacionID, strings.TrimSpace(p.Titulo), p.Tipo, strings.TrimSpace(p.Medio), p.Fecha,
			strings.TrimSpace(p.Relacion) != "", documentoID(p.Documento), v.postulanteID); err != nil {
			return err
		}
	}
	return borrarSobrantes(ctx, v.q, "publicacion", "postulacion_id", v.postulacionID, ids)
}

func (v *volcado) capacitacion(ctx context.Context, datos []byte) error {
	var s struct {
		Capacitaciones []struct {
			ID          string        `json:"id"`
			TipoEvento  string        `json:"tipoEvento"`
			Nombre      string        `json:"nombre"`
			Institucion string        `json:"institucion"`
			FechaInicio string        `json:"fechaInicio"`
			FechaFin    string        `json:"fechaFin"`
			Horas       string        `json:"horas"`
			Documento   *documentoRef `json:"documento"`
		} `json:"capacitaciones"`
	}
	if err := unmarshalOpcional(datos, &s); err != nil {
		return err
	}
	ids := make([]string, 0, len(s.Capacitaciones))
	for _, c := range s.Capacitaciones {
		id := v.idItem("capacitacion", c.ID)
		ids = append(ids, id)
		horas, err := strconv.Atoi(strings.TrimSpace(c.Horas))
		if err != nil || horas <= 0 {
			return fmt.Errorf("número de horas inválido %q", c.Horas)
		}
		if err := upsertPropio(ctx, v.q, `
			INSERT INTO capacitacion (
				id, postulacion_id, tipo_evento, nombre_evento, institucion, fecha_inicio, fecha_fin, numero_horas,
				documento_id)
			VALUES ($1::uuid, $2::uuid, $3::tipo_evento, $4, $5, $6::date, $7::date, $8, `+sqlDocumento(9, 10)+`)
			ON CONFLICT (id) DO UPDATE SET
				tipo_evento = EXCLUDED.tipo_evento, nombre_evento = EXCLUDED.nombre_evento,
				institucion = EXCLUDED.institucion, fecha_inicio = EXCLUDED.fecha_inicio,
				fecha_fin = EXCLUDED.fecha_fin, numero_horas = EXCLUDED.numero_horas,
				documento_id = EXCLUDED.documento_id, updated_at = now()
			WHERE capacitacion.postulacion_id = EXCLUDED.postulacion_id`,
			id, v.postulacionID, c.TipoEvento, strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Institucion),
			c.FechaInicio, c.FechaFin, horas, documentoID(c.Documento), v.postulanteID); err != nil {
			return err
		}
	}
	return borrarSobrantes(ctx, v.q, "capacitacion", "postulacion_id", v.postulacionID, ids)
}

func (v *volcado) inhabilidades(ctx context.Context, datos []byte) error {
	var s struct {
		Respuestas          []string `json:"respuestas"`
		RemocionInstitucion string   `json:"remocionInstitucion"`
		RemocionCargo       string   `json:"remocionCargo"`
		RemocionFechaFin    string   `json:"remocionFechaFin"`
		RemocionMotivo      string   `json:"remocionMotivo"`
	}
	if err := unmarshalOpcional(datos, &s); err != nil {
		return err
	}

	// La SPA envía las respuestas en el orden del catálogo (inhabilidad_pregunta.orden).
	type pregunta struct {
		id                  int16
		codigo, efecto      string
		respuestaBloqueante *bool
	}
	rows, err := v.q.Query(ctx,
		`SELECT id, codigo, efecto::text, respuesta_bloqueante FROM inhabilidad_pregunta WHERE activo ORDER BY orden`)
	if err != nil {
		return fmt.Errorf("error consultando preguntas: %w", err)
	}
	preguntas, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (pregunta, error) {
		var p pregunta
		err := row.Scan(&p.id, &p.codigo, &p.efecto, &p.respuestaBloqueante)
		return p, err
	})
	if err != nil {
		return fmt.Errorf("error leyendo preguntas: %w", err)
	}
	if len(s.Respuestas) > len(preguntas) {
		return fmt.Errorf("se recibieron %d respuestas para %d preguntas", len(s.Respuestas), len(preguntas))
	}

	respondidas := make([]int32, 0, len(s.Respuestas))
	requiereRevision := false
	for i, texto := range s.Respuestas {
		if texto == "" {
			continue
		}
		respuesta, err := siNo(texto)
		if err != nil {
			return err
		}
		p := preguntas[i]
		efecto := domain.EfectoInhabilidad(p.efecto, p.respuestaBloqueante, respuesta)
		requiereRevision = requiereRevision || efecto == domain.EfectoAlerta

		var institucion, cargo, fechaFin, motivo string
		if p.codigo == codigoPreguntaRemocion && respuesta {
			institucion, cargo = strings.TrimSpace(s.RemocionInstitucion), strings.TrimSpace(s.RemocionCargo)
			fechaFin, motivo = s.RemocionFechaFin, strings.TrimSpace(s.RemocionMotivo)
		}
		if _, err := v.q.Exec(ctx, `
			INSERT INTO inhabilidad_respuesta (
				postulacion_id, pregunta_id, respuesta, efecto_calculado, institucion_remocion, cargo_remocion,
				fecha_fin_remocion, motivo_remocion)
			VALUES ($1::uuid, $2, $3, $4::efecto_inhabilidad, NULLIF($5, ''), NULLIF($6, ''),
				NULLIF($7, '')::date, NULLIF($8, ''))
			ON CONFLICT (postulacion_id, pregunta_id) DO UPDATE SET
				respuesta = EXCLUDED.respuesta, efecto_calculado = EXCLUDED.efecto_calculado,
				institucion_remocion = EXCLUDED.institucion_remocion, cargo_remocion = EXCLUDED.cargo_remocion,
				fecha_fin_remocion = EXCLUDED.fecha_fin_remocion, motivo_remocion = EXCLUDED.motivo_remocion,
				updated_at = now()`,
			v.postulacionID, p.id, respuesta, efecto, institucion, cargo, fechaFin, motivo); err != nil {
			return err
		}
		respondidas = append(respondidas, int32(p.id))
	}

	if _, err := v.q.Exec(ctx,
		`DELETE FROM inhabilidad_respuesta WHERE postulacion_id = $1::uuid AND NOT (pregunta_id = ANY($2::int[]))`,
		v.postulacionID, respondidas); err != nil {
		return err
	}
	// RF-EF12-16: una alerta permite continuar, pero marca la postulación para revisión.
	_, err = v.q.Exec(ctx,
		`UPDATE postulacion SET requiere_revision = $2, updated_at = now() WHERE id = $1::uuid`,
		v.postulacionID, requiereRevision)
	return err
}

// --- Utilidades ---

// idItem conserva el id del ítem si es un UUID; si no (ids temporales "tmp-..." de la SPA),
// deriva uno determinista para que el reenvío de la misma sección reutilice la fila.
func (v *volcado) idItem(seccion, id string) string {
	if idsValidos(id) {
		return strings.ToLower(id)
	}
	h := sha1.Sum([]byte(v.postulacionID + "/" + seccion + "/" + id))
	h[6] = (h[6] & 0x0f) | 0x50 // versión 5
	h[8] = (h[8] & 0x3f) | 0x80 // variante RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", h[0:4], h[4:6], h[6:8], h[8:10], h[10:16])
}

// marcarRespuesta guarda una pregunta Sí/No de cabecera en `postulacion` (NULL si no se respondió).
func (v *volcado) marcarRespuesta(ctx context.Context, columna, valor string) error {
	var respuesta *bool
	if valor != "" {
		b, err := siNo(valor)
		if err != nil {
			return err
		}
		respuesta = &b
	}
	_, err := v.q.Exec(ctx,
		fmt.Sprintf(`UPDATE postulacion SET %s = $2, updated_at = now() WHERE id = $1::uuid`, columna),
		v.postulacionID, respuesta)
	return err
}

// upsertPropio ejecuta un INSERT ... ON CONFLICT (id) DO UPDATE ... WHERE <mismo dueño>.
// Si el id ya existe con otro dueño no se escribe nada y se devuelve ErrIDItemAjeno.
func upsertPropio(ctx context.Context, q Querier, sql string, args ...any) error {
	tag, err := q.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ErrIDItemAjeno
	}
	return nil
}

// borrarSobrantes elimina las filas del dueño que ya no están en la lista guardada.
func borrarSobrantes(ctx context.Context, q Querier, tabla, columnaDueno, dueno string, ids []string) error {
	_, err := q.Exec(ctx,
		fmt.Sprintf(`DELETE FROM %s WHERE %s = $1::uuid AND NOT (id = ANY($2::uuid[]))`, tabla, columnaDueno),
		dueno, ids)
	return err
}

// sqlDocumento resuelve documento_id solo si el documento existe y es del postulante; si no, NULL.
func sqlDocumento(paramID, paramPostulante int) string {
	return fmt.Sprintf(`(SELECT d.id FROM documento d WHERE d.id = $%d::uuid AND d.postulante_id = $%d::uuid)`,
		paramID, paramPostulante)
}

func documentoID(doc *documentoRef) *string {
	if doc == nil || !idsValidos(doc.ID) {
		return nil
	}
	return &doc.ID
}

func siNo(valor string) (bool, error) {
	switch valor {
	case "SI":
		return true, nil
	case "NO":
		return false, nil
	default:
		return false, fmt.Errorf("respuesta Sí/No inválida %q", valor)
	}
}

// unmarshalOpcional trata una sección nunca guardada como vacía.
func unmarshalOpcional(datos []byte, destino any) error {
	if datos == nil {
		return nil
	}
	return json.Unmarshal(datos, destino)
}

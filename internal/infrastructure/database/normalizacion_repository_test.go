package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
)

const (
	idFormacionA   = "aaaaaaaa-0000-4000-8000-000000000001"
	idFormacionB   = "aaaaaaaa-0000-4000-8000-000000000002"
	idExperiencia  = "bbbbbbbb-0000-4000-8000-000000000001"
	idFamiliar     = "cccccccc-0000-4000-8000-000000000001"
	idPublicacion  = "dddddddd-0000-4000-8000-000000000001"
	idCapacitacion = "eeeeeeee-0000-4000-8000-000000000001"
)

type escenarioNormalizacion struct {
	pool          *pgxpool.Pool
	repo          *database.PostgresFormularioRepository
	postulanteID  string
	postulacionID string
	version       int
	documentoID   string
}

func nuevoEscenarioNormalizacion(ctx context.Context, t *testing.T) *escenarioNormalizacion {
	t.Helper()
	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	aplicarMigraciones(ctx, t, func(ctx context.Context, sql string) error {
		_, err := pool.Exec(ctx, sql)
		return err
	})

	sufijo := time.Now().UnixNano()
	postulante := &domain.Postulante{
		Cedula:       fmt.Sprintf("17%08d", sufijo%100000000),
		Nombres:      "PRUEBA",
		Apellidos:    "NORMALIZACION",
		Email:        fmt.Sprintf("normalizacion_%d@biess.gob.ec", sufijo),
		PasswordHash: "hash_seguro",
	}
	if err := database.NewPostgresPostulanteRepository(pool).Save(ctx, postulante); err != nil {
		t.Fatalf("no se pudo crear el postulante: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM postulacion WHERE postulante_id = $1::uuid`, postulante.ID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM postulante WHERE id = $1::uuid`, postulante.ID)
	})

	repo := database.NewPostgresFormularioRepository(pool)
	borrador, err := repo.CrearBorrador(ctx, postulante.ID)
	if errors.Is(err, domain.ErrFaseCerrada) {
		t.Skip("No hay convocatoria con la fase habilitada en la fecha actual")
	}
	if err != nil {
		t.Fatalf("CrearBorrador: %v", err)
	}

	var documentoID string
	err = pool.QueryRow(ctx, `
		INSERT INTO documento (postulante_id, postulacion_id, tipo, nombre_original, nombre_almacenado,
			mime_type, tamano_bytes, hash_sha256, ruta)
		VALUES ($1::uuid, $2::uuid, 'TITULO', 'titulo.pdf', $3, 'application/pdf', 1024, repeat('a', 64), '/tmp/t.pdf')
		RETURNING id::text`, postulante.ID, borrador.ID, fmt.Sprintf("titulo_%d.pdf", sufijo)).Scan(&documentoID)
	if err != nil {
		t.Fatalf("no se pudo crear el documento: %v", err)
	}

	return &escenarioNormalizacion{
		pool: pool, repo: repo, postulanteID: postulante.ID,
		postulacionID: borrador.ID, version: borrador.Version, documentoID: documentoID,
	}
}

func (e *escenarioNormalizacion) guardar(ctx context.Context, t *testing.T, seccion, datos string) {
	t.Helper()
	estado, err := e.repo.GuardarSeccion(ctx, e.postulacionID, e.postulanteID, seccion, []byte(datos), true, e.version)
	if err != nil {
		t.Fatalf("GuardarSeccion %s: %v", seccion, err)
	}
	e.version = estado.Version
}

func (e *escenarioNormalizacion) normalizar(ctx context.Context, t *testing.T) {
	t.Helper()
	normalizador := database.NewPostgresNormalizador(e.pool)
	err := database.NewPgxTxManager(e.pool).WithTransaction(ctx, func(txCtx context.Context) error {
		return normalizador.Normalizar(txCtx, e.postulacionID)
	})
	if err != nil {
		t.Fatalf("Normalizar: %v", err)
	}
}

func (e *escenarioNormalizacion) contar(ctx context.Context, t *testing.T, tabla, columna, valor string) int {
	t.Helper()
	var n int
	sql := fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s = $1::uuid`, tabla, columna)
	if err := e.pool.QueryRow(ctx, sql, valor).Scan(&n); err != nil {
		t.Fatalf("contar %s: %v", tabla, err)
	}
	return n
}

func ubicacionEcuador(ctx context.Context, t *testing.T, pool *pgxpool.Pool) (pais, provincia, canton int) {
	t.Helper()
	err := pool.QueryRow(ctx, `
		SELECT p.id, pr.id, c.id
		FROM pais p JOIN provincia pr ON pr.pais_id = p.id JOIN canton c ON c.provincia_id = pr.id
		WHERE p.codigo_iso2 = 'EC' ORDER BY pr.id, c.id LIMIT 1`).Scan(&pais, &provincia, &canton)
	if err != nil {
		t.Skipf("catálogo geográfico sin datos de Ecuador: %v", err)
	}
	return
}

// Criterios del #42: filas exactas tras el envío, reenvío sin duplicados y con ids estables,
// y sin cambios en lo que la SPA lee (GET /postulaciones/{id}).
func TestNormalizador_Integracion(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	e := nuevoEscenarioNormalizacion(ctx, t)
	pais, provincia, canton := ubicacionEcuador(ctx, t, e.pool)

	e.guardar(ctx, t, "personal", `{"estadoCivil":"CASADO","parejaNombres":"ANA PEREZ","parejaCedula":"1710034065"}`)
	e.guardar(ctx, t, "domicilio", fmt.Sprintf(`{"paisId":"%d","provinciaId":"%d","cantonId":"%d",
		"provinciaManual":"","ciudadManual":"","telefonoCelular":"0991234567","correoAlterno":"",
		"callePrincipal":"Av. Amazonas","numero":"N34-12","calleSecundaria":"","sector":"",
		"referencia":"","telefonoDomicilio":""}`, pais, provincia, canton))
	e.guardar(ctx, t, "familiares", `{"tieneFamiliares":"SI","familiares":[{"id":"`+idFamiliar+`",
		"nombresApellidos":"LUIS PEREZ","parentesco":"HERMANO","institucion":"IESS","area":"TI","cargo":"Analista"}]}`)
	e.guardar(ctx, t, "cargo", `{"tieneRespaldo":"SI","nombreOrganizacion":"Federación de Cámaras",
		"documento":{"id":"`+e.documentoID+`","name":"respaldo.pdf"}}`)
	e.guardar(ctx, t, "formacion", `{"formaciones":[
		{"id":"`+idFormacionA+`","nivel":"TERCER_NIVEL","institucion":"ESPOL","tipoInstitucion":"NACIONAL",
		 "titulo":"Ingeniero","area":"Computación","numeroSenescyt":"1006-2015-1","fechaSenescyt":"2015-03-01",
		 "documento":{"id":"`+e.documentoID+`"}},
		{"id":"`+idFormacionB+`","nivel":"MAESTRIA","institucion":"UPSE","tipoInstitucion":"NACIONAL",
		 "titulo":"Magíster","area":"Finanzas","numeroSenescyt":"1006-2019-2","fechaSenescyt":"2019-08-15",
		 "documento":{"id":"99999999-9999-4999-8999-999999999999"}}]}`)
	e.guardar(ctx, t, "experiencia", `{"experiencias":[
		{"id":"`+idExperiencia+`","tipo":"DIRECCION","cargo":"Gerente","institucion":"BIESS",
		 "fechaInicio":"2020-01-15","actualmente":false,"fechaFin":"2023-03-10",
		 "actividades":"Dirección","funciones":"Gestión del área","documento":null},
		{"id":"tmp-123","tipo":"GENERAL","cargo":"Analista","institucion":"IESS",
		 "fechaInicio":"2023-04-01","actualmente":true,"fechaFin":"",
		 "actividades":"","funciones":"Análisis","documento":null}]}`)
	e.guardar(ctx, t, "publicaciones", `{"tienePublicaciones":"SI","publicaciones":[{"id":"`+idPublicacion+`",
		"titulo":"Riesgo previsional","tipo":"PUBLICACION","medio":"Revista X","fecha":"2021-06-01",
		"relacion":"Relacionada con la maestría","documento":null}]}`)
	e.guardar(ctx, t, "capacitacion", `{"capacitaciones":[{"id":"`+idCapacitacion+`","tipoEvento":"CURSO",
		"nombre":"Gobierno corporativo","institucion":"IAEN","fechaInicio":"2022-01-10","fechaFin":"2022-01-20",
		"horas":"40","documento":null}]}`)
	e.guardar(ctx, t, "inhabilidades", `{"respuestas":["SI","NO","NO","NO","NO","NO","SI","NO","NO","NO","NO","NO","NO"],
		"remocionInstitucion":"Superintendencia","remocionCargo":"Director","remocionFechaFin":"2022-12-31",
		"remocionMotivo":"Reestructuración"}`)

	antes, err := e.repo.ObtenerDetalle(ctx, e.postulacionID, e.postulanteID)
	if err != nil {
		t.Fatalf("ObtenerDetalle: %v", err)
	}

	e.normalizar(ctx, t)

	esperado := map[string]int{
		"formacion_academica": 2, "experiencia_profesional": 2, "publicacion": 1,
		"capacitacion": 1, "inhabilidad_respuesta": 13, "respaldo_organizacional": 1,
	}
	for tabla, n := range esperado {
		if obtenido := e.contar(ctx, t, tabla, "postulacion_id", e.postulacionID); obtenido != n {
			t.Errorf("%s: %d filas, esperadas %d", tabla, obtenido, n)
		}
	}
	for _, tabla := range []string{"familiar_institucion", "pareja_postulante", "domicilio_postulante"} {
		if obtenido := e.contar(ctx, t, tabla, "postulante_id", e.postulanteID); obtenido != 1 {
			t.Errorf("%s: %d filas, esperada 1", tabla, obtenido)
		}
	}

	// Documento propio enlazado; uno inexistente queda en NULL.
	var docA, docB *string
	_ = e.pool.QueryRow(ctx, `SELECT documento_id::text FROM formacion_academica WHERE id = $1::uuid`, idFormacionA).Scan(&docA)
	_ = e.pool.QueryRow(ctx, `SELECT documento_id::text FROM formacion_academica WHERE id = $1::uuid`, idFormacionB).Scan(&docB)
	if docA == nil || *docA != e.documentoID || docB != nil {
		t.Errorf("documento_id: A=%v B=%v, esperado A=%s y B=NULL", docA, docB, e.documentoID)
	}

	// Tiempo de experiencia (SRS 7.5) y experiencia actual sin fecha de fin.
	var anios, meses, dias int
	_ = e.pool.QueryRow(ctx, `SELECT tiempo_anios, tiempo_meses, tiempo_dias FROM experiencia_profesional WHERE id = $1::uuid`,
		idExperiencia).Scan(&anios, &meses, &dias)
	if anios != 3 || meses != 1 || dias != 23 {
		t.Errorf("tiempo de experiencia: %d años %d meses %d días, esperado 3/1/23", anios, meses, dias)
	}
	var actuales int
	_ = e.pool.QueryRow(ctx, `SELECT count(*) FROM experiencia_profesional
		WHERE postulacion_id = $1::uuid AND actualmente AND fecha_fin IS NULL`, e.postulacionID).Scan(&actuales)
	if actuales != 1 {
		t.Errorf("experiencia actual sin fecha de fin: %d, esperada 1", actuales)
	}

	// Inhabilidades: 1.9.1 con "Sí" no bloquea; 1.9.7 con "Sí" alerta, guarda la remoción y marca revisión.
	var efecto191, efecto197, institucion string
	var revision, tieneFamiliares, tienePublicaciones bool
	_ = e.pool.QueryRow(ctx, `SELECT r.efecto_calculado::text FROM inhabilidad_respuesta r
		JOIN inhabilidad_pregunta p ON p.id = r.pregunta_id WHERE r.postulacion_id = $1::uuid AND p.codigo = '1.9.1'`,
		e.postulacionID).Scan(&efecto191)
	_ = e.pool.QueryRow(ctx, `SELECT r.efecto_calculado::text, COALESCE(r.institucion_remocion, '') FROM inhabilidad_respuesta r
		JOIN inhabilidad_pregunta p ON p.id = r.pregunta_id WHERE r.postulacion_id = $1::uuid AND p.codigo = '1.9.7'`,
		e.postulacionID).Scan(&efecto197, &institucion)
	_ = e.pool.QueryRow(ctx, `SELECT requiere_revision, tiene_familiares_institucion, tiene_publicaciones
		FROM postulacion WHERE id = $1::uuid`, e.postulacionID).Scan(&revision, &tieneFamiliares, &tienePublicaciones)
	if efecto191 != "INFORMATIVO" || efecto197 != "ALERTA" || institucion != "Superintendencia" {
		t.Errorf("inhabilidades: 1.9.1=%s 1.9.7=%s institución=%q", efecto191, efecto197, institucion)
	}
	if !revision || !tieneFamiliares || !tienePublicaciones {
		t.Errorf("postulacion: requiere_revision=%v tiene_familiares=%v tiene_publicaciones=%v", revision, tieneFamiliares, tienePublicaciones)
	}

	// Lo que lee la SPA no cambia con el volcado.
	despues, err := e.repo.ObtenerDetalle(ctx, e.postulacionID, e.postulanteID)
	if err != nil {
		t.Fatalf("ObtenerDetalle: %v", err)
	}
	for seccion, datos := range antes.Secciones {
		if string(despues.Secciones[seccion]) != string(datos) {
			t.Errorf("la sección %s cambió tras normalizar", seccion)
		}
	}

	// Reenvío: se quita una formación, se responde "No" a familiares y publicaciones y cambia el estado civil.
	e.guardar(ctx, t, "personal", `{"estadoCivil":"SOLTERO","parejaNombres":"","parejaCedula":""}`)
	e.guardar(ctx, t, "familiares", `{"tieneFamiliares":"NO","familiares":[]}`)
	e.guardar(ctx, t, "publicaciones", `{"tienePublicaciones":"NO","publicaciones":[]}`)
	e.guardar(ctx, t, "formacion", `{"formaciones":[
		{"id":"`+idFormacionA+`","nivel":"TERCER_NIVEL","institucion":"ESPOL","tipoInstitucion":"NACIONAL",
		 "titulo":"Ingeniero en Computación","area":"Computación","numeroSenescyt":"1006-2015-1",
		 "fechaSenescyt":"2015-03-01","documento":{"id":"`+e.documentoID+`"}}]}`)

	var idExperienciaTmp string
	_ = e.pool.QueryRow(ctx, `SELECT id::text FROM experiencia_profesional
		WHERE postulacion_id = $1::uuid AND actualmente`, e.postulacionID).Scan(&idExperienciaTmp)

	e.normalizar(ctx, t)
	e.normalizar(ctx, t) // idempotente: un segundo volcado no acumula filas

	if n := e.contar(ctx, t, "formacion_academica", "postulacion_id", e.postulacionID); n != 1 {
		t.Errorf("formacion_academica tras reenvío: %d filas, esperada 1", n)
	}
	var titulo string
	_ = e.pool.QueryRow(ctx, `SELECT titulo_obtenido FROM formacion_academica WHERE id = $1::uuid`, idFormacionA).Scan(&titulo)
	if titulo != "Ingeniero en Computación" {
		t.Errorf("la formación conservada debía actualizarse con el mismo id, título=%q", titulo)
	}
	var idExperienciaTmp2 string
	_ = e.pool.QueryRow(ctx, `SELECT id::text FROM experiencia_profesional
		WHERE postulacion_id = $1::uuid AND actualmente`, e.postulacionID).Scan(&idExperienciaTmp2)
	if idExperienciaTmp == "" || idExperienciaTmp != idExperienciaTmp2 {
		t.Errorf("un ítem con id temporal debía conservar su fila: %q -> %q", idExperienciaTmp, idExperienciaTmp2)
	}
	for tabla, n := range map[string]int{"experiencia_profesional": 2, "publicacion": 0, "inhabilidad_respuesta": 13} {
		if obtenido := e.contar(ctx, t, tabla, "postulacion_id", e.postulacionID); obtenido != n {
			t.Errorf("%s tras reenvío: %d filas, esperadas %d", tabla, obtenido, n)
		}
	}
	for tabla, n := range map[string]int{"familiar_institucion": 0, "pareja_postulante": 0, "domicilio_postulante": 1} {
		if obtenido := e.contar(ctx, t, tabla, "postulante_id", e.postulanteID); obtenido != n {
			t.Errorf("%s tras reenvío: %d filas, esperadas %d", tabla, obtenido, n)
		}
	}
}

// Un error en cualquier sección revierte todo el volcado (la transacción de envío no queda a medias).
func TestNormalizador_ErrorRevierteTodo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	e := nuevoEscenarioNormalizacion(ctx, t)

	e.guardar(ctx, t, "formacion", `{"formaciones":[{"id":"`+idFormacionA+`","nivel":"TERCER_NIVEL",
		"institucion":"ESPOL","tipoInstitucion":"NACIONAL","titulo":"Ingeniero","area":"Computación",
		"numeroSenescyt":"1","fechaSenescyt":"2015-03-01","documento":null}]}`)
	e.guardar(ctx, t, "capacitacion", `{"capacitaciones":[{"id":"`+idCapacitacion+`","tipoEvento":"CURSO",
		"nombre":"X","institucion":"Y","fechaInicio":"2022-01-10","fechaFin":"2022-01-20","horas":"cero","documento":null}]}`)

	normalizador := database.NewPostgresNormalizador(e.pool)
	err := database.NewPgxTxManager(e.pool).WithTransaction(ctx, func(txCtx context.Context) error {
		return normalizador.Normalizar(txCtx, e.postulacionID)
	})
	if err == nil {
		t.Fatal("unas horas inválidas debían hacer fallar el volcado")
	}
	if n := e.contar(ctx, t, "formacion_academica", "postulacion_id", e.postulacionID); n != 0 {
		t.Errorf("formacion_academica: %d filas tras el rollback, esperadas 0", n)
	}
}

// Un id de ítem que ya pertenece a otra postulación no sobrescribe la fila ajena.
func TestNormalizador_IDAjenoNoSobrescribe(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	formacion := func(titulo string) string {
		return `{"formaciones":[{"id":"` + idFormacionA + `","nivel":"TERCER_NIVEL","institucion":"ESPOL",
			"tipoInstitucion":"NACIONAL","titulo":"` + titulo + `","area":"Computación","numeroSenescyt":"1",
			"fechaSenescyt":"2015-03-01","documento":null}]}`
	}

	duena := nuevoEscenarioNormalizacion(ctx, t)
	duena.guardar(ctx, t, "formacion", formacion("Original"))
	duena.normalizar(ctx, t)

	otra := nuevoEscenarioNormalizacion(ctx, t)
	otra.guardar(ctx, t, "formacion", formacion("Intruso"))
	err := database.NewPostgresNormalizador(otra.pool).Normalizar(ctx, otra.postulacionID)
	if !errors.Is(err, database.ErrIDItemAjeno) {
		t.Fatalf("esperado ErrIDItemAjeno, obtenido %v", err)
	}

	var titulo string
	_ = duena.pool.QueryRow(ctx, `SELECT titulo_obtenido FROM formacion_academica WHERE id = $1::uuid`, idFormacionA).Scan(&titulo)
	if titulo != "Original" {
		t.Errorf("la fila de otra postulación fue modificada: título=%q", titulo)
	}
}

func TestNormalizador_PostulacionInexistente(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	e := nuevoEscenarioNormalizacion(ctx, t)
	normalizador := database.NewPostgresNormalizador(e.pool)

	for _, id := range []string{"no-es-uuid", "44444444-4444-4444-8444-444444444444"} {
		if err := normalizador.Normalizar(ctx, id); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("Normalizar(%q): esperado ErrNotFound, obtenido %v", id, err)
		}
	}
}

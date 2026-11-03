package domain_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"srp-biess/internal/domain"
)

func TestSincronizacionReglasJSON(t *testing.T) {
	// 1. Verificar que el JSON embebido sea válido
	var backendReglas map[string]any
	if err := json.Unmarshal(domain.ReglasJSON, &backendReglas); err != nil {
		t.Fatalf("el archivo secciones_reglas.json en backend no es un JSON válido: %v", err)
	}

	// 2. Verificar que el archivo en frontend exista y sea idéntico
	frontendPath := filepath.Join("..", "..", "frontend", "src", "features", "postulacion", "schemas", "secciones_reglas.json")
	frontendBytes, err := os.ReadFile(frontendPath)
	if err != nil {
		t.Fatalf("no se pudo leer el archivo de reglas en frontend (%s): %v", frontendPath, err)
	}

	var frontendReglas map[string]any
	if err := json.Unmarshal(frontendBytes, &frontendReglas); err != nil {
		t.Fatalf("el archivo de reglas en frontend no es un JSON válido: %v", err)
	}

	// Normalizar y comparar
	normBackend, _ := json.Marshal(backendReglas)
	normFrontend, _ := json.Marshal(frontendReglas)
	if !bytes.Equal(normBackend, normFrontend) {
		t.Fatal("desincronización detectada: internal/domain/secciones_reglas.json y frontend/src/features/postulacion/schemas/secciones_reglas.json no coinciden")
	}

	// 3. Verificar que todas las 9 secciones editables estén definidas en el JSON
	seccionesMap, ok := backendReglas["secciones"].(map[string]any)
	if !ok {
		t.Fatal("el archivo secciones_reglas.json no contiene la clave 'secciones'")
	}

	for _, seccion := range domain.SeccionesFormulario {
		if _, existe := seccionesMap[seccion]; !existe {
			t.Errorf("la sección %q requerida por domain.SeccionesFormulario no está en secciones_reglas.json", seccion)
		}
	}
}

func TestValidarSeccion_Personal(t *testing.T) {
	ctx := domain.ContextoPostulante{
		CedulaPostulante: "0926687856",
	}

	t.Run("sección vacía falla con estadoCivil", func(t *testing.T) {
		err := domain.ValidarSeccion("personal", []byte(`{}`), ctx)
		if err == nil {
			t.Fatal("esperaba error de validación")
		}
		valErr, ok := err.(*domain.ErrorValidacionSeccion)
		if !ok || len(valErr.Detalles) == 0 || valErr.Detalles[0].Campo != "estadoCivil" {
			t.Fatalf("esperado error en estadoCivil, obtenido %+v", err)
		}
	})

	t.Run("soltero válido pasa", func(t *testing.T) {
		datos := []byte(`{"estadoCivil": "SOLTERO"}`)
		if err := domain.ValidarSeccion("personal", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})

	t.Run("casado sin datos de pareja falla", func(t *testing.T) {
		datos := []byte(`{"estadoCivil": "CASADO"}`)
		err := domain.ValidarSeccion("personal", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por falta de datos de pareja")
		}
		valErr := err.(*domain.ErrorValidacionSeccion)
		campos := map[string]bool{}
		for _, d := range valErr.Detalles {
			campos[d.Campo] = true
		}
		if !campos["parejaNombres"] || !campos["parejaCedula"] {
			t.Fatalf("esperados errores en parejaNombres y parejaCedula, obtenido: %+v", valErr.Detalles)
		}
	})

	t.Run("casado con misma cédula del postulante falla", func(t *testing.T) {
		datos := []byte(`{
			"estadoCivil": "CASADO",
			"parejaNombres": "Ana Pérez",
			"parejaCedula": "0926687856"
		}`)
		err := domain.ValidarSeccion("personal", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por cédula igual a la del postulante")
		}
		valErr := err.(*domain.ErrorValidacionSeccion)
		if len(valErr.Detalles) != 1 || valErr.Detalles[0].Campo != "parejaCedula" {
			t.Fatalf("esperado error en parejaCedula, obtenido: %+v", valErr.Detalles)
		}
	})

	t.Run("casado con datos completos pasa", func(t *testing.T) {
		datos := []byte(`{
			"estadoCivil": "CASADO",
			"parejaNombres": "Ana Pérez",
			"parejaCedula": "0912345674"
		}`)
		if err := domain.ValidarSeccion("personal", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})
}

func TestValidarSeccion_Domicilio(t *testing.T) {
	ctx := domain.ContextoPostulante{
		PaisEcuadorID:   "52",
		CorreoPrincipal: "postulante@biess.gob.ec",
	}

	t.Run("domicilio incompleto en Ecuador reporta campos faltantes", func(t *testing.T) {
		datos := []byte(`{
			"paisId": "EC",
			"callePrincipal": "Av. Amazonas"
		}`)
		err := domain.ValidarSeccion("domicilio", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por campos obligatorios faltantes")
		}
		valErr := err.(*domain.ErrorValidacionSeccion)
		campos := map[string]bool{}
		for _, d := range valErr.Detalles {
			campos[d.Campo] = true
		}
		if !campos["provinciaId"] || !campos["cantonId"] || !campos["telefonoCelular"] || !campos["numero"] {
			t.Fatalf("faltan campos esperados en el error: %+v", valErr.Detalles)
		}
	})

	t.Run("domicilio en Ecuador completo con codigo EC pasa", func(t *testing.T) {
		datos := []byte(`{
			"paisId": "EC",
			"provinciaId": "17",
			"cantonId": "1701",
			"telefonoCelular": "0991234567",
			"callePrincipal": "Av. 10 de Agosto",
			"numero": "N24-50"
		}`)
		if err := domain.ValidarSeccion("domicilio", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})

	t.Run("domicilio en Ecuador completo con paisId=52 (semilla db/init.sql) pasa con provincia y cantón", func(t *testing.T) {
		datos := []byte(`{
			"paisId": "52",
			"provinciaId": "17",
			"cantonId": "1701",
			"telefonoCelular": "0991234567",
			"callePrincipal": "Av. 10 de Agosto",
			"numero": "N24-50"
		}`)
		if err := domain.ValidarSeccion("domicilio", datos, ctx); err != nil {
			t.Fatalf("esperado válido para paisId 52, obtenido: %v", err)
		}
	})

	t.Run("domicilio en Ecuador con paisId=52 sin provincia ni cantón no exige campos manuales", func(t *testing.T) {
		datos := []byte(`{
			"paisId": "52",
			"telefonoCelular": "0991234567",
			"callePrincipal": "Av. 10 de Agosto",
			"numero": "N24-50"
		}`)
		err := domain.ValidarSeccion("domicilio", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por falta de provincia y cantón")
		}
		valErr := err.(*domain.ErrorValidacionSeccion)
		campos := map[string]bool{}
		for _, d := range valErr.Detalles {
			campos[d.Campo] = true
		}
		if !campos["provinciaId"] || !campos["cantonId"] {
			t.Fatalf("esperaba errores en provinciaId y cantonId, obtenido: %+v", valErr.Detalles)
		}
		if campos["provinciaManual"] || campos["ciudadManual"] {
			t.Fatalf("para paisId=52 no deben exigirse campos manuales, obtenido: %+v", valErr.Detalles)
		}
	})

	t.Run("domicilio en el extranjero requiere campos manuales", func(t *testing.T) {
		datos := []byte(`{
			"paisId": "US",
			"provinciaManual": "Florida",
			"ciudadManual": "Miami",
			"telefonoCelular": "+13051234567",
			"callePrincipal": "Ocean Drive",
			"numero": "100"
		}`)
		if err := domain.ValidarSeccion("domicilio", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})

	t.Run("correo alterno idéntico al principal falla", func(t *testing.T) {
		datos := []byte(`{
			"paisId": "EC",
			"provinciaId": "17",
			"cantonId": "1701",
			"telefonoCelular": "0991234567",
			"correoAlterno": "postulante@biess.gob.ec",
			"callePrincipal": "Av. 10 de Agosto",
			"numero": "N24-50"
		}`)
		err := domain.ValidarSeccion("domicilio", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por correo alterno idéntico al principal")
		}
		valErr := err.(*domain.ErrorValidacionSeccion)
		if len(valErr.Detalles) != 1 || valErr.Detalles[0].Campo != "correoAlterno" {
			t.Fatalf("esperado error en correoAlterno, obtenido: %+v", valErr.Detalles)
		}
	})
}

func TestValidarSeccion_Familiares(t *testing.T) {
	ctx := domain.ContextoPostulante{}

	t.Run("no tiene familiares pasa con arreglo vacío", func(t *testing.T) {
		datos := []byte(`{
			"tieneFamiliares": "NO",
			"familiares": []
		}`)
		if err := domain.ValidarSeccion("familiares", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})

	t.Run("tiene familiares SI pero arreglo vacío falla", func(t *testing.T) {
		datos := []byte(`{
			"tieneFamiliares": "SI",
			"familiares": []
		}`)
		err := domain.ValidarSeccion("familiares", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por lista vacía de familiares")
		}
	})

	t.Run("tiene familiar completo y válido pasa", func(t *testing.T) {
		datos := []byte(`{
			"tieneFamiliares": "SI",
			"familiares": [
				{
					"id": "1",
					"nombresApellidos": "Carlos Pérez",
					"parentesco": "HERMANO",
					"institucion": "BIESS",
					"area": "Riesgos",
					"cargo": "Analista"
				}
			]
		}`)
		if err := domain.ValidarSeccion("familiares", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})
}

func TestValidarSeccion_Cargo(t *testing.T) {
	ctx := domain.ContextoPostulante{}

	t.Run("sin respaldo pasa", func(t *testing.T) {
		datos := []byte(`{"tieneRespaldo": "NO"}`)
		if err := domain.ValidarSeccion("cargo", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})

	t.Run("con respaldo sin documento ni nombre falla", func(t *testing.T) {
		datos := []byte(`{"tieneRespaldo": "SI"}`)
		err := domain.ValidarSeccion("cargo", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por falta de nombre y documento")
		}
		valErr := err.(*domain.ErrorValidacionSeccion)
		campos := map[string]bool{}
		for _, d := range valErr.Detalles {
			campos[d.Campo] = true
		}
		if !campos["nombreOrganizacion"] || !campos["documento"] {
			t.Fatalf("esperados campos nombreOrganizacion y documento, obtenido: %+v", valErr.Detalles)
		}
	})

	t.Run("con respaldo completo pasa", func(t *testing.T) {
		datos := []byte(`{
			"tieneRespaldo": "SI",
			"nombreOrganizacion": "Confederación Nacional de Trabajadores",
			"documento": {"id": "doc-respaldo-123"}
		}`)
		if err := domain.ValidarSeccion("cargo", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})
}

func TestValidarSeccion_Formacion(t *testing.T) {
	ctx := domain.ContextoPostulante{}

	t.Run("arreglo vacío falla", func(t *testing.T) {
		datos := []byte(`{"formaciones": []}`)
		err := domain.ValidarSeccion("formacion", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por formaciones vacías")
		}
	})

	t.Run("formación completa pasa", func(t *testing.T) {
		datos := []byte(`{
			"formaciones": [
				{
					"id": "1",
					"nivel": "TERCER_NIVEL",
					"institucion": "Escuela Politécnica Nacional",
					"tipoInstitucion": "NACIONAL",
					"titulo": "Ingeniero en Finanzas",
					"area": "Ciencias Económicas",
					"numeroSenescyt": "1005-2015-123456",
					"fechaSenescyt": "2015-10-15",
					"documento": {"id": "doc-titulo-1"}
				}
			]
		}`)
		if err := domain.ValidarSeccion("formacion", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})

	t.Run("fecha senescyt futura falla", func(t *testing.T) {
		datos := []byte(`{
			"formaciones": [
				{
					"id": "1",
					"nivel": "TERCER_NIVEL",
					"institucion": "EPN",
					"tipoInstitucion": "NACIONAL",
					"titulo": "Ingeniero",
					"area": "Finanzas",
					"numeroSenescyt": "123",
					"fechaSenescyt": "2099-12-31",
					"documento": {"id": "doc-1"}
				}
			]
		}`)
		err := domain.ValidarSeccion("formacion", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por fecha futura")
		}
	})
}

func TestValidarSeccion_Experiencia(t *testing.T) {
	ctx := domain.ContextoPostulante{}

	t.Run("experiencias vacía es válida", func(t *testing.T) {
		datos := []byte(`{"experiencias": []}`)
		if err := domain.ValidarSeccion("experiencia", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})

	t.Run("experiencia con fin anterior a inicio falla", func(t *testing.T) {
		datos := []byte(`{
			"experiencias": [
				{
					"id": "1",
					"tipo": "GENERAL",
					"cargo": "Analista",
					"institucion": "Banco",
					"fechaInicio": "2022-01-01",
					"actualmente": false,
					"fechaFin": "2020-01-01",
					"actividades": "Análisis",
					"funciones": "Gestión",
					"documento": {"id": "doc-exp-1"}
				}
			]
		}`)
		err := domain.ValidarSeccion("experiencia", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por fecha fin anterior a inicio")
		}
	})

	t.Run("experiencia actualmente trabajando pasa sin fechaFin", func(t *testing.T) {
		datos := []byte(`{
			"experiencias": [
				{
					"id": "1",
					"tipo": "GENERAL",
					"cargo": "Director",
					"institucion": "Banco",
					"fechaInicio": "2021-01-01",
					"actualmente": true,
					"fechaFin": "",
					"actividades": "Dirección financiera",
					"funciones": "Liderazgo de equipo",
					"documento": {"id": "doc-exp-1"}
				}
			]
		}`)
		if err := domain.ValidarSeccion("experiencia", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})
}

func TestValidarSeccion_Inhabilidades(t *testing.T) {
	ctx := domain.ContextoPostulante{}

	t.Run("menos de 13 respuestas falla", func(t *testing.T) {
		datos := []byte(`{"respuestas": ["NO", "NO"]}`)
		err := domain.ValidarSeccion("inhabilidades", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por respuestas incompletas")
		}
	})

	t.Run("13 respuestas NO pasa", func(t *testing.T) {
		datos := []byte(`{
			"respuestas": ["NO", "NO", "NO", "NO", "NO", "NO", "NO", "NO", "NO", "NO", "NO", "NO", "NO"]
		}`)
		if err := domain.ValidarSeccion("inhabilidades", datos, ctx); err != nil {
			t.Fatalf("esperado válido, obtenido: %v", err)
		}
	})

	t.Run("pregunta 7 SI sin datos de remoción falla", func(t *testing.T) {
		datos := []byte(`{
			"respuestas": ["NO", "NO", "NO", "NO", "NO", "NO", "SI", "NO", "NO", "NO", "NO", "NO", "NO"]
		}`)
		err := domain.ValidarSeccion("inhabilidades", datos, ctx)
		if err == nil {
			t.Fatal("esperaba error por falta de campos de remoción")
		}
		valErr := err.(*domain.ErrorValidacionSeccion)
		campos := map[string]bool{}
		for _, d := range valErr.Detalles {
			campos[d.Campo] = true
		}
		if !campos["remocionInstitucion"] || !campos["remocionCargo"] || !campos["remocionFechaFin"] || !campos["remocionMotivo"] {
			t.Fatalf("faltan campos de remoción en el error: %+v", valErr.Detalles)
		}
	})
}

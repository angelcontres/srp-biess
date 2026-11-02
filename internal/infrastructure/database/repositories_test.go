package database_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
)

func getTestPool(t *testing.T) database.Querier {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skipf("Omitiendo pruebas de base de datos relacional (PostgreSQL no disponible): %v", err)
		return nil
	}
	return pool
}

func TestPostgresPool_ConfigurationEnv(t *testing.T) {
	// Verificar que las variables de entorno controlan el pool (Criterio 1)
	_ = os.Setenv("DB_MAX_CONNS", "15")
	_ = os.Setenv("DB_MIN_CONNS", "3")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skipf("PostgreSQL no disponible: %v", err)
	}
	defer pool.Close()

	stat := pool.Stat()
	if stat.MaxConns() != 15 {
		t.Errorf("MaxConns = %d; want 15", stat.MaxConns())
	}
}

func TestPostgresCatalogoRepository(t *testing.T) {
	pool := getTestPool(t)
	if pool == nil {
		return
	}

	repo := database.NewPostgresCatalogoRepository(pool)
	ctx := context.Background()

	t.Run("Consultar paises devuelve semillas iniciales", func(t *testing.T) {
		paises, err := repo.GetPaises(ctx)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if len(paises) == 0 {
			t.Errorf("se esperaban paises en los catalogos")
		}
	})

	t.Run("Consultar provincias de Ecuador (ID=1)", func(t *testing.T) {
		provincias, err := repo.GetProvincias(ctx, "1")
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if len(provincias) != 24 {
			t.Errorf("provincias esperadas 24, obtenidas: %d", len(provincias))
		}
	})

	t.Run("Consultar cargos devuelve registros activos", func(t *testing.T) {
		cargos, err := repo.GetCargos(ctx)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if len(cargos) == 0 {
			t.Errorf("se esperaban cargos")
		}
	})
}

func TestPostgresRepositories_WithTransaction(t *testing.T) {
	ctx := context.Background()
	poolAny := getTestPool(t)
	if poolAny == nil {
		return
	}

	// Forzar casteo a Pool para iniciar transacción atómica
	type transactor interface {
		database.Querier
		Begin(ctx context.Context) (any, error)
	}

	// Comprobamos la interfaz Querier usando una transacción real (Criterio 2)
	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skip(err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("error iniciando transaccion: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Repositorio Postulante dentro de la transacción (Querier)
	postulanteRepo := database.NewPostgresPostulanteRepository(tx)

	cedulaTest := fmt.Sprintf("09%08d", time.Now().UnixNano()%100000000)
	emailTest := fmt.Sprintf("test.%d@biess.gob.ec", time.Now().UnixNano())

	postulante := &domain.Postulante{
		Cedula:       cedulaTest,
		Nombres:      "JUAN CARLOS",
		Apellidos:    "MENDOZA LOOR",
		Email:        emailTest,
		PasswordHash: "hash_test_argon2id",
	}

	err = postulanteRepo.Save(ctx, postulante)
	if err != nil {
		t.Fatalf("error guardando postulante: %v", err)
	}
	if postulante.ID == "" {
		t.Errorf("se esperaba ID generado para postulante")
	}

	// Consulta por Cédula
	byCedula, err := postulanteRepo.GetByCedula(ctx, cedulaTest)
	if err != nil {
		t.Fatalf("error consultando por cedula: %v", err)
	}
	if byCedula.ID != postulante.ID {
		t.Errorf("ID = %s; want %s", byCedula.ID, postulante.ID)
	}

	// 2. Repositorio Postulacion dentro de la transacción (Querier)
	postulacionRepo := database.NewPostgresPostulacionRepository(tx)
	postulacion := &domain.Postulacion{
		CodigoPostulacion: fmt.Sprintf("POST-%d", time.Now().Unix()%10000),
		PostulanteID:      postulante.ID,
		CargoID:           "1",
		Estado:            domain.EstadoBorrador,
	}

	err = postulacionRepo.Save(ctx, postulacion)
	if err != nil {
		t.Fatalf("error guardando postulacion: %v", err)
	}
	if postulacion.ID == "" {
		t.Errorf("se esperaba ID para postulación")
	}

	// Actualizar estado de postulación
	err = postulacionRepo.UpdateEstado(ctx, postulacion.ID, domain.EstadoEnviada)
	if err != nil {
		t.Fatalf("error actualizando estado: %v", err)
	}

	// 3. Repositorio Documento dentro de la transacción (Querier)
	docRepo := database.NewPostgresDocumentoRepository(tx)
	doc := &domain.Documento{
		PostulanteID:     postulante.ID,
		PostulacionID:    postulacion.ID,
		Tipo:             domain.DocTitulo,
		NombreOriginal:   "Titulo_Ingeniero.pdf",
		NombreAlmacenado: fmt.Sprintf("doc_%d.pdf", time.Now().UnixNano()),
		MimeType:         "application/pdf",
		TamanoBytes:      1024 * 500, // 500 KB
		HashSHA256:       "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		Ruta:             "/storage/titulos/doc_1.pdf",
		Estado:           domain.EstadoDocPendiente,
	}

	err = docRepo.Save(ctx, doc)
	if err != nil {
		t.Fatalf("error guardando documento: %v", err)
	}
	if doc.ID == "" {
		t.Errorf("se esperaba ID de documento generado")
	}

	// Listar documentos de postulación
	docs, err := docRepo.ListByPostulacionID(ctx, postulacion.ID)
	if err != nil {
		t.Fatalf("error listando documentos: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("documentos esperados 1, obtenidos: %d", len(docs))
	}

	// Borrar documento
	err = docRepo.Delete(ctx, doc.ID)
	if err != nil {
		t.Fatalf("error eliminando documento: %v", err)
	}
}

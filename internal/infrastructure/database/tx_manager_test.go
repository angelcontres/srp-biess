package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/database"
)

// faultyDocRepo simula un fallo inducido durante la inserción de documento para forzar rollback
type faultyDocRepo struct {
	domain.DocumentoRepository
	shouldFail bool
}

func (f *faultyDocRepo) Save(ctx context.Context, doc *domain.Documento) error {
	if f.shouldFail {
		return errors.New("error simulado de I/O en almacenamiento documental")
	}
	return f.DocumentoRepository.Save(ctx, doc)
}

func TestTxManager_CommitAndRollback(t *testing.T) {
	ctx := context.Background()
	poolAny := getTestPool(t)
	if poolAny == nil {
		return
	}

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skip(err)
	}
	defer pool.Close()

	txManager := database.NewPgxTxManager(pool)
	postulanteRepo := database.NewPostgresPostulanteRepository(pool)

	t.Run("Commit exitoso persiste los datos en base de datos", func(t *testing.T) {
		cedula := fmt.Sprintf("09%08d", time.Now().UnixNano()%100000000)
		email := fmt.Sprintf("tx_commit_%d@biess.gob.ec", time.Now().UnixNano())
		p := &domain.Postulante{
			Cedula:       cedula,
			Nombres:      "MARIO",
			Apellidos:    "BROS",
			Email:        email,
			PasswordHash: "secret",
		}

		err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
			return postulanteRepo.Save(txCtx, p)
		})
		if err != nil {
			t.Fatalf("error en transaccion: %v", err)
		}

		// Validar fuera de la transacción
		found, err := postulanteRepo.GetByCedula(ctx, cedula)
		if err != nil {
			t.Fatalf("se esperaba encontrar al postulante tras commit: %v", err)
		}
		if found.ID != p.ID {
			t.Errorf("ID = %s; want %s", found.ID, p.ID)
		}
	})

	t.Run("Rollback forzado por error descarta los cambios (Base Intacta)", func(t *testing.T) {
		cedula := fmt.Sprintf("09%08d", time.Now().UnixNano()%100000000)
		email := fmt.Sprintf("tx_rollback_%d@biess.gob.ec", time.Now().UnixNano())
		p := &domain.Postulante{
			Cedula:       cedula,
			Nombres:      "LUIGI",
			Apellidos:    "BROS",
			Email:        email,
			PasswordHash: "secret",
		}

		simulatedErr := errors.New("fallo deliberado para activar rollback")

		err := txManager.WithTransaction(ctx, func(txCtx context.Context) error {
			if err := postulanteRepo.Save(txCtx, p); err != nil {
				return err
			}
			return simulatedErr
		})

		if !errors.Is(err, simulatedErr) {
			t.Fatalf("se esperaba error simulado, obtenido: %v", err)
		}

		// Validar que NO existe en la base de datos tras el rollback
		_, err = postulanteRepo.GetByCedula(ctx, cedula)
		if !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("se esperaba ErrNotFound tras rollback, obtenido: %v", err)
		}
	})
}

func TestPostulacionService_AtomicSubmitAndForcedRollback(t *testing.T) {
	ctx := context.Background()
	poolAny := getTestPool(t)
	if poolAny == nil {
		return
	}

	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		t.Skip(err)
	}
	defer pool.Close()

	txManager := database.NewPgxTxManager(pool)
	postulanteRepo := database.NewPostgresPostulanteRepository(pool)
	postulacionRepo := database.NewPostgresPostulacionRepository(pool)
	realDocRepo := database.NewPostgresDocumentoRepository(pool)
	notifRepo := database.NewPostgresNotificacionRepository(pool)

	// Crear Postulante base
	cedula := fmt.Sprintf("09%08d", time.Now().UnixNano()%100000000)
	email := fmt.Sprintf("postulante_%d@biess.gob.ec", time.Now().UnixNano())
	postulante := &domain.Postulante{
		Cedula:       cedula,
		Nombres:      "ANA GABRIELA",
		Apellidos:    "MOLINA SILVA",
		Email:        email,
		PasswordHash: "hash_seguro",
	}
	if err := postulanteRepo.Save(ctx, postulante); err != nil {
		t.Fatalf("error creando postulante base: %v", err)
	}

	t.Run("1. Happy Path Atómico: Actualiza estado, genera código, crea PDF y encola notificación", func(t *testing.T) {
		// Crear postulación en BORRADOR
		postulacion := &domain.Postulacion{
			PostulanteID: postulante.ID,
			CargoID:      "1",
			Estado:       domain.EstadoBorrador,
		}
		if err := postulacionRepo.Save(ctx, postulacion); err != nil {
			t.Fatalf("error guardando postulacion borrador: %v", err)
		}

		svc := application.NewPostulacionService(txManager, postulacionRepo, realDocRepo, notifRepo)

		res, err := svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
			PostulacionID: postulacion.ID,
			Email:         postulante.Email,
		})
		if err != nil {
			t.Fatalf("error en EnviarPostulacion: %v", err)
		}

		if res.CodigoUnico == "" {
			t.Errorf("se esperaba código único generado")
		}
		if res.Estado != string(domain.EstadoEnviada) {
			t.Errorf("estado = %s; want %s", res.Estado, domain.EstadoEnviada)
		}

		// 1.1 Validar postulación en PostgreSQL
		pDB, err := postulacionRepo.GetByID(ctx, postulacion.ID)
		if err != nil {
			t.Fatalf("error consultando postulación: %v", err)
		}
		if pDB.Estado != domain.EstadoEnviada {
			t.Errorf("estado en BD = %s; want %s", pDB.Estado, domain.EstadoEnviada)
		}
		if pDB.CodigoPostulacion != res.CodigoUnico {
			t.Errorf("código en BD = %s; want %s", pDB.CodigoPostulacion, res.CodigoUnico)
		}
		if pDB.FechaEnvio == nil {
			t.Errorf("se esperaba fecha_envio no nula")
		}

		// 1.2 Validar documento PDF en PostgreSQL
		docs, err := realDocRepo.ListByPostulacionID(ctx, postulacion.ID)
		if err != nil {
			t.Fatalf("error listando documentos: %v", err)
		}
		if len(docs) == 0 {
			t.Fatalf("se esperaba documento registrado para la postulación enviada")
		}
		if docs[0].Tipo != domain.DocFormularioPostulacion {
			t.Errorf("tipo doc = %s; want %s", docs[0].Tipo, domain.DocFormularioPostulacion)
		}
	})

	t.Run("2. Rollback Forzado: Fallo en inserción de documento deja la base de datos INTACTA", func(t *testing.T) {
		// Crear segundo Postulante para respetar la restricción única de postulación por convocatoria
		cedula2 := fmt.Sprintf("17%08d", time.Now().UnixNano()%100000000)
		email2 := fmt.Sprintf("postulante_rollback_%d@biess.gob.ec", time.Now().UnixNano())
		postulante2 := &domain.Postulante{
			Cedula:       cedula2,
			Nombres:      "CARLOS EDUARDO",
			Apellidos:    "LOPEZ MERA",
			Email:        email2,
			PasswordHash: "hash_seguro",
		}
		if err := postulanteRepo.Save(ctx, postulante2); err != nil {
			t.Fatalf("error creando postulante 2: %v", err)
		}

		// Crear postulación en BORRADOR
		postulacion := &domain.Postulacion{
			PostulanteID: postulante2.ID,
			CargoID:      "1",
			Estado:       domain.EstadoBorrador,
		}
		if err := postulacionRepo.Save(ctx, postulacion); err != nil {
			t.Fatalf("error guardando postulacion borrador: %v", err)
		}

		// Inyectamos un repositorio de documento que falla intencionalmente
		failingDocRepo := &faultyDocRepo{
			DocumentoRepository: realDocRepo,
			shouldFail:          true,
		}

		svc := application.NewPostulacionService(txManager, postulacionRepo, failingDocRepo, notifRepo)

		_, err := svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
			PostulacionID: postulacion.ID,
			Email:         postulante2.Email,
		})

		if err == nil {
			t.Fatalf("se esperaba que la postulación fallara debido al error documental")
		}

		// VERIFICACIÓN DEL CRITERIO 2: La base de datos debe permanecer 100% intacta
		pDB, err := postulacionRepo.GetByID(ctx, postulacion.ID)
		if err != nil {
			t.Fatalf("error consultando postulación tras rollback: %v", err)
		}

		// Estado debe continuar en BORRADOR
		if pDB.Estado != domain.EstadoBorrador {
			t.Errorf("ESTADO NO DEBIÓ CAMBIAR: estado = %s; want %s", pDB.Estado, domain.EstadoBorrador)
		}

		// Código único debe continuar vacío
		if pDB.CodigoPostulacion != "" {
			t.Errorf("CÓDIGO ÚNICO NO DEBIÓ ASIGNARSE: obtenido = %s", pDB.CodigoPostulacion)
		}

		// Fecha de envío debe ser nula
		if pDB.FechaEnvio != nil {
			t.Errorf("FECHA DE ENVÍO NO DEBIÓ REGISTRARSE: obtenido = %v", pDB.FechaEnvio)
		}

		// Ningún documento debió quedar persistido
		docs, err := realDocRepo.ListByPostulacionID(ctx, postulacion.ID)
		if err != nil {
			t.Fatalf("error consultando docs: %v", err)
		}
		if len(docs) > 0 {
			t.Errorf("DOCUMENTO PERSISTIDO TRAS ROLLBACK: se encontraron %d documentos", len(docs))
		}
	})
}

package database_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
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
	aplicarMigraciones(ctx, t, func(ctx context.Context, sql string) error {
		_, err := pool.Exec(ctx, sql)
		return err
	})

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
	t.Run("3. Transacción anidada reutiliza la transacción existente en el contexto", func(t *testing.T) {
		cedula := fmt.Sprintf("11%08d", time.Now().UnixNano()%100000000)
		email := fmt.Sprintf("nested_%d@biess.gob.ec", time.Now().UnixNano())
		p := &domain.Postulante{
			Cedula:       cedula,
			Nombres:      "JORGE",
			Apellidos:    "ANIDADO",
			Email:        email,
			PasswordHash: "secret",
		}

		err := txManager.WithTransaction(ctx, func(txCtx1 context.Context) error {
			if err := postulanteRepo.Save(txCtx1, p); err != nil {
				return err
			}
			// Llamada anidada que debe reutilizar txCtx1 sin abrir nuevo pool.Begin
			return txManager.WithTransaction(txCtx1, func(txCtx2 context.Context) error {
				pDB, err := postulanteRepo.GetByCedula(txCtx2, cedula)
				if err != nil {
					return err
				}
				if pDB.Nombres != "JORGE" {
					return fmt.Errorf("nombres inesperados en tx anidada: %s", pDB.Nombres)
				}
				return nil
			})
		})

		if err != nil {
			t.Fatalf("error inesperado en transacción anidada: %v", err)
		}

		// Debe persistir tras el commit de la transacción raíz
		pFinal, err := postulanteRepo.GetByCedula(ctx, cedula)
		if err != nil {
			t.Fatalf("se esperaba postulante persistido: %v", err)
		}
		if pFinal.Cedula != cedula {
			t.Errorf("cedula = %s; want %s", pFinal.Cedula, cedula)
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
	aplicarMigraciones(ctx, t, func(ctx context.Context, sql string) error {
		_, err := pool.Exec(ctx, sql)
		return err
	})

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
			Estado:       domain.EstadoListaParaEnvio,
		}
		if err := postulacionRepo.Save(ctx, postulacion); err != nil {
			t.Fatalf("error guardando postulacion borrador: %v", err)
		}

		svc := application.NewPostulacionService(txManager, postulacionRepo, realDocRepo, notifRepo)

		idempotencyKey := fmt.Sprintf("tx-happy-%d", time.Now().UnixNano())
		res, err := svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
			PostulacionID:        postulacion.ID,
			Email:                postulante.Email,
			AceptaConsentimiento: true,
			VersionTerminos:      "2025.1",
			IdempotencyKey:       idempotencyKey,
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

		resRetry, err := svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
			PostulacionID:        postulacion.ID,
			Email:                postulante.Email,
			AceptaConsentimiento: true,
			VersionTerminos:      "2025.1",
			IdempotencyKey:       idempotencyKey,
		})
		if err != nil {
			t.Fatalf("reintento idempotente falló: %v", err)
		}
		if resRetry.CodigoUnico != res.CodigoUnico {
			t.Fatalf("reintento devolvió código distinto: %s vs %s", resRetry.CodigoUnico, res.CodigoUnico)
		}
		var documentos, notificaciones int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM documento WHERE postulacion_id = $1::uuid`, postulacion.ID).Scan(&documentos); err != nil {
			t.Fatalf("contando documentos: %v", err)
		}
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM notificacion WHERE postulacion_id = $1::uuid`, postulacion.ID).Scan(&notificaciones); err != nil {
			t.Fatalf("contando notificaciones: %v", err)
		}
		if documentos != 1 || notificaciones != 1 {
			t.Fatalf("reintento duplicó efectos: documentos=%d notificaciones=%d", documentos, notificaciones)
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
			Estado:       domain.EstadoListaParaEnvio,
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
			PostulacionID:        postulacion.ID,
			Email:                postulante2.Email,
			AceptaConsentimiento: true,
			VersionTerminos:      "2025.1",
			IdempotencyKey:       fmt.Sprintf("tx-rollback-%d", time.Now().UnixNano()),
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
		if pDB.Estado != domain.EstadoListaParaEnvio {
			t.Errorf("ESTADO NO DEBIÓ CAMBIAR: estado = %s; want %s", pDB.Estado, domain.EstadoListaParaEnvio)
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
	t.Run("3. Prevención de doble envío: segundo intento devuelve ErrPostulacionYaEnviada", func(t *testing.T) {
		cedula3 := fmt.Sprintf("18%08d", time.Now().UnixNano()%100000000)
		email3 := fmt.Sprintf("postulante_doble_%d@biess.gob.ec", time.Now().UnixNano())
		postulante3 := &domain.Postulante{
			Cedula:       cedula3,
			Nombres:      "MARIA",
			Apellidos:    "DOBLE ENVIO",
			Email:        email3,
			PasswordHash: "hash_seguro",
		}
		if err := postulanteRepo.Save(ctx, postulante3); err != nil {
			t.Fatalf("error creando postulante 3: %v", err)
		}

		postulacion3 := &domain.Postulacion{
			PostulanteID: postulante3.ID,
			CargoID:      "1",
			Estado:       domain.EstadoListaParaEnvio,
		}
		if err := postulacionRepo.Save(ctx, postulacion3); err != nil {
			t.Fatalf("error guardando postulacion 3: %v", err)
		}

		svc := application.NewPostulacionService(txManager, postulacionRepo, realDocRepo, notifRepo)

		// Primer envío exitoso
		res1, err := svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
			PostulacionID:        postulacion3.ID,
			Email:                postulante3.Email,
			AceptaConsentimiento: true,
			VersionTerminos:      "2025.1",
			IdempotencyKey:       fmt.Sprintf("tx-duplicate-first-%d", time.Now().UnixNano()),
		})
		if err != nil {
			t.Fatalf("primer envío debió ser exitoso: %v", err)
		}
		if res1.CodigoUnico == "" {
			t.Fatalf("primer envío no generó código")
		}

		// Segundo envío concurrente/inmediato debe ser rechazado
		_, err = svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
			PostulacionID:        postulacion3.ID,
			Email:                postulante3.Email,
			AceptaConsentimiento: true,
			VersionTerminos:      "2025.1",
			IdempotencyKey:       fmt.Sprintf("tx-duplicate-second-%d", time.Now().UnixNano()),
		})
		if !errors.Is(err, application.ErrPostulacionYaEnviada) {
			t.Fatalf("se esperaba ErrPostulacionYaEnviada, obtenido: %v", err)
		}
	})

	t.Run("4. Envío exitoso desde estado LISTA_PARA_ENVIO con código correlativo", func(t *testing.T) {
		cedula4 := fmt.Sprintf("19%08d", time.Now().UnixNano()%100000000)
		email4 := fmt.Sprintf("postulante_lista_%d@biess.gob.ec", time.Now().UnixNano())
		postulante4 := &domain.Postulante{
			Cedula:       cedula4,
			Nombres:      "ESTEBAN",
			Apellidos:    "LISTO PARA ENVIO",
			Email:        email4,
			PasswordHash: "hash_seguro",
		}
		if err := postulanteRepo.Save(ctx, postulante4); err != nil {
			t.Fatalf("error creando postulante 4: %v", err)
		}

		postulacion4 := &domain.Postulacion{
			PostulanteID: postulante4.ID,
			CargoID:      "1",
			Estado:       domain.EstadoListaParaEnvio,
		}
		if err := postulacionRepo.Save(ctx, postulacion4); err != nil {
			t.Fatalf("error guardando postulacion 4: %v", err)
		}

		svc := application.NewPostulacionService(txManager, postulacionRepo, realDocRepo, notifRepo)

		res, err := svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
			PostulacionID:        postulacion4.ID,
			Email:                postulante4.Email,
			AceptaConsentimiento: true,
			VersionTerminos:      "2025.1",
			IdempotencyKey:       fmt.Sprintf("tx-list-path-%d", time.Now().UnixNano()),
		})
		if err != nil {
			t.Fatalf("envío desde LISTA_PARA_ENVIO falló: %v", err)
		}

		if res.Estado != string(domain.EstadoEnviada) {
			t.Errorf("estado = %s; want %s", res.Estado, domain.EstadoEnviada)
		}
		if res.CodigoUnico == "" {
			t.Errorf("se esperaba código único correlativo generado")
		}
	})

	t.Run("5. Dos envíos concurrentes con la misma llave producen una respuesta única", func(t *testing.T) {
		cedula5 := fmt.Sprintf("20%08d", time.Now().UnixNano()%100000000)
		postulante5 := &domain.Postulante{
			Cedula:       cedula5,
			Nombres:      "CONCURRENCIA",
			Apellidos:    "IDEMPOTENTE",
			Email:        fmt.Sprintf("postulante_concurrente_%d@biess.gob.ec", time.Now().UnixNano()),
			PasswordHash: "hash_seguro",
		}
		if err := postulanteRepo.Save(ctx, postulante5); err != nil {
			t.Fatalf("error creando postulante concurrente: %v", err)
		}
		postulacion5 := &domain.Postulacion{
			PostulanteID: postulante5.ID,
			CargoID:      "1",
			Estado:       domain.EstadoListaParaEnvio,
		}
		if err := postulacionRepo.Save(ctx, postulacion5); err != nil {
			t.Fatalf("error guardando postulación concurrente: %v", err)
		}

		svc := application.NewPostulacionService(txManager, postulacionRepo, realDocRepo, notifRepo)
		key := fmt.Sprintf("tx-concurrent-%d", time.Now().UnixNano())
		results := make(chan *application.EnviarPostulacionResult, 2)
		errs := make(chan error, 2)
		var wg sync.WaitGroup
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				res, err := svc.EnviarPostulacion(ctx, application.EnviarPostulacionRequest{
					PostulacionID:        postulacion5.ID,
					Email:                postulante5.Email,
					AceptaConsentimiento: true,
					VersionTerminos:      "2025.1",
					IdempotencyKey:       key,
				})
				results <- res
				errs <- err
			}()
		}
		wg.Wait()
		close(results)
		close(errs)

		var codigo string
		for err := range errs {
			if err != nil {
				t.Fatalf("envío concurrente falló: %v", err)
			}
		}
		for res := range results {
			if res == nil {
				t.Fatal("respuesta concurrente nula")
			}
			if codigo == "" {
				codigo = res.CodigoUnico
			} else if codigo != res.CodigoUnico {
				t.Fatalf("códigos concurrentes distintos: %s y %s", codigo, res.CodigoUnico)
			}
		}
		var notificaciones int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM notificacion WHERE postulacion_id = $1::uuid`, postulacion5.ID).Scan(&notificaciones); err != nil {
			t.Fatalf("contando notificaciones concurrentes: %v", err)
		}
		if notificaciones != 1 {
			t.Fatalf("se esperaban 1 notificación, obtenidas %d", notificaciones)
		}
	})
}

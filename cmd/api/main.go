package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/cache"
	"srp-biess/internal/infrastructure/database"
	"srp-biess/internal/infrastructure/external"
	"srp-biess/internal/infrastructure/security"
	"srp-biess/internal/infrastructure/storage"
	interfacesHTTP "srp-biess/internal/interfaces/http"
	"srp-biess/internal/interfaces/http/handlers"
	"srp-biess/internal/interfaces/http/middleware"
)

func main() {
	// =========================================================================
	// FASE 1: Variables de Entorno y Señales del Sistema Operativo
	// =========================================================================
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] Archivo .env no encontrado, usando variables de entorno del sistema")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("========================================================")
	log.Println("  SRP-BIESS: Backend API (Go 1.22+ & Fiber v2)")
	log.Println("  Composition Root & Inyección Manual de Dependencias (ADR-003)")
	log.Println("========================================================")

	// =========================================================================
	// FASE 2: Conexiones de Infraestructura (PostgreSQL & Redis)
	// =========================================================================
	pool, err := database.NewPostgresPool(ctx)
	if err != nil {
		log.Printf("[ALERTA] Error al iniciar pool de base de datos: %v", err)
	} else {
		defer pool.Close()
	}

	rdb := cache.NewRedisClient(ctx)
	defer rdb.Close()

	// Repositorios de Persistencia Relacional PostgreSQL (SRP-004 y SRP-005)
	var postulanteRepo domain.PostulanteRepository
	var catalogoRepo domain.CatalogoRepository
	var postulacionRepo domain.PostulacionRepository
	var documentoRepo domain.DocumentoRepository
	var notificacionRepo domain.NotificacionRepository
	var txManager domain.TxManager
	var postulacionService *application.PostulacionService
	// Se deja en nil (interfaz nula) sin PostgreSQL: el servicio responde 503 en lugar de fallar.
	var formularioRepo domain.FormularioRepository
	var cargoRepo domain.CargoRepository
	var auditoriaRepo domain.AuditoriaRepository
	var cargaDocumentoRepo domain.CargaDocumentoRepository

	if pool != nil {
		postulanteRepo = database.NewPostgresPostulanteRepository(pool)
		// SRP-007: cache-aside en Redis (TTL 24 h); si Redis falla, el decorador consulta PostgreSQL.
		catalogoRepo = cache.NewCachedCatalogoRepository(
			database.NewPostgresCatalogoRepository(pool), rdb, cacheEnv(), cache.DefaultCatalogTTL)
		postulacionRepo = database.NewPostgresPostulacionRepository(pool)
		documentoRepo = database.NewPostgresDocumentoRepository(pool)
		notificacionRepo = database.NewPostgresNotificacionRepository(pool)
		txManager = database.NewPgxTxManager(pool)
		postulacionService = application.NewPostulacionService(txManager, postulacionRepo, documentoRepo, notificacionRepo)
		_ = postulacionService
		formularioPg := database.NewPostgresFormularioRepository(pool)
		formularioRepo = formularioPg
		cargoRepo = formularioPg // SRP-017a: confirmación del cargo
		auditoriaRepo = database.NewPostgresAuditoriaRepository(pool)
		cargaDocumentoRepo = database.NewPostgresDocumentoRepository(pool) // SRP-017b
		_ = postulacionRepo
		_ = documentoRepo
	}

	// Almacén de sesiones, bloqueo y refresh tokens en Redis DB 1 (ADR-005, ADR-006)
	sessionStore := cache.NewRedisSessionStore(rdb)
	authAttemptStore := cache.NewRedisAuthAttemptStore(rdb)

	// =========================================================================
	// FASE 3: Adaptadores de Infraestructura Externa / Mocks (ADR-010)
	// =========================================================================
	registroCivilMock := external.NewMockIdentityProvider() // implements domain.IdentityProvider
	iessMock := external.NewMockEmploymentProvider()        // implements domain.EmploymentHistoryProvider

	// =========================================================================
	// FASE 4: Estrategias de Dominio (Patrón Strategy - ADR-013)
	// =========================================================================
	cedulaStrategy := domain.NewCedulaEcuatorianaStrategy() // implements domain.DocumentValidationStrategy

	// =========================================================================
	// FASE 5: Casos de Uso / Capa de Aplicación (Inyección Manual - ADR-003)
	// =========================================================================
	authService := application.NewAuthService(postulanteRepo, cedulaStrategy, sessionStore).
		WithAuthAttemptStore(authAttemptStore).
		WithIdentityProvider(registroCivilMock)
	eligibilityService := application.NewEligibilityService(iessMock)
	formularioService := application.NewFormularioService(formularioRepo) // SRP-013, integración mínima
	cargoService := application.NewCargoService(txManager, cargoRepo, auditoriaRepo, eligibilityService, iessTimeout())
	limitesDocumento := limitesDocumento()
	var documentoService *application.DocumentoService
	if cargaDocumentoRepo != nil {
		almacen, err := storage.NewAlmacenLocal(envOr("STORAGE_PATH", "./storage"))
		if err != nil {
			log.Fatalf("[FATAL] %v", err)
		}
		documentoService = application.NewDocumentoService(txManager, cargaDocumentoRepo, almacen,
			security.NewFirmadorHMAC(secretoDescarga()), auditoriaRepo, limitesDocumento, vigenciaDescarga())
	}
	// Sin PostgreSQL el handler recibe nil y responde 503 (no se invoca un repositorio nulo).
	var catalogoService *application.CatalogoService
	if catalogoRepo != nil {
		catalogoService = application.NewCatalogoService(catalogoRepo)
	}

	// =========================================================================
	// FASE 6: Capa de Presentación (Handlers HTTP & DTOs)
	// =========================================================================
	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(authService)
	eligibilityHandler := handlers.NewEligibilityHandler(eligibilityService)
	postulacionHandler := handlers.NewPostulacionHandler(formularioService)
	cargoHandler := handlers.NewCargoHandler(cargoService)
	documentoHandler := handlers.NewDocumentoHandler(documentoService)
	catalogoHandler := handlers.NewCatalogoHandler(catalogoService)

	// =========================================================================
	// FASE 7: Servidor Fiber & Middlewares Globales (SRP-001)
	// =========================================================================
	app := fiber.New(fiber.Config{
		AppName:      "SRP-BIESS API v1.0",
		ServerHeader: "Fiber",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		// Corte defensivo de uploads (SDD 13, 9.7): el mayor PDF permitido más 1 MB de margen multipart.
		BodyLimit: int(limitesDocumento.RespaldoBytes) + 1024*1024,
	})

	app.Use(middleware.NewRequestID())
	app.Use(middleware.NewCORS())
	app.Use(middleware.NewStructuredLogger())
	app.Use(middleware.NewRecover())

	// =========================================================================
	// FASE 8: Configuración de Rutas Desacopladas
	// =========================================================================
	interfacesHTTP.SetupRoutes(app, interfacesHTTP.RouterConfig{
		HealthHandler:      healthHandler,
		AuthHandler:        authHandler,
		EligibilityHandler: eligibilityHandler,
		PostulacionHandler: postulacionHandler,
		CargoHandler:       cargoHandler,
		DocumentoHandler:   documentoHandler,
		CatalogoHandler:    catalogoHandler,
		AuthService:        authService,
	})

	// =========================================================================
	// FASE 9: Servidor HTTP & Graceful Shutdown
	// =========================================================================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	serverShutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		log.Println("[INFO] Señal de apagado recibida. Cerrando servidor Fiber...")
		if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
			log.Printf("[ERROR] Error durante shutdown: %v", err)
		}
		close(serverShutdown)
	}()

	log.Printf("[INFO] Servidor escuchando en http://localhost%s\n", addr)
	if err := app.Listen(addr); err != nil {
		log.Printf("[INFO] Servidor detenido: %v\n", err)
	}

	<-serverShutdown
	log.Println("[INFO] Apagado limpio completado. Hasta pronto.")
}

// cacheEnv devuelve el segmento {env} de las claves de Redis (srp:{env}:...), por defecto "dev".
func cacheEnv() string {
	switch env := os.Getenv("APP_ENV"); env {
	case "", "development":
		return "dev"
	default:
		return env
	}
}

// iessTimeout lee IESS_TIMEOUT (p. ej. "5s"); un valor ausente o inválido usa el valor por defecto.
func iessTimeout() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("IESS_TIMEOUT")); err == nil && d > 0 {
		return d
	}
	return application.DefaultIESSTimeout
}

func envOr(clave, porDefecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return porDefecto
}

// limitesDocumento lee MAX_ORGANIZATIONAL_SUPPORT_MB, MAX_DOCUMENT_MB y MAX_TOTAL_UPLOAD_MB (SDD 16.4.4).
// El respaldo no puede pasar de 4 MB: es el CHECK ck_documento_tamano de la base.
func limitesDocumento() domain.LimitesDocumento {
	l := domain.LimitesPorDefecto
	leer := func(clave string, actual int64) int64 {
		if n, err := strconv.ParseInt(os.Getenv(clave), 10, 64); err == nil && n > 0 {
			return n * 1024 * 1024
		}
		return actual
	}
	l.RespaldoBytes = min(leer("MAX_ORGANIZATIONAL_SUPPORT_MB", l.RespaldoBytes), domain.LimitesPorDefecto.RespaldoBytes)
	l.OtrosBytes = min(leer("MAX_DOCUMENT_MB", l.OtrosBytes), l.RespaldoBytes)
	l.TotalBytes = leer("MAX_TOTAL_UPLOAD_MB", l.TotalBytes)
	return l
}

// secretoDescarga lee DOCUMENTOS_URL_SECRETO. Sin él, en desarrollo se genera uno aleatorio (las
// URL firmadas dejan de valer al reiniciar); en producción es obligatorio.
func secretoDescarga() []byte {
	if s := os.Getenv("DOCUMENTOS_URL_SECRETO"); s != "" {
		return []byte(s)
	}
	if os.Getenv("APP_ENV") == "production" {
		log.Fatal("[FATAL] DOCUMENTOS_URL_SECRETO es obligatorio en producción")
	}
	log.Println("[WARN] DOCUMENTOS_URL_SECRETO no definido: se usa un secreto aleatorio de desarrollo")
	secreto := make([]byte, 32)
	if _, err := rand.Read(secreto); err != nil {
		log.Fatalf("[FATAL] no se pudo generar el secreto de descarga: %v", err)
	}
	return secreto
}

// vigenciaDescarga lee DOCUMENTOS_URL_VIGENCIA (p. ej. "5m").
func vigenciaDescarga() time.Duration {
	if d, err := time.ParseDuration(os.Getenv("DOCUMENTOS_URL_VIGENCIA")); err == nil && d > 0 {
		return d
	}
	return application.DefaultVigenciaDescarga
}

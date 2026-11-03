package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"srp-biess/internal/application"
	"srp-biess/internal/domain"
	"srp-biess/internal/infrastructure/cache"
	"srp-biess/internal/infrastructure/database"
	"srp-biess/internal/infrastructure/external"
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

	if pool != nil {
		postulanteRepo = database.NewPostgresPostulanteRepository(pool)
		catalogoRepo = database.NewPostgresCatalogoRepository(pool)
		postulacionRepo = database.NewPostgresPostulacionRepository(pool)
		documentoRepo = database.NewPostgresDocumentoRepository(pool)
		notificacionRepo = database.NewPostgresNotificacionRepository(pool)
		txManager = database.NewPgxTxManager(pool)
		postulacionService = application.NewPostulacionService(txManager, postulacionRepo, documentoRepo, notificacionRepo)
		_ = postulacionService
		formularioRepo = database.NewPostgresFormularioRepository(pool)
		_ = postulacionRepo
		_ = documentoRepo
	}

	// Almacén de sesiones y refresh tokens en Redis (ADR-006)
	sessionStore := cache.NewRedisSessionStore(rdb)

	// Bloqueo de 10 minutos tras 3 intentos fallidos de registro y login (SRP-006)
	attemptLimiter := cache.NewRedisAttemptLimiter(rdb, cache.KeyEnv())

	// =========================================================================
	// FASE 3: Adaptadores de Infraestructura Externa / Mocks (ADR-010)
	// =========================================================================
	// TODO(backend: SRP-020): si la SPA se sirve en otro origen, CORS debe exponer Retry-After (423 de SRP-006).
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
		WithIdentityProvider(registroCivilMock).
		WithAttemptLimiter(attemptLimiter)
	eligibilityService := application.NewEligibilityService(iessMock)
	formularioService := application.NewFormularioService(formularioRepo) // SRP-013, integración mínima
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

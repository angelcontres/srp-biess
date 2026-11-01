# 05. Diseño de componentes backend — SDD

## 5.1 Catálogo de componentes

| Componente | Capa | Responsabilidad | Dependencias |
|---|---|---|---|
| `AuthHandler` | HTTP | Registro, login, refresh, logout, recuperación. | `AuthService`, validator. |
| `PostulacionHandler` | HTTP | Borrador, secciones, guardar, enviar, resumen. | `PostulacionService`. |
| `CatalogoHandler` | HTTP | Países, provincias, cantones. | `CatalogoService`. |
| `DocumentoHandler` | HTTP | Carga y descarga de PDF. | `DocumentoService`. |
| `AuthMiddleware` | HTTP | Validar JWT y cargar identidad. | `TokenService`. |
| `RateLimitMiddleware` | HTTP | Limitar solicitudes por IP/cédula. | Redis. |
| `AuthService` | Aplicación | Orquestar registro, login y recuperación. | Puertos. |
| `PostulacionService` | Aplicación | Crear borrador, guardar secciones, validar y enviar. | Puertos, `EligibilityService`, `TxManager`. |
| `EligibilityService` | Aplicación | Validar afiliado, jubilado o empleador. | `EmploymentHistoryProvider`. |
| `CatalogoService` | Aplicación | Leer catálogos con cache-aside. | `CatalogoRepository`, `CatalogoCache`. |
| `DocumentoService` | Aplicación | Validar, almacenar y asociar PDF. | `DocumentStorage`, `DocumentoRepository`. |
| `NotificationService` | Aplicación | Encolar y enviar correos. | `NotificacionRepository`, `EmailSender`. |
| `AuditService` | Aplicación | Registrar eventos y cambios. | `AuditoriaRepository`. |
| `TimeCalculator` | Dominio | Calcular años, meses y días. | — |
| `CodeGenerator` | Dominio | Generar código único. | Secuencia PostgreSQL. |
| `PdfGenerator` | Infraestructura | Renderizar el formulario. | Plantilla. |
| `PostgresRepositories` | Infraestructura | Implementar puertos con pgx. | PostgreSQL. |
| `RedisCache` / `RedisAttempts` | Infraestructura | Caché, bloqueo, rate limit, tokens. | Redis. |
| `RegistroCivilAdapter` | Infraestructura | Validar identidad. | HTTP o mock. |
| `IessAdapter` | Infraestructura | Consultar calidad laboral. | HTTP o mock. |
| `SmtpAdapter` | Infraestructura | Enviar correos. | SMTP. |
| `LocalStorage` / `S3Storage` | Infraestructura | Guardar PDF. | Disco / S3. |

---

## 5.2 Puertos principales

### 5.2.1 Repositorios

```go
type Querier interface {
    Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
    Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
    QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TxManager interface {
    WithTx(ctx context.Context, fn func(ctx context.Context, q Querier) error) error
}

type PostulanteRepository interface {
    Create(ctx context.Context, q Querier, p *domain.Postulante) error
    FindByCedula(ctx context.Context, q Querier, cedula string) (*domain.Postulante, error)
    FindByID(ctx context.Context, q Querier, id uuid.UUID) (*domain.Postulante, error)
    Update(ctx context.Context, q Querier, p *domain.Postulante) error
}

type PostulacionRepository interface {
    Create(ctx context.Context, q Querier, p *domain.Postulacion) error
    FindByID(ctx context.Context, q Querier, id uuid.UUID) (*domain.Postulacion, error)
    FindVigente(ctx context.Context, q Querier, postulanteID uuid.UUID, convocatoriaID int) (*domain.Postulacion, error)
    FindForUpdate(ctx context.Context, q Querier, id uuid.UUID) (*domain.Postulacion, error)
    Update(ctx context.Context, q Querier, p *domain.Postulacion) error
}
```

### 5.2.2 Servicios externos

```go
type NationalIDValidator interface {
    ValidateCedula(cedula string) (bool, error)
}

type IdentityProvider interface {
    Validate(ctx context.Context, cedula string, tipo domain.TipoValidacion, valor string) (*domain.IdentityResult, error)
}

type EmploymentHistoryProvider interface {
    GetQuality(ctx context.Context, cedula string, sector domain.SectorCargo) (*domain.EmploymentQuality, error)
}

type PaymentGateway interface {
    AuthorizeAndCapture(ctx context.Context, req domain.PaymentRequest) (*domain.PaymentResult, error)
}

type InvoicingService interface {
    GenerateAndSendInvoice(ctx context.Context, req domain.InvoiceRequest) (*domain.InvoiceResult, error)
}

type EmailSender interface {
    Send(ctx context.Context, msg domain.EmailMessage) error
}

type DocumentStorage interface {
    Save(ctx context.Context, key string, content io.Reader, size int64) error
    Open(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}
```

### 5.2.3 Caché e intentos

```go
type CatalogoCache interface {
    GetPaises(ctx context.Context) ([]domain.Pais, bool)
    SetPaises(ctx context.Context, paises []domain.Pais)
    GetProvincias(ctx context.Context, paisID int) ([]domain.Provincia, bool)
    SetProvincias(ctx context.Context, paisID int, provincias []domain.Provincia)
    GetCantones(ctx context.Context, provinciaID int) ([]domain.Canton, bool)
    SetCantones(ctx context.Context, provinciaID int, cantones []domain.Canton)
    Invalidate(ctx context.Context)
}

type AttemptRepository interface {
    CheckLock(ctx context.Context, scope, cedula string) (time.Duration, error)
    RegisterFailure(ctx context.Context, scope, cedula string) (attempts int, locked bool, retryAfter time.Duration, err error)
    Reset(ctx context.Context, scope, cedula string) error
}
```

---

## 5.3 Composition root

La inyección de dependencias es **manual y explícita** en `cmd/api/main.go`:

```go
func main() {
    cfg := config.MustLoad()
    logger := logger.New(cfg.LogLevel)

    db := postgres.NewPool(cfg.DatabaseURL)
    defer db.Close()
    rdb := redis.NewClient(cfg.Redis)
    defer rdb.Close()

    tx := postgres.NewTxManager(db)

    // Repositorios
    postulanteRepo := postgres.NewPostulanteRepository()
    postulacionRepo := postgres.NewPostulacionRepository()
    catalogoRepo := postgres.NewCatalogoRepository()
    documentoRepo := postgres.NewDocumentoRepository()
    auditoriaRepo := postgres.NewAuditoriaRepository()
    notificacionRepo := postgres.NewNotificacionRepository()

    // Infraestructura e Integraciones (Mocks intercambiables con reales)
    cedulaValidator := validator.NewEcuadorianIDValidator()
    identity := registrocivil.New(cfg.RegistroCivil) // Usa mock o API según env
    employment := iess.New(cfg.IESS)
    mailer := mail.NewSMTP(cfg.SMTP)
    storage := storage.New(cfg.Storage)
    pdf := pdf.NewGenerator(cfg.PDF)
    cache := redis.NewCatalogoCache(rdb)
    attempts := redis.NewAttemptRepository(rdb)

    // Pasarela de pago y facturación (ADR-014: Mockeable pero 100% implementable)
    var paymentGateway domain.PaymentGateway
    if cfg.Env == "production" && cfg.PaymentProvider == "kushki" {
        paymentGateway = payments.NewKushkiAdapter(cfg.PaymentSecret)
    } else {
        paymentGateway = payments.NewMockPaymentGateway(logger)
    }

    var invoicingSvc domain.InvoicingService
    if cfg.Env == "production" && cfg.BillingProvider == "sri" {
        invoicingSvc = billing.NewSriAdapter(cfg.SriConfig)
    } else {
        invoicingSvc = billing.NewMockInvoicingService(mailer, pdf, logger)
    }

    // Servicios
    auditSvc := auditoria.NewService(auditoriaRepo, logger)
    authSvc := auth.NewService(postulanteRepo, identity, cedulaValidator, attempts, security.NewArgon2Hasher(), security.NewTokenService(cfg.JWT), mailer, auditSvc, tx)
    catalogoSvc := catalogo.NewService(catalogoRepo, cache)
    elegibilidadSvc := elegibilidad.NewService(employment)
    documentoSvc := documento.NewService(documentoRepo, storage, cfg.Upload)
    notificacionSvc := notificacion.NewService(notificacionRepo, mailer, logger)
    pagoSvc := pago.NewService(paymentGateway, invoicingSvc, auditSvc, tx)
    postulacionSvc := postulacion.NewService(postulacionRepo, postulanteRepo, catalogoRepo, elegibilidadSvc, documentoSvc, notificacionSvc, pagoSvc, auditSvc, pdf, tx)

    // Handlers
    authHandler := handler.NewAuthHandler(authSvc)
    postulacionHandler := handler.NewPostulacionHandler(postulacionSvc)
    catalogoHandler := handler.NewCatalogoHandler(catalogoSvc)
    documentoHandler := handler.NewDocumentoHandler(documentoSvc)
    pagoHandler := handler.NewPagoHandler(pagoSvc)

    app := router.New(router.Deps{
        Auth: authHandler, Postulacion: postulacionHandler,
        Catalogo: catalogoHandler, Documento: documentoHandler,
        Pago: pagoHandler,
        Tokens: security.NewTokenService(cfg.JWT), Logger: logger,
    })

    logger.Info("api_started", "addr", cfg.HTTPAddr)
    log.Fatal(app.Listen(cfg.HTTPAddr))
}
```

### Por qué manual

- Se ve explícitamente qué depende de qué.
- Facilita la defensa ante el docente.
- No hay contenedores mágicos ni reflexión.
- Si se requiere, se puede migrar a `google/wire` sin cambiar los constructores.

---

## 5.4 Ejemplo de handler

```go
func (h *AuthHandler) ValidateIdentity(c *fiber.Ctx) error {
    var req dto.ValidateIdentityRequest
    if err := c.BodyParser(&req); err != nil {
        return httperr.BadRequest(c, "AUTH_REQUEST_INVALIDO", "Solicitud inválida.")
    }
    if err := h.validator.Struct(req); err != nil {
        return httperr.Validation(c, err)
    }

    result, err := h.svc.ValidateIdentity(c.UserContext(), req.ToCommand())
    if err != nil {
        return httperr.From(c, err)
    }
    return c.Status(fiber.StatusOK).JSON(result)
}
```

**Regla:** el handler no valida reglas de negocio; solo contrato, autenticación y formato.

---

## 5.5 Ejemplo de servicio

```go
func (s *AuthService) ValidateIdentity(ctx context.Context, cmd ValidateIdentityCommand) (*ValidateIdentityResult, error) {
    if retry, err := s.attempts.CheckLock(ctx, "reg", cmd.Cedula); err != nil {
        return nil, err
    } else if retry > 0 {
        return nil, domain.NewLockedError(retry)
    }

    identity, err := s.identity.Validate(ctx, cmd.Cedula, cmd.Tipo, cmd.Valor)
    if err != nil {
        return nil, domain.NewExternalError("registro_civil", err)
    }
    if !identity.Match {
        attempts, locked, retry, _ := s.attempts.RegisterFailure(ctx, "reg", cmd.Cedula)
        _ = s.audit.Record(ctx, domain.AuditEvent{Action: "IDENTITY_VALIDATION_FAILED", ...})
        if locked {
            return nil, domain.NewLockedError(retry)
        }
        return nil, domain.NewInvalidIdentityError(attempts)
    }

    _ = s.attempts.Reset(ctx, "reg", cmd.Cedula)
    token, err := s.tokens.IssueValidationToken(ctx, identity)
    if err != nil {
        return nil, err
    }
    return &ValidateIdentityResult{IdentidadValidada: true, TokenValidacion: token, Datos: identity.ToDTO()}, nil
}
```

### Responsabilidades del servicio

- Aplicar reglas de negocio.
- Orquestar puertos.
- Decidir transacciones.
- Emitir errores de dominio.
- Auditar.
- No conocer Fiber ni SQL.

---

## 5.6 Repositorio y transacción

```go
func (r *PostulacionRepository) Submit(ctx context.Context, id uuid.UUID, fn func(context.Context, Querier, *domain.Postulacion) error) error {
    return r.tx.WithTx(ctx, func(ctx context.Context, q Querier) error {
        p, err := r.FindForUpdate(ctx, q, id)
        if err != nil {
            return err
        }
        if p.Estado != domain.EstadoListaParaEnvio {
            return domain.ErrPostulacionCerrada
        }
        if err := fn(ctx, q, p); err != nil {
            return err
        }
        return r.Update(ctx, q, p)
    })
}
```

- `FindForUpdate` usa `SELECT ... FOR UPDATE`.
- Si `fn` falla, `WithTx` hace `ROLLBACK`.
- El código único se obtiene con `nextval('seq_codigo_postulacion')` dentro de la misma transacción.

---

## 5.7 Middlewares

| Middleware | Función |
|---|---|
| `Recover` | Convierte panic en 500 sin exponer detalles. |
| `RequestID` | Genera/propaga `trace_id`. |
| `Logger` | Registra método, ruta, estado y latencia. |
| `CORS` | Permite solo orígenes configurados. |
| `SecurityHeaders` | CSP, `X-Content-Type-Options`, `X-Frame-Options`. |
| `RateLimit` | Limita por IP y por cédula. |
| `Auth` | Valida JWT, extrae `sub`, `cedula`, `role`, `jti`. |
| `RequireRole` | Autoriza roles (futuro admin). |
| `BodyLimit` | Limita tamaño de request. |

---

## 5.8 Validación

- **DTO:** `go-playground/validator` para campos obligatorios, longitudes, formatos y enums.
- **Dominio:** reglas de negocio (nacionalidad, elegibilidad, estado, fechas, total de documentos).
- **Base de datos:** restricciones e integridad referencial.
- **Frontend:** Zod para feedback inmediato; nunca sustituye al servidor.

### Errores de validación

```json
{
  "code": "VALIDATION_ERROR",
  "status": 422,
  "errors": [
    {"field": "correo", "message": "El correo no es válido."},
    {"field": "password", "message": "Debe tener al menos 10 caracteres."}
  ]
}
```

---

## 5.9 Pruebas y puntos de sustitución

| Puerto | Implementación real | Doble de prueba |
|---|---|---|
| `IdentityProvider` | `RegistroCivilAdapter` | `MockIdentityProvider` |
| `EmploymentHistoryProvider` | `IessAdapter` | `MockEmploymentProvider` |
| `EmailSender` | `SmtpAdapter` | `MemoryEmailSender` |
| `DocumentStorage` | `LocalStorage` / `S3Storage` | `MemoryStorage` |
| `CatalogoCache` | `RedisCatalogoCache` | `MemoryCache` |
| `AttemptRepository` | `RedisAttemptRepository` | `MemoryAttempts` |
| `PostulanteRepository` | `PostgresPostulanteRepository` | `InMemoryPostulanteRepository` |

Esto permite probar servicios sin PostgreSQL, Redis, SMTP ni servicios externos.

---

## 5.10 Secuencias relacionadas

- Registro: `10-diseno-procesos-y-transacciones.md` §10.2.
- Bloqueo: §10.3.
- Login: §10.4.
- Guardado de sección: §10.6.
- Elegibilidad: §10.7.
- Envío: §10.9.
- Caché: §10.10.

# 12. Trazabilidad — SDD

## 12.1 Propósito

Este capítulo conecta los requisitos del SRS con las decisiones de diseño, los componentes, los endpoints, las tablas y las pruebas. Permite demostrar que no hay requisitos “huérfanos” ni componentes sin justificación.

---

## 12.2 Matriz RF → diseño → prueba

| RF | Componente principal | Endpoint / UI | Tablas | Prueba |
|---|---|---|---|---|
| RF-EF11-01 | `LandingPage` | `/` | — | E2E-01 |
| RF-EF11-02 | `AuthService`, `IdentityProvider` | `POST /auth/registro/validar-identidad` | `postulante` | IT-01 |
| RF-EF11-03 | `AuthService`, `PostulanteRepository` | `POST /auth/registro` | `postulante` | IT-02 |
| RF-EF11-04 | `AttemptRepository`, Redis | Validar identidad | — | IT-03 |
| RF-EF11-05 | `AuthService`, `TokenService` | `POST /auth/login` | `postulante` | IT-04 |
| RF-EF11-06 | `AttemptRepository` | `POST /auth/login` | — | IT-05 |
| RF-EF11-07 | `AuthService`, `EmailSender` | `POST /auth/recuperar-password` | `notificacion` | IT-06 |
| RF-EF11-08 | `AuthService` | `POST /auth/restablecer-password` | `postulante` | IT-07 |
| RF-EF11-09 | `AuthMiddleware`, `ProtectedRoute` | `/me`, rutas SPA | — | E2E-07 |
| RF-EF11-10 | `PostulacionService` | `GET /postulaciones/actual` | `postulacion` | E2E-08 |
| RF-EF11-11 | `AuditService` | Eventos | `auditoria_evento` | IT-08 |
| RF-EF12-01 | `PostulacionService` | `POST /postulaciones` | `postulacion` | IT-09 |
| RF-EF12-02 | `PostulacionService`, `SectionForm` | `PATCH/PUT` secciones | `postulacion` | E2E-09 |
| RF-EF12-03 | `PostulacionService` | `PATCH /informacion-personal` | `postulante`, `pareja_postulante` | E2E-10 |
| RF-EF12-04 | `EligibilityService` | `PATCH /informacion-personal` | `postulacion` | IT-10 |
| RF-EF12-05 | `PostulacionService` | `PATCH /informacion-personal` | `pareja_postulante` | IT-11 |
| RF-EF12-06 | `PostulacionService` | `PUT /domicilio` | `domicilio_postulante` | IT-12 |
| RF-EF12-07 | `CatalogoService` | `GET /catalogos/...` | `pais`, `provincia`, `canton` | IT-13 |
| RF-EF12-08 | `FamiliarService` | CRUD `/familiares` | `familiar_institucion` | IT-14 |
| RF-EF12-09 | `PostulacionService` | `PUT /cargo` | `postulacion` | E2E-13 |
| RF-EF12-10 | `EligibilityService` | `PUT /cargo` | `postulacion` | IT-15 |
| RF-EF12-11 | `DocumentoService` | `POST /documentos` | `respaldo_organizacional`, `documento` | IT-16 |
| RF-EF12-12 | `FormacionService` | CRUD `/formacion-academica` | `formacion_academica` | IT-17 |
| RF-EF12-13 | `ExperienciaService`, `TimeCalculator` | CRUD `/experiencia` | `experiencia_profesional` | UT-01 |
| RF-EF12-14 | `PublicacionService` | CRUD `/publicaciones` | `publicacion` | IT-19 |
| RF-EF12-15 | `CapacitacionService` | CRUD `/capacitaciones` | `capacitacion` | IT-20 |
| RF-EF12-16 | `InhabilidadService` | `PUT /inhabilidades` | `inhabilidad_respuesta` | IT-21 |
| RF-EF12-17 | `DocumentoService` | `POST /documentos` | `documento` | IT-22 |
| RF-EF12-18 | `PostulacionService`, `ConsentService` | `POST /guardar` | `consentimiento` | IT-23 |
| RF-EF12-19 | `SubmitService`, `TxManager` | `POST /enviar` | `postulacion`, `documento`, `notificacion` | IT-24 |
| RF-EF12-20 | `PdfGenerator`, `CodeGenerator` | `POST /enviar` | `documento` | IT-25 |
| RF-EF12-21 | `PostulacionService` | `GET /resumen` | `postulacion` | E2E-19 |
| RF-EF12-22 | `PostulacionService` | Escrituras | `postulacion` | IT-26 |
| RF-EF12-23 | `NotificationWorker` | Outbox | `notificacion` | IT-27 |
| RF-EF12-24 | Handlers/services/repos | Toda la API | — | IT-28 |
| RF-EF12-25 | `AuditService` | Eventos | `auditoria_evento` | IT-29 |
| RF-EF12-26 | Servicios de ítems | CRUD | tablas hijas | IT-30 |
| RF-CAT-01 | `CatalogoService` | `GET /catalogos/paises` | `pais` | IT-31 |
| RF-CAT-02 | `CatalogoService` | `GET /paises/{id}/provincias` | `provincia` | IT-32 |
| RF-CAT-03 | `CatalogoService` | `GET /provincias/{id}/cantones` | `canton` | IT-33 |
| RF-CAT-04 | `CatalogoCache` | Redis | — | IT-34 |
| RF-CAT-05 | `CatalogoCache` | Invalidación | — | IT-35 |
| RF-VAL-01 | `NationalIDValidator` (Módulo 10) | `POST /validaciones/cedula` | — | UT-02 |
| RF-PAG-01 | `PagoService`, `PaymentGateway` (Luhn/Mock) | `POST /postulaciones/{id}/pagos` | `transaccion_pago` | IT-36 |
| RF-PAG-02 | `InvoicingService`, `NotificationWorker` | `GET /postulaciones/{id}/factura` | `factura`, `notificacion` | IT-37 |

---

## 12.3 Matriz RNF → diseño → verificación

| RNF | Diseño | Verificación |
|---|---|---|
| RNF-ARQ-01 | Capas + DI manual | Inspección de `main.go` y handlers. |
| RNF-ARQ-02 | SPA por features | Revisión de estructura. |
| RNF-ARQ-03 | Dominio sin infraestructura | Pruebas unitarias. |
| RNF-ARQ-04 | Lint/format | CI. |
| RNF-ARQ-05 | Cobertura | Reporte. |
| RNF-REN-01 | Índices/pool | k6. |
| RNF-REN-02 | Redis | Métricas. |
| RNF-REN-03 | Stateless | Carga 200 usuarios. |
| RNF-DIS-03 | Health checks | `curl`. |
| RNF-DIS-04 | Fallback/fail-closed | Caos Redis. |
| RNF-SEG-02 | Argon2id | Inspección. |
| RNF-SEG-04 | JWT/refresh | Pruebas. |
| RNF-SEG-05 | Redis/Lua | Pruebas. |
| RNF-SEG-06 | Magic bytes/storage | Pruebas. |
| RNF-USA-02 | Componentes accesibles | axe. |
| RNF-COM-01 | Matriz navegadores | Manual. |
| RNF-OBS-01 | Logs JSON | Revisión. |
| RNF-POR-01 | Docker Compose | `up`. |
| RNF-INT-01 | Timeouts/circuit breaker | Mocks. |
| RNF-LEG-01 | Consentimiento | Prueba. |
| RNF-DAT-05 | Transacciones | Rollback. |
| RNF-ESC-01 | Stateless | 2 réplicas. |

### 12.3.1 Cobertura RNF por categoría

La tabla anterior destaca los RNF que gobiernan decisiones de arquitectura. Esta cobertura agrupa todos los RNF del SRS y los vincula con el capítulo del SDD y la verificación prevista.

| Categoría | RNF cubiertos | Diseño / componente | Verificación |
|---|---|---|---|
| Arquitectura y mantenibilidad | `RNF-ARQ-01`, `RNF-ARQ-02`, `RNF-ARQ-03`, `RNF-ARQ-04`, `RNF-ARQ-05`, `RNF-ARQ-06` | Capas, DI manual, SPA por features, ADR. | Inspección; CI; cobertura; revisión de ADR. |
| Rendimiento y capacidad | `RNF-REN-01`, `RNF-REN-02`, `RNF-REN-03`, `RNF-REN-04`, `RNF-REN-05`, `RNF-REN-06` | Índices, pool, Redis, SPA, generador PDF. | k6; Lighthouse; benchmark de envío/PDF. |
| Disponibilidad y continuidad | `RNF-DIS-01`, `RNF-DIS-02`, `RNF-DIS-03`, `RNF-DIS-04`, `RNF-DIS-05` | Health checks, backups, fallback, mocks. | Uptime; restauración; caos; pruebas de fallo. |
| Seguridad | `RNF-SEG-01`, `RNF-SEG-02`, `RNF-SEG-03`, `RNF-SEG-04`, `RNF-SEG-05`, `RNF-SEG-06`, `RNF-SEG-07`, `RNF-SEG-08`, `RNF-SEG-09` | Argon2id, JWT, refresh, rate limit, uploads, RBAC. | ASVS; ZAP; pruebas de auth/autorización; logs; dependencias. |
| Usabilidad y accesibilidad | `RNF-USA-01`, `RNF-USA-02`, `RNF-USA-03`, `RNF-USA-04`, `RNF-USA-05`, `RNF-USA-06` | Layout responsive, formularios, feedback, guards. | axe; teclado; responsive; revisión de textos. |
| Compatibilidad | `RNF-COM-01`, `RNF-COM-02`, `RNF-COM-03` | SPA, visor PDF, breakpoints. | Matriz de navegadores; descarga PDF; móvil/tablet. |
| Observabilidad | `RNF-OBS-01`, `RNF-OBS-02`, `RNF-OBS-03`, `RNF-OBS-04` | Logs JSON, métricas, alertas, `traceId`. | Revisión de logs; `/metrics`; alertas; error con `traceId`. |
| Portabilidad y despliegue | `RNF-POR-01`, `RNF-POR-02`, `RNF-POR-03`, `RNF-POR-04` | Docker Compose, `.env`, migraciones, build estático. | `compose up`; revisión de secretos; `migrate up`; build. |
| Interoperabilidad | `RNF-INT-01`, `RNF-INT-02`, `RNF-INT-03` | Puertos `IdentityProvider`, `EmploymentHistory`, mocks. | Mocks lento/caído; adaptadores; contrato. |
| Legalidad y privacidad | `RNF-LEG-01`, `RNF-LEG-02`, `RNF-LEG-03`, `RNF-LEG-04`, `RNF-LEG-05` | Consentimiento, minimización, TLS, retención, LOPDP. | Envío sin consentimiento; revisión; TLS; retención; procedimiento. |
| Calidad de datos | `RNF-DAT-01`, `RNF-DAT-02`, `RNF-DAT-03`, `RNF-DAT-04`, `RNF-DAT-05` | DDL, validaciones, FK, transacciones. | Duplicados; fechas; integridad; rollback. |
| Escalabilidad | `RNF-ESC-01`, `RNF-ESC-02`, `RNF-ESC-03` | API stateless, servicios externos, `Storage`. | 2 réplicas; revisión; adaptador S3. |
| Restricciones tecnológicas | `RNF-RES-01`, `RNF-RES-02`, `RNF-RES-03`, `RNF-RES-04`, `RNF-RES-05`, `RNF-RES-06`, `RNF-RES-07` | Stack Go/Fiber/React/PostgreSQL/Redis; PDF; zona horaria; español. | Revisión de stack; Compose; límites; configuración. |

---

## 12.4 Cobertura de la rúbrica

| Criterio | Evidencia principal | Documento |
|---|---|---|
| 1. Arquitectura backend e DI | Capas, puertos, `main.go`. | `03`, `05`, `14` |
| 2. Persistencia y transacciones | ER, DDL, `TxManager`, envío. | `04`, `10`, `16` |
| 3. Redis | Claves, Lua, TTL, cache-aside. | `04`, `09`, `10` |
| 4. Frontend SPA | Rutas, estado, componentes. | `06`, `08` |
| 5. API y rutas protegidas | Contratos, guards, JWT. | `06`, `07`, `09` |
| 6. Reglas EF-11/EF-12 | Dominio, validaciones, flujos. | `05`, `09`, `10`; SRS `03`, `07` |

---

## 12.5 Cobertura resumida

| Módulo | RF | Componentes | Endpoints | Pruebas |
|---|---|---|---|---|
| EF-11 | 11 | 6 | 8 | 12 |
| EF-12 | 26 | 12 | 30 | 25 |
| Catálogos | 5 | 2 | 3 | 5 |
| **Total** | **42** | **20** | **41** | **42** |

### Requisitos sin diseño

Ninguno. Todos los RF del SRS tienen al menos un componente, un endpoint o una pantalla asociada.

### Componentes sin requisito

- `AuditService` se justifica por RNF-LEG y RN-33.
- `NotificationWorker` se justifica por RF-EF12-23 y RN-37.
- `PdfGenerator` se justifica por RF-EF12-20.
- `TimeCalculator` se justifica por RF-EF12-13 y RN-23.

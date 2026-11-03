# 08. Trazabilidad y verificación — SRS

## 8.1 Propósito

Este capítulo demuestra que cada requisito tiene un origen, una regla asociada, un componente de diseño y una forma de verificación. La trazabilidad permite responder: **¿de dónde sale este requisito?, ¿dónde se implementa?, ¿cómo sé que se cumplió?**

---

## 8.2 Métodos de verificación

| Método | Cuándo se usa | Evidencia |
|---|---|---|
| **Inspección** | Requisitos de arquitectura, seguridad, documentación y configuración. | Checklist, revisión de código, capturas. |
| **Análisis** | Modelo de datos, índices, consultas, amenazas, capacidad. | Diagramas, EXPLAIN, informe. |
| **Demostración** | Flujos funcionales ante el docente. | Guion de sustentación, video. |
| **Prueba** | Comportamiento observable y medible. | Reporte de pruebas, cobertura, logs. |

### Niveles de prueba

| Nivel | Alcance | Herramientas sugeridas |
|---|---|---|
| Unitaria | Servicios, validadores, cálculos. | `go test`, Vitest. |
| Integración | Repositorios, transacciones, Redis, API. | Testcontainers, `httptest`. |
| E2E | Flujos completos de usuario. | Playwright. |
| Carga | Rendimiento y concurrencia. | k6. |
| Seguridad | Autenticación, autorización, uploads. | OWASP ZAP, pruebas manuales. |
| Accesibilidad | Formularios y navegación. | axe, pruebas con teclado. |

---

## 8.3 Matriz de trazabilidad funcional

| RF | Regla(s) | Componente SDD | Endpoint / pantalla | Prueba |
|---|---|---|---|---|
| RF-EF11-01 | RN-11 | `AuthHandler`, `LandingPage` | `GET /`, UI-01 | E2E-01 |
| RF-EF11-02 | RN-02, RN-06 | `AuthService`, `IdentityProvider` | `POST /auth/registro/validar-identidad` | IT-01, E2E-02 |
| RF-EF11-03 | RN-03, RN-04, RN-05 | `AuthService`, `PostulanteRepository` | `POST /auth/registro` | IT-02, E2E-03 |
| RF-EF11-04 | RN-06, RN-08 | `AttemptService`, Redis | `POST /auth/registro/validar-identidad` | IT-03, E2E-04 |
| RF-EF11-05 | RN-05, RN-07 | `AuthService`, `SessionService` | `POST /auth/login` | IT-04, E2E-05 |
| RF-EF11-06 | RN-07, RN-08 | `AttemptService`, Redis | `POST /auth/login` | IT-05 |
| RF-EF11-07 | RN-02, RN-09 | `PasswordRecoveryService`, `Mailer` | `POST /auth/recuperar-password` | IT-06, E2E-06 |
| RF-EF11-08 | RN-05, RN-09 | `PasswordRecoveryService` | `POST /auth/restablecer-password` | IT-07 |
| RF-EF11-09 | RN-10 | `AuthMiddleware`, `ProtectedRoute` | `/me`, rutas SPA | E2E-07 |
| RF-EF11-10 | RN-11, RN-12 | `PostulacionService`, router | `GET /postulaciones/actual` | E2E-08 |
| RF-EF11-11 | RN-33 | `AuditService` | Eventos internos | IT-08 |
| RF-EF12-01 | RN-11, RN-12, RN-13 | `PostulacionService` | `POST /postulaciones` | IT-09 |
| RF-EF12-02 | RN-14, RN-36 | `PostulacionService`, `SectionForm` | `PATCH /postulaciones/{id}/...` | E2E-09 |
| RF-EF12-03 | RN-02, RN-15 | `PostulacionService`, `PersonalInfoForm` | `PATCH /informacion-personal` | E2E-10 |
| RF-EF12-04 | RN-01 | `EligibilityService` | `PATCH /informacion-personal` | IT-10, E2E-11 |
| RF-EF12-05 | RN-15 | `PostulacionService` | `PATCH /informacion-personal` | IT-11 |
| RF-EF12-06 | RN-16, RN-17 | `PostulacionService` | `PUT /domicilio` | IT-12 |
| RF-EF12-07 | RN-16, RN-39 | `CatalogoService`, `CascadingSelect` | `GET /catalogos/...` | IT-13, E2E-12 |
| RF-EF12-08 | RN-18 | `FamiliarService` | CRUD `/familiares` | IT-14 |
| RF-EF12-09 | RN-19 | `PostulacionService` | `PUT /cargo` | E2E-13 |
| RF-EF12-10 | RN-20 | `EligibilityService`, `IessProvider` | `PUT /cargo` | IT-15, E2E-14 |
| RF-EF12-11 | RN-21, RN-27 | `DocumentService` | `POST /documentos` | IT-16 |
| RF-EF12-12 | RN-22, RN-27 | `FormacionService`, `DocumentService` | CRUD `/formacion-academica` | IT-17 |
| RF-EF12-13 | RN-23, RN-27 | `ExperienciaService`, `TimeCalculator` | CRUD `/experiencia` | UT-01, IT-18 |
| RF-EF12-14 | RN-24, RN-27 | `PublicacionService` | CRUD `/publicaciones` | IT-19 |
| RF-EF12-15 | RN-25, RN-27 | `CapacitacionService` | CRUD `/capacitaciones` | IT-20 |
| RF-EF12-16 | RN-26 | `InhabilidadService` | `PUT /inhabilidades` | IT-21, E2E-15 |
| RF-EF12-17 | RN-27, RN-28 | `DocumentService`, `Storage` | `POST /documentos` | IT-22, E2E-16 |
| RF-EF12-18 | RN-29, RN-30 | `PostulacionService`, `ConsentService` | `POST /guardar` | IT-23, E2E-17 |
| RF-EF12-19 | RN-30, RN-31, RN-35 | `SubmitService`, `TxManager` | `POST /enviar` | IT-24, E2E-18 |
| RF-EF12-20 | RN-32, RN-33 | `PdfGenerator`, `CodeGenerator` | `POST /enviar` | IT-25 |
| RF-EF12-21 | RN-31 | `PostulacionService` | `GET /resumen` | E2E-19 |
| RF-EF12-22 | RN-31 | `PostulacionService` | Escrituras posteriores | IT-26 |
| RF-EF12-23 | RN-32, RN-37 | `NotificationWorker`, `Mailer` | Outbox | IT-27 |
| RF-EF12-24 | Todas | Handlers, services, repos | Toda la API | IT-28 |
| RF-EF12-25 | RN-33 | `AuditService` | Eventos internos | IT-29 |
| RF-EF12-26 | RN-31, RN-33, RN-38 | Servicios de ítems | CRUD de ítems | IT-30 |
| RF-CAT-01 | RN-39 | `CatalogoService` | `GET /catalogos/paises` | IT-31 |
| RF-CAT-02 | RN-39 | `CatalogoService` | `GET /paises/{id}/provincias` | IT-32 |
| RF-CAT-03 | RN-39 | `CatalogoService` | `GET /provincias/{id}/cantones` | IT-33 |
| RF-CAT-04 | RN-39 | `CatalogoCache` | Redis | IT-34 |
| RF-CAT-05 | RN-39 | `CatalogoCache` | Invalidación | IT-35 |
| RF-VAL-01 | RN-01, RN-41 | `NationalIDValidator` | `POST /validaciones/cedula` | UT-02 |
| RF-PAG-01 (retirado) | RN-42 | `PagoService`, `PaymentGateway` | `POST /postulaciones/{id}/pagos` | IT-36 |
| RF-PAG-02 (retirado) | RN-43 | `InvoicingService`, `NotificationWorker` | `GET /postulaciones/{id}/factura` | IT-37 |

**Convención de pruebas:** `UT` unitaria, `IT` integración, `E2E` extremo a extremo. Los identificadores se desarrollan en `SDD/15-plan-de-pruebas.md`.

---

## 8.4 Matriz de trazabilidad no funcional

| RNF | Decisión de diseño | Verificación |
|---|---|---|
| RNF-ARQ-01 | Capas handler/service/repository e inyección manual. | Inspección de código y diagrama de componentes. |
| RNF-ARQ-02 | SPA modular por `features`. | Revisión de estructura y lint. |
| RNF-ARQ-03 | Dominio sin dependencias de infraestructura. | Pruebas unitarias sin DB/Redis. |
| RNF-REN-01 | Índices, pool de conexiones, consultas optimizadas. | k6 y `EXPLAIN ANALYZE`. |
| RNF-REN-02 | Cache-aside en Redis. | Medición de latencia. |
| RNF-REN-03 | API stateless y réplicas. | Prueba de carga con 200 usuarios. |
| RNF-DIS-03 | Endpoints de salud. | `curl /health/ready`. |
| RNF-DIS-04 | Fallback de catálogos y fail-closed de auth. | Prueba de caos con Redis detenido. |
| RNF-SEG-02 | Argon2id. | Inspección de hash. |
| RNF-SEG-04 | JWT corto + refresh rotativo. | Prueba de expiración y revocación. |
| RNF-SEG-05 | Redis + rate limiting. | Prueba de 3 intentos y 429. |
| RNF-SEG-06 | Validación de magic bytes y almacenamiento privado. | Prueba con PDF falso. |
| RNF-USA-02 | Componentes accesibles y foco visible. | axe + prueba de teclado. |
| RNF-COM-01 | Matriz de navegadores. | Prueba manual en 4 navegadores. |
| RNF-OBS-01 | Logs JSON con `trace_id`. | Revisión de logs. |
| RNF-POR-01 | Docker Compose. | `docker compose up` en máquina limpia. |
| RNF-INT-01 | Timeouts y circuit breaker. | Mock lento y caído. |
| RNF-LEG-01 | Consentimiento versionado. | Envío sin aceptar. |
| RNF-DAT-05 | Transacciones. | Prueba de rollback. |
| RNF-ESC-01 | Stateless y servicios externos. | Prueba con 2 réplicas. |

### 8.4.1 Cobertura por categoría

La tabla anterior prioriza los RNF con mayor impacto. Esta cobertura agrupa **todos** los RNF del catálogo para asegurar que ninguno quede sin verificación.

| Categoría | RNF cubiertos | Evidencia de diseño (SDD) | Verificación |
|---|---|---|---|
| Arquitectura y mantenibilidad | `RNF-ARQ-01`, `RNF-ARQ-02`, `RNF-ARQ-03`, `RNF-ARQ-04`, `RNF-ARQ-05`, `RNF-ARQ-06` | `03`, `05`, `06`, `14` | Revisión de capas y DI; ADR; CI; reporte de cobertura. |
| Rendimiento y capacidad | `RNF-REN-01`, `RNF-REN-02`, `RNF-REN-03`, `RNF-REN-04`, `RNF-REN-05`, `RNF-REN-06` | `04`, `11`, `15` | k6; Lighthouse; benchmark de envío y PDF. |
| Disponibilidad y continuidad | `RNF-DIS-01`, `RNF-DIS-02`, `RNF-DIS-03`, `RNF-DIS-04`, `RNF-DIS-05` | `11`, `13` | Uptime; prueba de restauración; health checks; caos de Redis; mocks de fallo. |
| Seguridad | `RNF-SEG-01`, `RNF-SEG-02`, `RNF-SEG-03`, `RNF-SEG-04`, `RNF-SEG-05`, `RNF-SEG-06`, `RNF-SEG-07`, `RNF-SEG-08`, `RNF-SEG-09` | `09` | ASVS; ZAP; pruebas de auth, autorización y uploads; revisión de logs; escaneo de dependencias. |
| Usabilidad y accesibilidad | `RNF-USA-01`, `RNF-USA-02`, `RNF-USA-03`, `RNF-USA-04`, `RNF-USA-05`, `RNF-USA-06` | `08` | Pruebas responsive; axe; navegación por teclado; revisión de textos. |
| Compatibilidad | `RNF-COM-01`, `RNF-COM-02`, `RNF-COM-03` | `06`, `08` | Matriz de navegadores; descarga de PDF; pruebas móviles y tablet. |
| Observabilidad | `RNF-OBS-01`, `RNF-OBS-02`, `RNF-OBS-03`, `RNF-OBS-04` | `11`, `13` | Logs JSON; métricas; alertas; `traceId` en errores. |
| Portabilidad y despliegue | `RNF-POR-01`, `RNF-POR-02`, `RNF-POR-03`, `RNF-POR-04` | `13` | `docker compose up`; revisión de `.env.example`; `migrate up`; build estático. |
| Interoperabilidad | `RNF-INT-01`, `RNF-INT-02`, `RNF-INT-03` | `05`, `09` | Mocks lento/caído; adaptadores sustituibles; pruebas de contrato. |
| Legalidad y privacidad | `RNF-LEG-01`, `RNF-LEG-02`, `RNF-LEG-03`, `RNF-LEG-04`, `RNF-LEG-05` | `09` | Envío sin consentimiento; revisión de datos; TLS; retención; procedimiento LOPDP. |
| Calidad de datos | `RNF-DAT-01`, `RNF-DAT-02`, `RNF-DAT-03`, `RNF-DAT-04`, `RNF-DAT-05` | `04`, `10` | Duplicados; fechas borde; integridad referencial; rollback. |
| Escalabilidad | `RNF-ESC-01`, `RNF-ESC-02`, `RNF-ESC-03` | `03`, `11` | Prueba con 2 réplicas; servicios externos; adaptador de storage alternativo. |
| Restricciones tecnológicas | `RNF-RES-01`, `RNF-RES-02`, `RNF-RES-03`, `RNF-RES-04`, `RNF-RES-05`, `RNF-RES-06`, `RNF-RES-07` | `03`, `13` | Revisión de stack; Compose; límites PDF; zona horaria; idioma. |

---

## 8.5 Cobertura de la rúbrica

| Criterio | Requisitos que lo evidencian | Documento |
|---|---|---|
| Arquitectura backend e inyección de dependencias | RNF-ARQ-01, RNF-ARQ-03; RF-EF11/12 | `SDD/03`, `SDD/05`, `SDD/14` |
| Persistencia relacional y transacciones | RNF-DAT-05; RF-EF12-19 | `SDD/04`, `SDD/10` |
| Redis: bloqueo, TTL y caché | RNF-SEG-05, RNF-REN-02, RNF-DIS-04; RF-EF11-04, RF-CAT-04 | `SDD/04`, `SDD/09`, `SDD/10` |
| Frontend SPA reactiva y modular | RNF-ARQ-02, RNF-USA-01/02; RF-EF12-02 | `SDD/06`, `SDD/08` |
| Consumo de API y rutas protegidas | RNF-SEG-08; RF-EF11-09 | `SDD/06`, `SDD/07` |
| Reglas de negocio EF-11 y EF-12 | RN-01 a RN-40; RF-EF11/12 | `SRS/03`, `SRS/07`, `SRS/09` |

---

## 8.6 Gestión de cambios y estado

| Estado | Significado |
|---|---|
| **Propuesto** | Requisito identificado, no aprobado. |
| **Aprobado** | Incluido en la línea base. |
| **Implementado** | Existe código y prueba. |
| **Verificado** | Prueba ejecutada y evidencia adjunta. |
| **Diferido** | Se posterga a otra fase. |

### Reglas de cambio

1. Todo cambio de alcance se registra en el control de versiones.
2. Se actualizan los RF, RN, RNF y la matriz.
3. Se revisa el impacto en el SDD y en las pruebas.
4. No se acepta un requisito sin criterio de verificación.
5. Los cambios posteriores al cierre del parcial se documentan como deuda técnica.

---

## 8.7 Evidencias para la sustentación

| Evidencia | Dónde se demuestra |
|---|---|
| Diagrama de arquitectura y capas | `SDD/03` |
| Inyección de dependencias en `main.go` | `SDD/05` |
| Modelo entidad-relación y DDL | `SDD/04`, `SDD/16` |
| Transacción de envío | `SDD/10` |
| Bloqueo con Redis y TTL | `SDD/09`, `SDD/10` |
| Caché de catálogos | `SDD/04`, `SDD/10` |
| Formulario SPA y estados | `SDD/06`, `SDD/08` |
| Rutas protegidas | `SDD/06`, `SDD/07` |
| Pruebas y cobertura | `SDD/15` |
| Despliegue local | `SDD/13` |

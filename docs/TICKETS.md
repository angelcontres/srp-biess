# Catálogo de Tickets de Implementación (Work Items / Backlog) — SRP-BIESS

| Metadato | Detalle |
|---|---|
| **Proyecto** | Sistema de Registro y Postulación al Directorio del BIESS |
| **Metodología** | Scrum / Kanban (Sprints Ágiles) |
| **Integrantes y Asignados** | • **Paulo Rafael Orrala Arriaga** (Backend Lead & DI)<br>• **Angel Alejandro Villon Panimboza** (Data & Persistence Lead)<br>• **Carlos Fernando Patiño Garcia** (Frontend SPA Lead)<br>• **Jose Daniel Gonzalez Cochea** (Cache, Integraciones & QA) |
| **Versión** | 1.0 |
| **Formato** | Compatible con GitHub Issues, Jira, Plane, GitLab y Trello |

---

## Índice de Épicas

* [Épica 1: Arquitectura Backend & Inyección de Dependencias (Go)](#épica-1-arquitectura-backend--inyección-de-dependencias-go)
* [Épica 2: Persistencia Relacional & Transacciones Atómicas (PostgreSQL)](#épica-2-persistencia-relacional--transacciones-atómicas-postgresql)
* [Épica 3: Caché, Expiración & Control de Intentos (Redis)](#épica-3-caché-expiración--control-de-intentos-redis)
* [Épica 4: Interoperabilidad, Mocks, Pagos y Facturación (SRI / Módulo 10 / Luhn)](#épica-4-interoperabilidad-mocks-pagos-y-facturación)
* [Épica 5: Frontend SPA Modular & Formularios Reactivos (React + Tailwind)](#épica-5-frontend-spa-modular--formularios-reactivos)
* [Épica 6: Consumo de APIs, Rutas Protegidas & Cascada Geográfica](#épica-6-consumo-de-apis-rutas-protegidas--cascada-geográfica)

---

## Épica 1: Arquitectura Backend & Inyección de Dependencias (Go)
**Responsable Principal:** Paulo Rafael Orrala Arriaga  
**Ponderación Rúbrica:** 8.0 Puntos

### Ticket `SRP-001`: Inicialización del Proyecto Go y Estructura en Capas Limpias
* **Tipo:** Task / Setup
* **Asignado:** Paulo Rafael Orrala Arriaga
* **Rama:** `feature/backend-clean-arch`
* **Descripción:** Configurar el módulo Go (`go.mod`) con Fiber v2 y establecer el árbol de paquetes desacoplado conforme a `Docs/SDD/03-arquitectura-de-software.md`:
  * `cmd/api/main.go` (Composition Root)
  * `internal/domain/` (Entidades y puertos)
  * `internal/application/` (Servicios de casos de uso)
  * `internal/infrastructure/` (Implementaciones de repositorios y adaptadores)
  * `internal/interfaces/http/` (Handlers, DTOs y Router)
* **Criterios de Aceptación:**
  - [x] El proyecto compila con `go build ./cmd/api`.
  - [x] Ningún paquete de dominio importa dependencias de infraestructura o base de datos.
  - [x] Middleware global de recover, logger estructurado JSON y CORS configurados.

---

### Ticket `SRP-002`: Implementación del Composition Root con Inyección Manual
* **Tipo:** Feature
* **Asignado:** Paulo Rafael Orrala Arriaga
* **Rama:** eature/composition-root-di
* **Descripción:** Implementar en `cmd/api/main.go` la instanciación e inyección manual de dependencias (`Handler` $\to$ `Service` $\to$ `Repository`) sin magia de frameworks ni contenedores de reflexión (ADR-003).
* **Criterios de Aceptación:**
  - [x] Cada constructor recibe interfaces explícitas.
  - [x] La inicialización de la conexión de PostgreSQL (`pgxpool`) y Redis se inyecta desde `main.go`.
  - [x] Facilidad para sustentar el flujo de dependencias en menos de 2 minutos ante el docente.

---

### Ticket `SRP-003`: Middleware de Autenticación JWT y Protección de Rutas
* **Tipo:** Feature
* **Asignado:** Paulo Rafael Orrala Arriaga
* **Rama:** `feature/auth-jwt-di`
* **Descripción:** Implementar el servicio de tokens con Argon2id para contraseñas, Access Token JWT (15 min) y verificación de Refresh Token opaco en cookie `HttpOnly` (ADR-006).
* **Criterios de Aceptación:**
  - [x] Middleware valida cabecera `Authorization: Bearer <token>`.
  - [x] Peticiones sin token o expiradas devuelven HTTP 401 estructurado.
  - [x] Inyección de contexto del usuario autenticado (`postulante_id`) en cada request.

---

### Ticket `SRP-018`: El middleware JWT del grupo `/auth` convierte rutas inexistentes en 401
* **Tipo:** Bug / Backend
* **Asignado:** Paulo Rafael Orrala Arriaga
* **Issue:** [#35](https://github.com/angelcontres/srp-biess/issues/35)
* **Rama:** `fix/auth-jwt-route-scope`
* **Origen:** pendiente de SRP-003 (issue cerrado). Detectado en SRP-012: con un backend sin la ruta `/auth/registro/validar-identidad`, el registro respondía 401 "Cabecera Authorization no proporcionada" en lugar de 404, lo que ocultó la causa real.
* **Descripción:** en `internal/interfaces/http/router.go`, `auth.Group("", middleware.JWTProtected(...))` aplica el middleware a todo `/api/v1/auth/*`. Registrar el middleware por ruta (`auth.Get("/me", middleware.JWTProtected(cfg.AuthService), cfg.AuthHandler.Me)`, ídem `/logout`), como ya se hace con el grupo con prefijo propio de `/postulaciones`.
* **Criterios de Aceptación:**
  - [ ] Una ruta inexistente bajo `/api/v1/auth/` responde 404.
  - [ ] `GET /auth/me` y `POST /auth/logout` siguen respondiendo 401 sin token.
  - [ ] Prueba de handler que cubre ambos casos.
* **Nota de revisión:** `GET /auth/logout` daría 405 después del arreglo, por eso el método se fija en el criterio. `GET /postulaciones/no-existe` también responde 401 hoy (coincide con `/:id`), así que el comentario de `router.go` que lo usa como ejemplo de 404 es incorrecto.

---

### Ticket `SRP-020`: Política CORS configurable por entorno y `Retry-After` expuesto
* **Tipo:** Bug / Backend
* **Asignado:** Paulo Rafael Orrala Arriaga
* **Issue:** [#36](https://github.com/angelcontres/srp-biess/issues/36)
* **Rama:** `fix/cors-policy`
* **Origen:** pendiente de SRP-001 (issue cerrado, criterio "CORS configurado").
* **Descripción:** en `internal/interfaces/http/middleware/cors.go`:
  * Leer los orígenes desde una variable de entorno (`CORS_ALLOWED_ORIGINS`, SDD 16.4) en lugar de fijarlos en código, definiendo el comportamiento cuando la variable está vacía.
  * Exponer `Retry-After` (`ExposeHeaders`) para que la SPA lea la duración del bloqueo 423 cuando se sirve desde otro origen (SRP-006, SRP-012).
  * Completar las pruebas: origen permitido con credenciales, origen no permitido sin cabecera, y preflight `OPTIONS`.
* **Criterios de Aceptación:**
  - [ ] `go test ./internal/interfaces/http/middleware` pasa.
  - [ ] Una petición desde un origen permitido recibe `Access-Control-Allow-Origin` con ese origen y `Access-Control-Allow-Credentials: true`; un origen no permitido no lo recibe.
  - [ ] `Access-Control-Expose-Headers` incluye `Retry-After`.
  - [ ] Cambiar `CORS_ALLOWED_ORIGINS` cambia la política sin recompilar.
* **Nota de revisión:** dos premisas del ticket original ya estaban resueltas en HEAD y por eso se quitaron: la prueba no espera `*` (`middleware_test.go:43` compara contra `http://localhost:5173`) y el origen ya dice `:3000` (`cors.go:10`). `go test ./internal/interfaces/http/middleware` ya pasa hoy.

---

## Épica 2: Persistencia Relacional & Transacciones Atómicas (PostgreSQL)
**Responsable Principal:** Angel Alejandro Villon Panimboza  
**Ponderación Rúbrica:** 8.0 Puntos

### Ticket `SRP-004`: Configuración del Pool de Conexiones y Repositorios con `pgx/v5`
* **Tipo:** Feature
* **Asignado:** Angel Alejandro Villon Panimboza
* **Rama:** `feature/postgres-transactions-ddl`
* **Descripción:** Crear capa de acceso a datos utilizando `pgxpool` implementando los repositorios para `postulante`, `postulacion`, `catalogo` y `documento`.
* **Criterios de Aceptación:**
  - [x] Conexiones concurrentes controladas mediante parámetros de entorno.
  - [x] Repositorios utilizan interfaces que aceptan tanto el pool como una transacción activa (`Querier`).

---

### Ticket `SRP-005`: Gestor Transaccional Atómico (`TxManager`) para Envío de Postulación
* **Tipo:** Feature / Crítico
* **Asignado:** Angel Alejandro Villon Panimboza
* **Rama:** `feature/atomic-tx-manager`
* **Descripción:** Diseñar e implementar el componente `TxManager` con soporte de `BEGIN`, `COMMIT` y `ROLLBACK` automático para operaciones multi-tabla (ADR-004).
* **Criterios de Aceptación:**
  - [x] El guardado y envío de postulación actualiza el estado, genera código único, crea el PDF y encola la notificación en una sola transacción atómica.
  - [x] Si falla la inserción del documento o del correo, se ejecuta `ROLLBACK` total y la base queda intacta.
  - [x] Prueba de integración verificando el rollback forzado.

---

### Ticket `SRP-017`: API REST de Postulación — Cargo, Documentos, Envío y Persistencia Normalizada
* **Tipo:** Feature / Backend
* **Asignado:** Angel Alejandro Villon Panimboza
* **Issue:** [#38](https://github.com/angelcontres/srp-biess/issues/38)
* **Rama:** `feature/postulacion-api`
* **Origen:** detectado al cerrar SRP-013. Ningún ticket asignaba estos endpoints de SDD 7.3.4, y el formulario los necesita para funcionar de extremo a extremo con datos reales. La base ya existe: los endpoints de borrador y guardado por sección (ver SRP-013) y los catálogos (ver SRP-014).
* **Dependencias:** SRP-005 (`TxManager`) para el envío; SRP-009 (mock de Historial Laboral IESS) para la elegibilidad.
* **Descripción:** Exponer, protegidos con JWT y respetando las capas de ADR-003, los cinco bloques del ticket. Se dividió en sub-issues (issue paraguas [#38](https://github.com/angelcontres/srp-biess/issues/38)); el alcance y los criterios de cada uno viven ahí:

  | Parte | Alcance | Depende de |
  |---|---|---|
  | [#39](https://github.com/angelcontres/srp-biess/issues/39) SRP-017a | `PUT /{id}/cargo` con elegibilidad y contrato `{ postulacion, elegibilidad }` para la SPA | — |
  | [#40](https://github.com/angelcontres/srp-biess/issues/40) SRP-017b | Documentos: PDF real, cuotas, UUID, SHA-256 y URL firmada | — |
  | [#43](https://github.com/angelcontres/srp-biess/issues/43) SRP-017c | Envío atómico e idempotente con `TxManager` | #39, #42 |
  | [#41](https://github.com/angelcontres/srp-biess/issues/41) SRP-017d | Validación de secciones en el servidor antes de `COMPLETA` | — |
  | [#42](https://github.com/angelcontres/srp-biess/issues/42) SRP-017e | Normalización de `postulacion_seccion` a las tablas de SDD 4.6 | #41 |

  Orden sugerido: #39 → #40 y #41 en paralelo → #42 → #43. Cada parte va en su propia rama para que el PR quede por debajo de 400 líneas.
* **Criterios de Aceptación (del paraguas):**
  - [ ] Las cinco partes cerradas.
  - [ ] El recorrido completo (secciones 1 a 10, adjuntos y envío) funciona con `VITE_POSTULACION_MOCK` apagado.
  - [ ] `go build ./cmd/api` y `go test ./...` en verde, con las pruebas de integración contra PostgreSQL incluidas.
* **Decisiones tomadas** (el detalle vive en cada sub-issue):
  * **#39 Cargo:** la respuesta sigue entregando el objeto `elegibilidad` derivado que la SPA consume.
  * **#40 Documentos:** carga y borrado solo en `BORRADOR` (otro estado responde 409); `FORMULARIO_POSTULACION` se rechaza con 400 porque lo genera el envío; firma `%PDF-` o 415; 4 MB respaldo, 2 MB resto y 50 MB acumulado, todo con 413. La base garantiza solo 4 MB (`ck_documento_tamano`, sin migración); el 2 MB y el acumulado son reglas de aplicación, y un rechazo del `CHECK` (`23514`) se traduce a 413, nunca 500. Cuota con `SELECT … FOR UPDATE` sobre la postulación. Descarga con doble validación, firma HMAC (`DOCUMENTOS_URL_SECRETO`, 5 min) y JWT: sin JWT 401, firma vencida o manipulada 403. La respuesta incluye `fechaCarga` desde `created_at`.
  * **#41 Validación:** las reglas se definen por sección en una fuente única de la que derivan Go y la SPA. Pendiente: el formato de esa fuente.
  * **#42 Normalización:** al enviar; el JSONB es la fuente durante el borrador y el volcado corre dentro de la transacción de #43.
  * **#43 Envío:** tabla nueva para la `Idempotency-Key`, sin confiar en el estado. Pendiente: resolver el envío concurrente con la misma llave (`FOR UPDATE`, `UNIQUE` o ambos).
* **Nota de revisión:** el mapeo de secciones no es 1:1: `PUT /{id}/respaldo-organizacional` guarda la sección `cargo`, que contiene los datos de respaldo organizacional.

---

### Ticket `SRP-019`: Catálogo Geográfico Completo con Códigos Oficiales del INEC
* **Tipo:** Data / Seed
* **Asignado:** Angel Alejandro Villon Panimboza
* **Issue:** [#37](https://github.com/angelcontres/srp-biess/issues/37)
* **Rama:** `feature/catalogo-geografico-inec`
* **Origen:** pendiente de SRP-004 (issue cerrado), detectado en SRP-014.
* **Descripción:** `db/init.sql` trae 194 países, las 24 provincias del Ecuador con códigos INEC y 199 cantones asignados a su provincia, pero:
  * el sufijo del código de cantón es secuencial dentro de la provincia, no el código oficial (Azuay carga 8 de sus 15 cantones y asigna `0103` a Chordeleg);
  * faltan cantones en varias provincias (entre otros Azuay, Guayas, Loja, Morona Santiago, Chimborazo, Los Ríos y Santa Elena);
  * el `CHECK ck_domicilio_pais_ec` de `domicilio_postulante` depende del id de Ecuador en la semilla (`52`).
* **Criterios de Aceptación:**
  - [x] Los cantones vigentes con su código oficial del INEC, cargados desde una fuente citada en el script con fecha de corte.
  - [x] El script sigue siendo idempotente (se puede ejecutar dos veces sin errores).
  - [x] La restricción de domicilio identifica a Ecuador sin depender de un id fijo (por ejemplo, con un trigger que compare `codigo_iso2 = 'EC'`).
  - [x] Se invalida la caché de catálogos incrementando `srp:dev:cat:version` (SRP-007). Es un paso del despliegue (ver abajo), no del script SQL.
* **Notas de implementación (integración autorizada por el responsable del frontend):**
  * **Fuente:** INEC, Clasificador Geográfico Estadístico, códigos DPA 2026 (`cdpa2026.xlsx`, hoja CANTONES, publicado el 27/02/2026, descargado el 03/11/2026 de `aplicaciones2.ecuadorencifras.gob.ec/SIN`). 222 cantones en las 24 provincias; se excluye la "Zona en estudio" (código 90, 2 cantones), que no pertenece a ninguna provincia. Los nombres se pasan de mayúsculas a mayúscula inicial (DESIGN.md prohíbe mayúsculas sostenidas); Quito figura con su nombre oficial, "Distrito Metropolitano de Quito".
  * El ejemplo del issue #37 tampoco coincidía con el INEC: `0102` es Girón, `0103` Gualaceo, `0105` Paute y `0111` Chordeleg. Hay huecos oficiales en la numeración (Guayas no tiene `0915`, `0917` ni `0926`; Napo no tiene `1502`, `1505`, `1506` ni `1508`; Pichincha no tiene `1706`).
  * **Idempotencia:** el id del cantón es su código sin el cero inicial (`0101` → 101) y la carga hace `ON CONFLICT (id) DO UPDATE`, así que un cambio de nombre o código sí se aplica. Los cantones del Ecuador que no están en la fuente se desactivan (`activo = false`), no se borran, para no romper domicilios. En bases creadas con la semilla anterior (ids que no coincidían con el código) se libera primero su código; esas bases quedan corregidas, pero los ids viejos que tuvieran los borradores pasan a otro cantón, así que conviene recrearlas.
  * **Domicilio:** el `CHECK ck_domicilio_pais_ec` se reemplazó por el trigger `trg_domicilio_valida_ubicacion` (`INSERT` y `UPDATE`), que reconoce a Ecuador por `codigo_iso2 = 'EC'` y además verifica que el cantón pertenezca a la provincia y la provincia al país. En una base existente lo aplica `db/migrations/003_domicilio_ecuador_por_iso.sql`.
  * **Pruebas:** `repositories_test.go` busca a Ecuador por `EC` (antes usaba el id 1) y verifica los 222 cantones, códigos conocidos y que los códigos inexistentes no estén activos. Las pruebas se omiten si PostgreSQL no está disponible.
  * **Al desplegar:** recrear la base o volver a ejecutar `db/init.sql`, y luego invalidar la caché: `docker exec srp-biess-redis redis-cli INCR srp:dev:cat:version`.
* **Nota de revisión:** el `ON CONFLICT (provincia_id, codigo) DO NOTHING` con ids explícitos no actualiza las filas cuyo código cambia: el conflicto aparece por la clave primaria `id` y el catálogo queda corrido aunque el script "pase dos veces". Ver el issue #37.

---

## Épica 3: Caché, Expiración & Control de Intentos (Redis)
**Responsable Principal:** Jose Daniel Gonzalez Cochea  
**Ponderación Rúbrica:** 6.0 Puntos

### Ticket `SRP-006`: Bloqueo Temporal de 10 Minutos tras 3 Intentos Fallidos (EF-11)
* **Tipo:** Feature / Seguridad
* **Asignado:** Jose Daniel Gonzalez Cochea
* **Rama:** `feature/redis-cache-lock`
* **Descripción:** Implementar el control de intentos en el registro e inicio de sesión mediante Redis y un script Lua atómico:
  * Llave: `srp:dev:auth:reg:attempts:{cedula}` con TTL de 600 segundos (10 min).
  * Al 3er intento fallido consecutivo: bloqueo automático y respuesta HTTP 423 con `Retry-After: 600`.
* **Criterios de Aceptación:**
  - [ ] Al 3er intento fallido consecutivo el sistema bloquea y responde con mensaje exacto del RF: *"registro fallido, inténtelo nuevamente en 10 minutos"*.
  - [ ] Transcurridos los 10 minutos (TTL expira), el contador se reinicia automáticamente sin intervención manual.
  - [ ] Demostrable en vivo mediante `redis-cli TTL`.

---

### Ticket `SRP-007`: Caching con Estrategia Cache-Aside y TTL para Catálogos Geográficos
* **Tipo:** Performance / Cache
* **Asignado:** Jose Daniel Gonzalez Cochea
* **Rama:** `feature/redis-cache-lock`
* **Descripción:** Implementar capa de caché en Redis para países, provincias y cantones con TTL de 24 horas y soporte de fallback resiliente a PostgreSQL si Redis no responde (ADR-005).
* **Criterios de Aceptación:**
  - [ ] Primera consulta busca en PostgreSQL y guarda en Redis; siguientes consultas leen de Redis en < 5 ms.
  - [ ] Si el contenedor de Redis se detiene, la API degrada elegantemente consultando la base de datos sin caerse (RNF-DIS-04).

---

## Épica 4: Interoperabilidad, Mocks, Pagos y Facturación
**Responsable Principal:** Jose Daniel Gonzalez Cochea  
**Ponderación Rúbrica:** 6.0 Puntos

### Ticket `SRP-008`: Validador Algorítmico Nativo de Cédula Ecuatoriana (Módulo 10)
* **Tipo:** Feature / Dominio
* **Asignado:** Jose Daniel Gonzalez Cochea
* **Rama:** `feature/mocks-payments-invoicing`
* **Descripción:** Implementar en Go el validador estricto de cédula de identidad conforme a la normativa ecuatoriana (ADR-013):
  * Longitud de 10 dígitos numéricos.
  * Código de provincia entre `01` y `24` o `30`.
  * Tercer dígito menor a 6.
  * Ponderación alterna `[2,1,2,1,2,1,2,1,2]` y cálculo del dígito verificador.
* **Criterios de Aceptación:**
  - [ ] Cédulas con dígito verificador erróneo son rechazadas en milisegundos con HTTP 422 (*fail-fast*).
  - [ ] No se realizan llamadas de red al Registro Civil si el algoritmo matemático falla.
  - [ ] Pruebas unitarias cubriendo cédulas válidas e inválidas.

---

### Ticket `SRP-009`: Adaptadores Mocks de Registro Civil e Historial Laboral IESS
* **Tipo:** Integration / Mocks
* **Asignado:** Jose Daniel Gonzalez Cochea
* **Rama:** `feature/mocks-payments-invoicing`
* **Descripción:** Diseñar puertos desacoplados `IdentityProvider` y `EmploymentHistoryProvider` con implementaciones Mock deterministas para la demostración académica (ADR-010).
* **Criterios de Aceptación:**
  - [x] Simulación de validación dactilar y fecha de expedición con respuestas exitosas y de fallo.
  - [x] Validación de condición de Afiliado, Jubilado o Empleador según la cédula ingresada.
  - [x] Rechazo automático con mensaje reglamentario si no cumple la calidad solicitada.

---

### Ticket `SRP-010`: Pasarela de Pago con Tarjeta (Luhn) y Facturación SRI Mockeable
* **Estado:** **Retirado (03/11/2026).** La postulación al Directorio del BIESS es gratuita: no se cobra arancel ni tasa y no se emiten facturas. Se eliminó el código (`domain/payment.go`, `application/payment_service.go`, `infrastructure/payments/`, `database/payment_repository.go`, `handlers/payment_handler.go` y la ruta `POST /postulaciones/{id}/pagos`), las tablas `transaccion_pago` y `factura`, sus tipos, la columna `postulacion.estado_arancel`, el valor `COMPROBANTE_PAGO` de `tipo_documento` y el campo `Cargo.Arancel`. En una base existente se aplica `db/migrations/002_eliminar_pagos.sql`. RF-PAG-01, RF-PAG-02, RN-42, RN-43 y el ADR-014 quedan marcados como retirados en el SRS y el SDD. Se conserva la descripción original solo como registro.
* **Tipo:** Feature / Pagos
* **Asignado:** Jose Daniel Gonzalez Cochea
* **Rama:** `feature/mocks-payments-invoicing`
* **Nota de alcance:** el proceso de postulación es gratuito (ver SRP-016), así que ninguna pantalla consume este módulo. Se debe revisar con el equipo si se retira junto con RF-PAG-01, RF-PAG-02 y RN-42 del SRS, el ADR-014 y las secciones 7.6.2, 7.6.3 y 10.15 del SDD, y los campos `cargo.arancel` y `postulacion.estado_arancel`.
* **Descripción:** Implementar arquitectura hexagonal para pagos y facturación (ADR-014):
  * Validador de tarjetas de crédito con Algoritmo de Luhn.
  * Puerto `PaymentGateway` con adaptador `MockPaymentGateway` (soporta tarjeta aprobada y tarjeta con fondos insuficientes).
  * Puerto `InvoicingService` que emite factura electrónica con clave de acceso de 49 dígitos del SRI y encola el correo con el comprobante PDF.
* **Criterios de Aceptación:**
  - [ ] Transacción de pago y generación de factura son atómicas.
  - [ ] Si la tarjeta es declinada, no se emite factura y se retorna error 402.
  - [ ] Correo con la factura aparece visible en Mailpit (`http://localhost:8025`).

---

## Épica 5: Frontend SPA Modular & Formularios Reactivos (React + Tailwind)
**Responsable Principal:** Carlos Fernando Patiño Garcia  
**Ponderación Rúbrica:** 6.0 Puntos

### Ticket `SRP-011`: Setup de la SPA Modular en Vite con React, TypeScript y Tailwind
* **Tipo:** Frontend Setup
* **Asignado:** Carlos Fernando Patiño Garcia
* **Rama:** `feature/frontend-spa-ui`
* **Descripción:** Inicializar el proyecto SPA con Vite, Tailwind CSS utilitario, estructura organizada por `features/` (`auth/`, `postulacion/`, `pagos/`, `shared/`) y configuración de Axios con `baseURL` dinámico.
* **Criterios de Aceptación:**
  - [x] Interfaz completamente responsiva (Mobile, Tablet y Desktop).
  - [x] Componentes base reutilizables: botones, inputs con validación, modales, alertas y spinners/loaders.

---

### Ticket `SRP-012`: Pantalla de Autenticación y Bloqueo Reactivo (EF-11)
* **Tipo:** Frontend Feature
* **Asignado:** Carlos Fernando Patiño Garcia
* **Rama:** `feature/frontend-spa-ui`
* **Descripción:** Desarrollar vistas de acceso:
  * Landing page con selector "¿Registrarse?" / "¿Iniciar sesión?".
  * Formulario de validación de identidad con cédula y selector de "Código Dactilar" o "Fecha de Expedición".
  * Componente de **Cuenta Regresiva de 10 Minutos** cuando el servidor responda HTTP 423 de bloqueo temporal.
* **Criterios de Aceptación:**
  - [x] Feedback visual instantáneo ante credenciales incorrectas.
  - [x] Si se bloquea la cuenta, los inputs se deshabilitan y un temporizador en vivo muestra el tiempo restante en segundos.

---

### Ticket `SRP-013`: Formulario Dinámico Multi-sección de Postulación (EF-12)
* **Tipo:** Frontend Feature
* **Asignado:** Carlos Fernando Patiño Garcia
* **Rama:** `feature/cascading-selects-forms`
* **Descripción:** Crear el wizard/formulario de postulación con React Hook Form + Zod:
  * Sección 1: Datos Personales (autocompletados en solo lectura).
  * Sección 2: Información Domiciliaria.
  * Sección 3: Información adicional (Familiares en IESS/BIESS).
  * Sección 4: Selección de Cargo y Elegibilidad.
  * Sección 5: Formación Académica, Experiencia y Capacitación.
  * Sección 6: Inhabilidades (13 preguntas con alertas).
  * Sección 7: Consentimiento informado y firma digital.
* **Criterios de Aceptación:**
  - [x] Botón de "Guardar Progreso" en cada sección para evitar pérdida de datos.
  - [x] Cálculo automático del tiempo de experiencia laboral acumulada en años, meses y días.
  - [x] Validación de archivos PDF (máximo 2 MB / 4 MB según corresponda) con vista previa del nombre cargado.
* **Notas de implementación:**
  * El formulario sigue las 10 secciones de SRS 3.5.1 / SDD 8.3 (Personal, Domicilio, Familiares, Cargo, Formación, Experiencia, Publicaciones, Capacitación, Inhabilidades, Resumen). La "firma digital" no se implementa: el SDD 8.5.9 indica que el PDF se imprime y firma a mano.
  * Código en `frontend/src/features/postulacion/`. Pruebas: `cd frontend && npm test`.
* **Integración backend mínima (autorizada por el responsable del frontend):** ningún ticket asignaba la exposición HTTP de la postulación (SDD 7.3.4) y el formulario la necesita para guardar contra PostgreSQL. Se implementaron, protegidos con JWT:

  | Método | Ruta | Función |
  |---|---|---|
  | GET | `/api/v1/postulaciones/actual` | Postulación vigente del postulante o 404 `POS_NO_ENCONTRADA`. |
  | POST | `/api/v1/postulaciones` | Crea el borrador único por convocatoria (409 `POS_YA_EXISTE`, 409 `POS_FASE_CERRADA`). |
  | GET | `/api/v1/postulaciones/{id}` | Estado, datos verificados del postulante y valores guardados por sección. |
  | PATCH / PUT | `/api/v1/postulaciones/{id}/{sección}` | Guarda una sección con control de versión (409 `POS_CONFLICTO_VERSION`). Rutas del SDD: `informacion-personal`, `domicilio`, `familiares`, `respaldo-organizacional`, `formacion-academica`, `experiencia`, `publicaciones`, `capacitaciones`, `inhabilidades`. |

  * Capas (ADR-003): `domain/formulario.go` → `application/formulario_service.go` → `infrastructure/database/formulario_repository.go` → `interfaces/http/handlers/postulacion_handler.go`; inyección en `cmd/api/main.go`.
  * **Almacenamiento transitorio:** cada sección se guarda como JSONB en la tabla nueva `postulacion_seccion` (`db/migrations/001_postulacion_seccion.sql`, también añadida a `db/init.sql`). En una base existente se aplica con `docker exec -i srp-biess-postgres psql -U srp_user -d srp_biess < db/migrations/001_postulacion_seccion.sql`. La prueba `TestFormularioRepository_Integracion` aplica `db/migrations/*.sql` (idempotentes) antes de ejecutarse, por lo que no depende de este paso manual.
  * **Usuario de prueba (solo desarrollo):** `db/seeds/postulante_prueba.sql` crea el postulante `0926687856` / `BiessPostulante2026*` (idempotente). Se aplica con `docker exec -i srp-biess-postgres psql -U srp_user -d srp_biess < db/seeds/postulante_prueba.sql`. No se incluye en `db/init.sql` para que no llegue a otros entornos.
  * El guardado es atómico en una sola sentencia SQL (versión + sección), sin depender del `TxManager`.
  * **Pendiente en SRP-017:** `PUT /cargo` con elegibilidad, `/documentos`, `/enviar` (con el `TxManager` de SRP-005), validación de cada sección en servidor y normalización de las secciones en las tablas de SDD 4.6. Mientras tanto, el recorrido completo (secciones 5 a 10, adjuntos y envío) se demuestra con `VITE_POSTULACION_MOCK=true npm run dev`. Los catálogos geográficos se documentan en SRP-014.

---

## Épica 6: Consumo de APIs, Rutas Protegidas & Cascada Geográfica
**Responsable Principal:** Carlos Fernando Patiño Garcia  
**Ponderación Rúbrica:** 6.0 Puntos

### Ticket `SRP-014`: Cascada Geográfica Reactiva (País → Provincia → Cantón)
* **Tipo:** Frontend Feature
* **Asignado:** Carlos Fernando Patiño Garcia
* **Rama:** `feature/cascading-selects-forms`
* **Descripción:** Implementar los combos anidados dependientes:
  * Al seleccionar "Ecuador", se consultan las 24 provincias mediante API/Caché.
  * Al seleccionar una provincia, se cargan únicamente los cantones de dicha provincia.
  * Si el país es distinto de Ecuador, los selectores de provincia y cantón se convierten dinámicamente en inputs de texto para captura manual.
* **Criterios de Aceptación:**
  - [x] Limpieza automática de campos hijos cuando cambia la selección del padre.
  - [x] Manejo de estados de carga (skeletons/spinners) mientras se obtienen los datos.
* **Notas de implementación:** componente `frontend/src/features/postulacion/components/UbicacionCascada.tsx`, usado en la sección Domicilio de SRP-013. Al cambiar el país se reinician provincia y cantón; al cambiar la provincia se reinicia el cantón; ambos cambios se anuncian a lectores de pantalla. Si la provincia no tiene cantones en el catálogo, se muestra un aviso y el cantón deja de ser obligatorio (RF-EF12-07), para no bloquear el guardado.
* **Integración backend mínima (autorizada por el responsable del frontend):** se exponen los catálogos públicos de SDD 7.3.2 con el `CatalogoService` y el repositorio existentes de SRP-004:

  | Método | Ruta | Respuesta |
  |---|---|---|
  | GET | `/api/v1/catalogos/paises` | Países activos `[{ id, nombre, codigo }]`. |
  | GET | `/api/v1/catalogos/paises/{paisId}/provincias` | Provincias del país (24 para Ecuador). |
  | GET | `/api/v1/catalogos/provincias/{provinciaId}/cantones` | Solo los cantones de esa provincia; `[]` si no tiene. |

  * Handler `internal/interfaces/http/handlers/catalogo_handler.go`, rutas en `router.go` e inyección en `cmd/api/main.go`. Sin PostgreSQL responden 503; con un identificador no numérico, 400.
  * Prueba de integración contra el seed: `go test ./internal/interfaces/http/handlers -run Catalogo -v`.
  * **Caché (SRP-007, PR #32):** `cmd/api/main.go` envuelve el repositorio con `cache.NewCachedCatalogoRepository` (cache-aside, TTL 24 h ± 5 min, claves `srp:{env}:cat:{recurso}:v{n}`). Si Redis no responde, los catálogos se leen de PostgreSQL. Tras actualizar la semilla (SRP-019) se debe invalidar la caché incrementando `srp:dev:cat:version`.
  * `GET /catalogos/paises` incluye `codigo` (ISO 3166-1 alfa-2): la SPA reconoce a Ecuador por `EC` y no por su id (en la semilla actual, `id = 52`).
  * **Semilla (`db/init.sql`):** 194 países, las 24 provincias del Ecuador y sus 222 cantones con el código oficial del INEC (ver SRP-019). Con el catálogo completo, todas las provincias tienen cantones y el aviso de "provincia sin cantones" ya no aparece.

---

### Ticket `SRP-015`: Rutas Protegidas, Interceptor de Sesión y Renovación Automática
* **Tipo:** Frontend Architecture
* **Asignado:** Carlos Fernando Patiño Garcia
* **Rama:** `feature/frontend-spa-ui`
* **Descripción:** Configurar `ProtectedRoute` en React Router y un interceptor de Axios que gestione el ciclo de vida del JWT:
  * Redirección al login si el usuario no tiene sesión activa.
  * Renovación transparente del Access Token ante respuestas 401 llamando a `/auth/refresh`.
* **Criterios de Aceptación:**
  - [x] No es posible acceder a `/formulario` sin haberse autenticado.
  - [x] Cierre de sesión limpio que elimina el estado global y limpia las cookies.
* **Notas de implementación:**
  * Rutas (React Router 6, `App.tsx`): `/` portada, `/ingresar`, `/registro` y `/formulario`; cualquier otra ruta vuelve a `/`. `ProtectedRoute` y `PublicOnlyRoute` en `features/auth/session/routes.tsx`: sin sesión, `/formulario` lleva a `/ingresar` y, tras ingresar, regresa a la página pedida; con sesión, las páginas públicas llevan a `/formulario` (RF-EF11-10). Mientras se restaura la sesión se muestra un skeleton.
  * Estado global de sesión en `features/auth/session/SessionProvider.tsx`: restaura la sesión al cargar con `/auth/refresh` y `/auth/me`; el cierre de sesión llama a `/auth/logout` (el backend borra la cookie `refresh_token`), limpia el token en memoria y los borradores locales.
  * Interceptor (`services/api/client.ts`): agrega el `Bearer`, y ante un 401 renueva el token una sola vez (una única llamada aunque fallen varias peticiones a la vez) y reintenta. Si la renovación falla con una sesión abierta, `expireSession()` lleva a `/ingresar` con el aviso "Su sesión expiró…" y borra los borradores.
  * Al navegar, el foco pasa al `h1` de la página nueva.
  * Pruebas: `frontend/src/App.test.tsx` y `frontend/src/services/api/client.test.ts`.
  * **Despliegue:** el servidor que sirva la SPA debe devolver `index.html` para cualquier ruta (fallback de SPA); si no, recargar en `/formulario` da 404. En desarrollo, Vite ya lo hace.

---

### Ticket `SRP-016`: Pantalla de Checkout / Pago de Arancel y Descarga de Comprobante
* **Estado:** **No se implementa (fuera de alcance).** La postulación al Directorio del BIESS es gratuita y administrativa: la interfaz no debe aludir en ningún momento a un arancel, tasa o valor económico. El recorrido termina con la declaración jurada, el código único, el envío del resumen al correo y la entrega física del formulario impreso y firmado (SRP-013, SDD 8.5.8 y 8.5.9). Se conserva la descripción original solo como registro.
* **Tipo:** Frontend Feature
* **Asignado:** Carlos Fernando Patiño Garcia
* **Rama:** `feature/cascading-selects-forms`
* **Descripción:** Desarrollar el modal/pantalla de pago de tasa con tarjeta:
  * Input interactivo de tarjeta con detección automática del logo (Visa, Mastercard).
  * Validación en tiempo real con Luhn en el frontend antes de enviar.
  * Confirmación de pago exitoso con botón de descarga inmediata de la factura y del PDF del concurso.
* **Criterios de Aceptación:**
  - [ ] Bloqueo de doble clic en el botón de pago para evitar cobros duplicados.
  - [ ] Visualización clara del número de autorización y clave de acceso del SRI.

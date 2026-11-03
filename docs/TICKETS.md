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
* **Tipo:** Feature / Pagos
* **Asignado:** Jose Daniel Gonzalez Cochea
* **Rama:** `feature/mocks-payments-invoicing`
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
  * **Pendiente para SRP-005:** normalizar las secciones en las tablas de SDD 4.6; validar en servidor cada sección (RF-EF12-24; hoy la completitud la informa el cliente); `PUT /cargo` con elegibilidad, `/documentos`, `/guardar` y `/enviar`. Los catálogos geográficos se documentan en SRP-014.

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
  | GET | `/api/v1/catalogos/paises` | Países activos `[{ id, nombre }]`. |
  | GET | `/api/v1/catalogos/paises/{paisId}/provincias` | Provincias del país (24 para Ecuador). |
  | GET | `/api/v1/catalogos/provincias/{provinciaId}/cantones` | Solo los cantones de esa provincia; `[]` si no tiene. |

  * Handler `internal/interfaces/http/handlers/catalogo_handler.go`, rutas en `router.go` e inyección en `cmd/api/main.go`. Sin PostgreSQL responden 503; con un identificador no numérico, 400.
  * Prueba de integración contra el seed: `go test ./internal/interfaces/http/handlers -run Catalogo -v`.
  * **Pendiente para SRP-007:** caché cache-aside en Redis (TTL 24 h) con fallback a PostgreSQL.
  * `GET /catalogos/paises` incluye `codigo` (ISO 3166-1 alfa-2): la SPA reconoce a Ecuador por `EC` y no por su id (en la semilla actual, `id = 52`).
  * **Semilla (`db/init.sql`):** 194 países, las 24 provincias del Ecuador con códigos INEC y 199 cantones asignados a su provincia. **Pendiente (SRP-004):** el sufijo del código de cantón es secuencial y faltan cantones en algunas provincias (por ejemplo Azuay, Guayas, Loja, Morona Santiago y Santa Elena); el catálogo productivo debe cargarse desde la fuente oficial del INEC.

---

### Ticket `SRP-015`: Rutas Protegidas, Interceptor de Sesión y Renovación Automática
* **Tipo:** Frontend Architecture
* **Asignado:** Carlos Fernando Patiño Garcia
* **Rama:** `feature/frontend-spa-ui`
* **Descripción:** Configurar `ProtectedRoute` en React Router y un interceptor de Axios que gestione el ciclo de vida del JWT:
  * Redirección al login si el usuario no tiene sesión activa.
  * Renovación transparente del Access Token ante respuestas 401 llamando a `/auth/refresh`.
* **Criterios de Aceptación:**
  - [ ] No es posible acceder a `/formulario` sin haberse autenticado.
  - [ ] Cierre de sesión limpio que elimina el estado global y limpia las cookies.

---

### Ticket `SRP-016`: Pantalla de Checkout / Pago de Arancel y Descarga de Comprobante
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

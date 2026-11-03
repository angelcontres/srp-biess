# 14. Decisiones de arquitectura — SDD

Cada ADR registra el contexto, la decisión, las consecuencias y las alternativas consideradas. El objetivo es que la sustentación pueda explicar **por qué** se eligió cada tecnología o patrón.

---

## ADR-001 — Microservicio modular único

**Contexto.** El alcance del primer parcial incluye registro, autenticación, formulario, catálogos, documentos y correo. Se exige arquitectura desacoplada, pero no se justifica operar varios servicios distribuidos.

**Decisión.** Implementar un único microservicio transaccional con arquitectura interna en capas y puertos/adaptadores. Los módulos se separan por paquetes y límites de dominio.

**Consecuencias.**

- Menor complejidad de despliegue y transacciones locales.
- Mayor disciplina para no mezclar capas.
- Posibilidad de extraer módulos a futuro.
- No hay escalado independiente por módulo.

**Alternativas.** Microservicios separados (descartado por complejidad); monolito sin capas (descartado por rúbrica).

---

## ADR-002 — Fiber v2 como framework HTTP

**Contexto.** Se requiere alto rendimiento, middleware simple y buena integración con Go.

**Decisión.** Usar Fiber v2.

**Consecuencias.**

- API ergonómica y rápida.
- Middleware de CORS, logger, recover y rate limit.
- No usa `net/http` directamente; el equipo debe conocer Fiber.
- Compatible con el stack solicitado (Fiber/Gin).

**Alternativas.** Gin (válido, pero se elige Fiber por simplicidad y rendimiento); `net/http` puro (más código repetitivo).

---

## ADR-003 — Inyección de dependencias manual

**Contexto.** La rúbrica exige inyección de dependencias clara y verificable.

**Decisión.** Construir todas las dependencias en `cmd/api/main.go` con constructores explícitos. Sin contenedor de DI.

**Consecuencias.**

- Flujo de dependencias visible.
- Fácil de explicar y testear.
- Más líneas de wiring.
- Si crece, se puede migrar a `google/wire`.

**Alternativas.** `wire` (codegen, más complejidad); `fx`/`dig` (reflexión, menos didáctico).

---

## ADR-004 — pgx + TxManager

**Contexto.** Se necesita persistencia relacional, transacciones atómicas y buen rendimiento.

**Decisión.** Usar `pgx/v5` con `pgxpool` y un `TxManager` que ejecuta funciones dentro de `BEGIN/COMMIT/ROLLBACK`.

**Consecuencias.**

- Control explícito de transacciones.
- Repositorios reciben `Querier` (pool o tx).
- Mejor rendimiento que `database/sql` puro.
- Requiere disciplina para no abrir transacciones anidadas.

**Alternativas.** `database/sql` (más simple, menos eficiente); ORM (GORM) (más magia, menos control).

---

## ADR-005 — Redis cache-aside con TTL y versionado

**Contexto.** Los catálogos se consultan frecuentemente y la rúbrica exige caché con TTL.

**Decisión.** Cache-aside en Redis para países, provincias y cantones; TTL 24 h + jitter; invalidación por versión global.

**Consecuencias.**

- Baja latencia y menos carga en PostgreSQL.
- Consistencia eventual tras cambios de catálogo.
- Dependencia de Redis; fallback a PostgreSQL.
- No se cachean datos personales.

**Alternativas.** Cache in-memory (no distribuida); materialized views (más complejas); sin caché (no cumple rúbrica).

---

## ADR-006 — JWT corto + refresh opaco en cookie HttpOnly

**Contexto.** La SPA necesita sesión segura y escalable sin almacenar tokens en `localStorage`.

**Decisión.** Access JWT de 15 min en memoria; refresh opaco de 8 h en cookie `HttpOnly`, `Secure`, `SameSite=Strict`, con rotación y revocación en Redis.

**Consecuencias.**

- Menor exposición a XSS.
- Revocación y rotación controladas.
- Complejidad de refresh y CORS con credenciales.
- Requiere validar `Origin` en refresh/logout.

**Alternativas.** Tokens en `localStorage` (descartado por XSS); sesión en servidor (menos escalable); cookies para todo (más CSRF).

---

## ADR-007 — TanStack Query + Zustand

**Contexto.** La SPA maneja estado del servidor (postulación, catálogos) y estado de UI (sesión, modales).

**Decisión.** TanStack Query para estado del servidor; Zustand para sesión y UI.

**Consecuencias.**

- Caché, loading, error y reintentos integrados.
- Estado global simple.
- Dos librerías, pero con responsabilidades claras.
- Menos código boilerplate que Redux.

**Alternativas.** Redux Toolkit (más pesado); Context API (menos eficiente para datos remotos); SWR (similar).

---

## ADR-008 — React Hook Form + Zod

**Contexto.** El formulario tiene muchas secciones, campos condicionales y validaciones.

**Decisión.** React Hook Form para el estado del formulario y Zod para esquemas y tipos.

**Consecuencias.**

- Buen rendimiento y menos re-renderizados.
- Validación declarativa y tipos inferidos.
- Curva de aprendizaje baja.
- La validación del servidor sigue siendo obligatoria.

**Alternativas.** Formik (más re-renderizados); validación manual (más código); Yup (válido, pero Zod tipa mejor).

---

## ADR-009 — Almacenamiento abstracto local/S3

**Contexto.** En desarrollo se usa disco; en producción se espera object storage.

**Decisión.** Puerto `DocumentStorage` con implementaciones `LocalStorage` y `S3Storage`.

**Consecuencias.**

- Portabilidad sin cambiar la lógica de negocio.
- URLs firmadas en producción.
- Configuración por variables de entorno.
- Mayor complejidad inicial.

**Alternativas.** Solo disco (no escala); solo S3 (difícil en local); guardar en DB (no recomendado).

---

## ADR-010 — Adaptadores simulados para Registro Civil e IESS

**Contexto.** No se dispone de convenios ni credenciales productivas para la demo.

**Decisión.** Definir puertos `IdentityProvider` y `EmploymentHistoryProvider` con implementaciones HTTP y mocks deterministas.

**Consecuencias.**

- Demo reproducible sin servicios externos.
- Pruebas automatizadas estables.
- Los mocks deben documentarse claramente.
- La integración real requiere configuración y convenios.

**Alternativas.** Depender de servicios reales (riesgo de demo); hardcodear respuestas (menos testeable).

---

## ADR-011 — Outbox de notificaciones

**Contexto.** El envío de correo no debe bloquear ni revertir la postulación.

**Decisión.** Insertar la notificación en la misma transacción del envío y procesarla con un worker que reintenta con backoff.

**Consecuencias.**

- No se pierden correos.
- La postulación queda enviada aunque el SMTP falle.
- Se requiere worker y monitoreo.
- La entrega es eventual.

**Alternativas.** Enviar dentro de la transacción (bloquea y puede fallar); enviar después sin outbox (puede perderse).

---

## ADR-012 — PDF generado en el servidor

**Contexto.** El formulario debe imprimirse, firmarse y conservarse como evidencia.

**Decisión.** Generar el PDF en el servidor con una plantilla institucional y almacenarlo como documento.

**Consecuencias.**

- Formato consistente y verificable.
- Hash de integridad.
- Uso de CPU en el envío.
- La plantilla debe versionarse.

**Alternativas.** Generar en el navegador (menos control); imprimir HTML (menos portable); usar un servicio externo (dependencia).

---

## ADR-013 — Estrategia de Validación de Cédula de Identidad Ecuatoriana (Módulo 10 + Interoperabilidad)

**Contexto.**  
El acceso y registro al sistema exige verificar la identidad del ciudadano. Para evitar registros con identificaciones fraudulentas o errores tipográficos antes de invocar servicios externos, se requiere una validación exhaustiva tanto sintáctica/algorítmica como semántica de la cédula de ciudadanía ecuatoriana.

**Decisión.**  
Implementar una estrategia de validación en dos capas acoplada mediante una interfaz `NationalIDValidator`:
1. **Validación algorítmica local (offline/síncrona):** Algoritmo Módulo 10 ecuatoriano que verifica:
   - Longitud exacta de 10 dígitos numéricos.
   - Código de provincia válido (dos primeros dígitos entre `01` y `24`, o `30` para ecuatorianos en el exterior).
   - Tercer dígito menor a 6 (personas naturales).
   - Coeficientes multiplicadores `[2, 1, 2, 1, 2, 1, 2, 1, 2]`, resta de 9 si el producto es $\ge 10$, sumatoria de dígitos y cálculo de decena superior inmediata para verificar el dígito verificador.
2. **Validación de existencia y biometría (online/interoperable):** Verificación contra el Registro Civil de la fecha de expedición o código dactilar (simulado/mockeado mediante adaptador desacoplado para entorno académico, pero con firma de contrato idéntica a los servicios REST/SOAP gubernamentales).

**Consecuencias.**
- **Positivas:** Fallo rápido (*fail-fast*) que ahorra llamadas de red innecesarias y mitiga denegaciones de servicio si la cédula es matemáticamente inválida; total independencia entre la regla matemática de validación y la fuente externa de datos.
- **Negativas:** Obliga a mantener actualizado el catálogo de códigos de provincias si hubiese reformas territoriales; requiere sincronización de formatos de fecha entre cliente y servicio.

**Alternativas consideradas.**  
- *Validar solo en el Registro Civil:* Descartado porque satura la red e incrementa la latencia ante errores tipográficos evidentes.
- *Validar solo con expresión regular `^\d{10}$`:* Descartado por ser insuficiente según la normativa técnica ecuatoriana y la rúbrica de reglas de negocio.

---

## ADR-014 — Procesamiento de Pagos con Tarjeta y Facturación Electrónica mediante Puertos y Adaptadores (Mockeable e Implementable)

> **Retirado (03/11/2026).** La postulación al Directorio del BIESS es gratuita: no se cobra arancel ni tasa y no se emiten facturas. Esta sección se conserva solo como registro. Ver `db/migrations/002_eliminar_pagos.sql` y SRP-010 en `docs/TICKETS.md`.

**Contexto.**  
La gestión de trámites administrativos y aranceles/tasas de postulación exige validar instrumentos de pago electrónico (tarjeta de crédito/débito) y emitir la correspondiente factura electrónica con su respectivo comprobante enviado por correo. Dado que en el entorno de desarrollo y sustentación académica no se cuenta con credenciales activas del SRI ni pasarelas de pago transaccionales bancarias de producción (como Kushki, Payphone o Stripe), se requiere una solución que sea **completamente funcional y demostrable mediante Mocks**, pero **100% implementable en producción** con solo inyectar el adaptador real sin alterar el dominio.

**Decisión.**  
Adoptar el patrón de **Arquitectura Hexagonal (Puertos y Adaptadores)** mediante dos interfaces desacopladas en el núcleo del backend:
1. **Puerto `PaymentGateway`:**  
   - Define el contrato: `AuthorizeAndCapture(ctx, PaymentRequest) (*PaymentResult, error)`.
   - **Validación previa:** Validación de tarjeta en servidor aplicando el **Algoritmo de Luhn (Módulo 10)** para el PAN (Primary Account Number), detección del emisor (Visa, MasterCard, American Express, Diners), verificación de mes/año de expiración vigente y formato de CVV/CVC (3 o 4 dígitos).
   - **Adaptador `MockPaymentGateway`:** Permite números de tarjeta de prueba estandarizados (ej. tarjetas válidas de éxito, tarjetas que simulan fondos insuficientes o tarjeta declinada), simulando latencias reales de red (`200ms - 500ms`) y generando IDs de transacción idempotentes (`TXN-BIESS-...`).
   - **Adaptador `ProductionPaymentGateway`:** Preparado para interactuar con APIs REST de pasarelas reales vía Webhooks y Tokens seguros.
2. **Puerto `InvoicingService`:**  
   - Define el contrato: `GenerateAndSendInvoice(ctx, InvoiceRequest) (*InvoiceResult, error)`.
   - **Adaptador `MockInvoicingService`:** Genera un número de autorización secuencial de 49 dígitos conforme a la estructura de clave de acceso del SRI, crea el comprobante de factura en PDF y lo encola en la bandeja `outbox_notificaciones` para despacho por correo mediante la infraestructura de SMTP/Mailpit.
   - **Adaptador `SriElectronicBillingAdapter`:** Adaptador productivo que genera el XML firmado electrónicamente con formato XAdES-BES y lo envía a los Web Services del SRI.

**Consecuencias.**
- **Positivas:**  
  - La aplicación funciona al 100% en modo local/demo sin costo ni contratos bancarios/SRI.
  - La sustitución por un proveedor bancario o pasarela real se efectúa cambiando una variable de entorno (`PAYMENT_PROVIDER=mock|kushki|stripe` y `INVOICING_PROVIDER=mock|sri`) en el archivo `main.go` durante la inyección de dependencias.
  - El flujo transaccional garantiza que si el pago es declinado, la postulación o arancel no se confirma y la factura jamás se emite (integridad transaccional).
- **Negativas:**  
  - Requiere modelar entidades adicionales en base de datos (`transacciones_pago`, `facturas`) y sus tablas DDL correspondientes.
  - El equipo debe conocer el algoritmo de Luhn y el flujo asíncrono de facturación.

**Alternativas consideradas.**  
- *Llamar directamente a APIs externas en los controladores:* Descartado categóricamente porque viola el principio de inyección de dependencias (Rúbrica criterio 1) y hace imposible la demo sin conexión a internet o credenciales privadas.
- *Omitir la facturación y el pago:* Descartado para cumplir con el alcance integral de trámites institucionales y evaluación técnica.

---

## Resumen

| ADR | Decisión | Estado |
|---|---|---|
| 001 | Microservicio modular | Aceptada |
| 002 | Fiber v2 | Aceptada |
| 003 | DI manual | Aceptada |
| 004 | pgx + TxManager | Aceptada |
| 005 | Redis cache-aside | Aceptada |
| 006 | JWT + refresh cookie | Aceptada |
| 007 | TanStack Query + Zustand | Aceptada |
| 008 | RHF + Zod | Aceptada |
| 009 | Storage abstracto | Aceptada |
| 010 | Mocks de integraciones | Aceptada |
| 011 | Outbox | Aceptada |
| 012 | PDF en servidor | Aceptada |
| 013 | Validación de cédula Módulo 10 e Interoperabilidad | Aceptada |
| 014 | Pagos con tarjeta y facturación con Mocks implementables | Retirada (03/11/2026) |

# 10. Diseño de procesos y transacciones — SDD

## 10.1 Propósito

Este capítulo describe los procesos críticos del módulo, sus secuencias, límites transaccionales, estados, idempotencia y compensaciones. Es la guía para implementar los casos de uso y para demostrar la atomicidad exigida por la rúbrica.

---

## 10.2 Registro de postulante

```mermaid
sequenceDiagram
  participant P as Postulante
  participant SPA as SPA
  participant API as AuthHandler
  participant S as AuthService
  participant R as Redis
  participant RC as Registro Civil
  participant DB as PostgreSQL

  P->>SPA: Cédula + fecha/código dactilar
  SPA->>API: POST /auth/registro/validar-identidad
  API->>S: ValidateIdentity
  S->>R: CheckLock(reg, cedula)
  R-->>S: Sin bloqueo
  S->>RC: Validate(cedula, tipo, valor)
  RC-->>S: Identidad válida
  S->>R: Reset(attempts)
  S->>R: Guardar token de validación (10 min)
  S-->>API: Datos + token
  API-->>SPA: 200 OK
  P->>SPA: Correo + contraseña
  SPA->>API: POST /auth/registro
  API->>S: Register
  S->>DB: BEGIN
  S->>DB: INSERT postulante
  S->>DB: INSERT auditoría
  S->>DB: COMMIT
  S->>R: Invalidar token de validación
  S-->>API: Cuenta creada
  API-->>SPA: 201 Created
```

### Reglas

- La cuenta se crea solo con identidad validada.
- La cédula y el correo son únicos.
- La transacción incluye el alta y la auditoría.
- El token de validación se invalida al registrarse.
- No se inicia sesión automáticamente.

---

## 10.3 Bloqueo por intentos fallidos

```mermaid
sequenceDiagram
  participant SPA as SPA
  participant API as AuthHandler
  participant S as AuthService
  participant R as Redis
  participant RC as Registro Civil

  SPA->>API: POST validar-identidad
  API->>S: ValidateIdentity
  S->>R: CheckLock
  R-->>S: Sin bloqueo
  S->>RC: Validate
  RC-->>S: No coincide
  S->>R: RegisterFailure (Lua)
  R-->>S: {attempts: 3, locked: 1, ttl: 600}
  S-->>API: LockedError(600)
  API-->>SPA: 423 + retryAfterSeconds
  Note over SPA: Cuenta regresiva 10:00
  SPA->>API: Nuevo intento antes de expirar
  API->>S: ValidateIdentity
  S->>R: CheckLock
  R-->>S: Bloqueado (ttl)
  S-->>API: LockedError(ttl)
  API-->>SPA: 423
```

### Claves

- `srp:{env}:auth:reg:attempts:{cedula}`: contador con TTL 600 s.
- `srp:{env}:auth:reg:lock:{cedula}`: marca de bloqueo con TTL 600 s.
- El script Lua incrementa y crea el lock de forma atómica.
- Un error técnico del proveedor no incrementa el contador.
- Al expirar, ambas claves desaparecen y el contador se reinicia.

---

## 10.4 Login y refresh

```mermaid
sequenceDiagram
  participant SPA as SPA
  participant API as AuthHandler
  participant S as AuthService
  participant DB as PostgreSQL
  participant R as Redis

  SPA->>API: POST /auth/login
  API->>S: Login(cedula, password)
  S->>R: CheckLock(login, cedula)
  S->>DB: FindByCedula
  DB-->>S: Postulante + hash
  S->>S: Verificar Argon2id
  alt Credenciales válidas
    S->>R: Reset(login attempts)
    S->>S: Emitir access JWT + refresh opaco
    S->>R: Guardar hash refresh (8 h)
    S->>DB: UPDATE ultimo_acceso
    S-->>API: Tokens
    API-->>SPA: 200 + Set-Cookie refresh
  else Credenciales inválidas
    S->>R: RegisterFailure(login)
    S-->>API: InvalidCredentials
    API-->>SPA: 401
  end
```

### Refresh

1. La SPA llama `POST /auth/refresh` con la cookie.
2. El servidor busca el hash en Redis.
3. Si es válido, rota el refresh y emite un nuevo access.
4. Si no, responde 401 y la SPA redirige al login.

### Logout

1. `POST /auth/logout`.
2. Se elimina el refresh de Redis.
3. Se agrega el `jti` del access a denylist con TTL restante.
4. Se limpia la cookie.

---

## 10.5 Recuperación de contraseña

```mermaid
sequenceDiagram
  participant P as Postulante
  participant API as AuthHandler
  participant S as AuthService
  participant RC as Registro Civil
  participant R as Redis
  participant M as SMTP

  P->>API: POST /auth/recuperar-password
  API->>S: RequestRecovery
  S->>RC: Validate identidad
  RC-->>S: Válida
  S->>S: Generar token aleatorio
  S->>R: Guardar hash (30 min)
  S->>M: Enviar correo con enlace
  S-->>API: 202 Accepted (genérico)
  P->>API: POST /auth/restablecer-password
  API->>S: ResetPassword(token, nueva)
  S->>R: Buscar hash y validar
  S->>S: Validar política
  S->>R: Invalidar token
  S->>R: Revocar refresh tokens
  S-->>API: 204 No Content
```

### Reglas

- La respuesta es genérica para evitar enumeración.
- El token es de un solo uso y expira en 30 min.
- Al cambiar la contraseña se revocan las sesiones previas.
- El correo incluye un enlace con el token; el token no se registra en logs.

---

## 10.6 Creación de borrador y guardado de sección

```mermaid
sequenceDiagram
  participant SPA as SPA
  participant API as PostulacionHandler
  participant S as PostulacionService
  participant DB as PostgreSQL

  SPA->>API: GET /postulaciones/actual
  API->>S: GetActual
  S->>DB: Buscar vigente
  alt No existe
    S->>DB: BEGIN
    S->>DB: INSERT postulacion BORRADOR
    S->>DB: COMMIT
    S-->>API: Borrador creado
  else Existe
    S-->>API: Borrador
  end
  API-->>SPA: Postulación + sección

  SPA->>API: PATCH /informacion-personal
  API->>S: SavePersonalInfo
  S->>DB: BEGIN
  S->>DB: SELECT ... FOR UPDATE
  S->>DB: UPSERT datos
  S->>DB: UPDATE version
  S->>DB: INSERT auditoría
  S->>DB: COMMIT
  S-->>API: OK + version
  API-->>SPA: 200
```

### Reglas

- El borrador se crea una sola vez por postulante y convocatoria.
- Cada sección se guarda en su propia transacción.
- Se usa `version` para detectar conflictos.
- Si la sección tiene ítems repetibles, se guardan en la misma transacción.
- La auditoría registra antes/después.

---

## 10.7 Validación de elegibilidad

```mermaid
sequenceDiagram
  participant SPA as SPA
  participant API as PostulacionHandler
  participant S as PostulacionService
  participant E as EligibilityService
  participant IESS as Historial Laboral
  participant DB as PostgreSQL

  SPA->>API: PUT /cargo (confirmacion=true)
  API->>S: SaveCargo
  S->>E: Validate(cedula, sector)
  E->>IESS: GetQuality(cedula, sector)
  alt Elegible
    IESS-->>E: esAfiliado/jubilado/empleador = true
    E-->>S: Elegible
    S->>DB: BEGIN
    S->>DB: UPDATE cargo + elegibilidad
    S->>DB: INSERT auditoría
    S->>DB: COMMIT
    S-->>API: Elegible + requiereRespaldo
    API-->>SPA: 200
  else No elegible
    IESS-->>E: false
    E-->>S: No elegible
    S->>DB: BEGIN
    S->>DB: UPDATE estado NO_ELEGIBLE
    S->>DB: INSERT auditoría
    S->>DB: COMMIT
    S-->>API: BusinessError
    API-->>SPA: 422 + mensaje institucional
  else Servicio caído
    IESS-->>E: timeout
    E-->>S: Indeterminado
    S-->>API: ExternalError
    API-->>SPA: 503
  end
```

### Reglas

- La confirmación del cargo es obligatoria antes de consultar.
- Un fallo técnico no marca `NO_ELEGIBLE`.
- Un resultado negativo cierra la postulación y audita el motivo.
- El mensaje es el definido en el SRS (`07-reglas-de-negocio.md`).

---

## 10.8 Carga de documento

```mermaid
sequenceDiagram
  participant SPA as SPA
  participant API as DocumentoHandler
  participant S as DocumentoService
  participant ST as Storage
  participant DB as PostgreSQL

  SPA->>API: POST /documentos (multipart)
  API->>S: Upload
  S->>S: Validar MIME, magic bytes, tamaño, total
  alt Válido
    S->>ST: Save(uuid.pdf)
    ST-->>S: OK
    S->>DB: BEGIN
    S->>DB: INSERT documento
    S->>DB: INSERT auditoría
    S->>DB: COMMIT
    S-->>API: Documento
    API-->>SPA: 201 Created
  else Inválido
    S-->>API: ValidationError
    API-->>SPA: 413/415
  end
```

### Compensación

Si el archivo se guarda pero falla la transacción de base de datos:

1. Se ejecuta `ROLLBACK`.
2. Se llama `storage.Delete(key)`.
3. Se registra el incidente.
4. Se devuelve error al usuario.

---

## 10.9 Envío de postulación (transacción crítica)

```mermaid
sequenceDiagram
  participant SPA as SPA
  participant API as PostulacionHandler
  participant S as SubmitService
  participant DB as PostgreSQL
  participant PDF as PdfGenerator
  participant ST as Storage
  participant OUT as Outbox
  participant W as Worker
  participant SMTP as SMTP

  SPA->>API: POST /enviar (Idempotency-Key)
  API->>S: Submit
  S->>S: Verificar idempotencia (Redis)
  S->>DB: BEGIN
  S->>DB: SELECT postulacion FOR UPDATE
  S->>S: Validar estado, secciones, consentimiento, documentos
  S->>DB: nextval(seq_codigo_postulacion)
  S->>S: Componer código único
  S->>PDF: Generar PDF
  PDF-->>S: bytes
  S->>ST: Guardar PDF
  S->>DB: UPDATE estado ENVIADA + codigo + fecha
  S->>DB: INSERT documento (FORMULARIO_POSTULACION)
  S->>DB: INSERT notificacion (PENDIENTE)
  S->>DB: INSERT auditoría
  S->>DB: COMMIT
  S->>S: Guardar respuesta idempotente (24 h)
  S-->>API: Código + estado + PDF
  API-->>SPA: 200
  W->>OUT: Tomar notificación
  W->>SMTP: Enviar PDF
  W->>DB: UPDATE notificacion ENVIADA
```

### Límites transaccionales

| Paso | ¿Dentro de la transacción? | Razón |
|---|---|---|
| Validar estado y datos | Sí | Consistencia. |
| Obtener secuencia | Sí | Evitar códigos duplicados. |
| Generar PDF | Sí (o generar antes y confirmar) | Si falla, rollback. |
| Guardar PDF | Sí (registro) / archivo antes | Si falla DB, borrar archivo. |
| Actualizar estado | Sí | Atomicidad. |
| Insertar documento | Sí | Atomicidad. |
| Insertar notificación | Sí | Outbox. |
| Enviar correo | No | No bloquear la transacción. |

### Idempotencia

- Si la postulación ya está `ENVIADA`, se devuelve el mismo código.
- Si existe `srp:{env}:idem:submit:{id}`, se devuelve la respuesta guardada.
- Si no se puede obtener el lock `srp:{env}:lock:submit:{id}`, se responde 409.

### Rollback

Si falla cualquier paso dentro de la transacción:

- No se actualiza el estado.
- No se genera código.
- No se inserta documento.
- No se encola correo.
- Si el PDF se guardó antes, se elimina.

---

## 10.10 Caché de catálogos

```mermaid
sequenceDiagram
  participant SPA as SPA
  participant API as CatalogoHandler
  participant S as CatalogoService
  participant R as Redis
  participant DB as PostgreSQL

  SPA->>API: GET /catalogos/paises
  API->>S: GetPaises
  S->>R: GET cat:paises:v1
  alt Hit
    R-->>S: JSON
    S-->>API: Lista
  else Miss
    S->>DB: SELECT paises
    DB-->>S: Filas
    S->>R: SETEX cat:paises:v1 86400
    S-->>API: Lista
  end
  API-->>SPA: 200
```

### Invalidación

- Se incrementa `cat:version` cuando cambia un catálogo.
- Las claves nuevas usan la versión nueva.
- Las claves antiguas expiran por TTL.
- No se usa `KEYS *` en producción.

### Fallback

Si Redis falla:

- Se consulta PostgreSQL.
- Se registra una advertencia.
- No se bloquea la lectura.

---

## 10.11 Máquinas de estado

### Postulante

```mermaid
stateDiagram-v2
  [*] --> REGISTRADO
  REGISTRADO --> ACTIVO: validación completada
  ACTIVO --> BLOQUEADO: 3 intentos fallidos
  BLOQUEADO --> ACTIVO: expira bloqueo
  ACTIVO --> INACTIVO: baja administrativa
  INACTIVO --> [*]
```

### Postulación

```mermaid
stateDiagram-v2
  [*] --> BORRADOR
  BORRADOR --> LISTA_PARA_ENVIO: guardar + consentimiento
  LISTA_PARA_ENVIO --> ENVIADA: enviar
  BORRADOR --> NO_ELEGIBLE: elegibilidad negativa
  LISTA_PARA_ENVIO --> NO_ELEGIBLE: elegibilidad negativa
  ENVIADA --> EN_REVISION: revisión SDNGTH (futuro)
  ENVIADA --> NO_ELEGIBLE: validación posterior (futuro)
  EN_REVISION --> CERRADA: cierre (futuro)
  NO_ELEGIBLE --> CERRADA: cierre
  ENVIADA --> CERRADA: cierre
```

### Reglas

- `ENVIADA` es inmutable.
- `NO_ELEGIBLE` no permite continuar.
- `CERRADA` es solo lectura.
- Las transiciones se validan en el servicio y en la base.

---

## 10.12 Límites transaccionales

| Caso de uso | Transacción | Aislamiento |
|---|---|---|
| Registro | Alta + auditoría. | Read committed. |
| Guardar sección | Upsert + auditoría. | Read committed + `version`. |
| Guardar ítem | Ítem + documento + auditoría. | Read committed. |
| Elegibilidad | Cargo + estado + auditoría. | Read committed. |
| Envío | Estado + código + documento + outbox + auditoría. | Read committed + `FOR UPDATE`. |
| Carga documento | Documento + auditoría. | Read committed. |
| Catálogos | Solo lectura. | — |

### Regla

> Las operaciones de escritura que afectan más de una tabla usan `TxManager.WithTx`. Las lecturas pueden usar el pool directamente.

---

## 10.13 Concurrencia e idempotencia

### Concurrencia optimista

```sql
UPDATE postulacion
SET estado = $1, version = version + 1, fecha_guardado = now()
WHERE id = $2 AND version = $3;
```

Si `RowsAffected() == 0`, se devuelve `POS_CONFLICTO_VERSION`.

### Idempotencia de envío

- `Idempotency-Key` o `postulacionId`.
- Redis guarda la respuesta por 24 h.
- Si la clave existe, se devuelve la misma respuesta.
- Si el estado ya es `ENVIADA`, se devuelve el código existente.

### Lock distribuido

```text
SET srp:{env}:lock:submit:{id} "1" NX PX 10000
```

Si no se obtiene, otro proceso está enviando; se responde 409.

---

## 10.14 Errores y compensaciones

| Falla | Acción |
|---|---|
| Registro Civil no responde | 503; no consumir intento. |
| IESS no responde | 503; no marcar `NO_ELEGIBLE`. |
| Redis no disponible | Catálogos desde DB; auth fail-closed. |
| PDF falla | Rollback; no cambiar estado. |
| Storage falla | Rollback; no insertar documento. |
| DB falla en upload | Borrar archivo; error. |
| SMTP falla | Postulación enviada; outbox reintenta. |
| Correo duplicado | Idempotencia evita duplicado. |
| Conflicto de versión | 409; recargar. |

### Outbox

- La notificación se inserta en la misma transacción del envío.
- El worker la procesa después del commit.
- Reintentos con backoff: 1 min, 5 min, 15 min, 1 h, 6 h.
- Después de 5 intentos, estado `FALLIDA` y alerta.
- La postulación no se revierte por un fallo de correo.

---

## 10.15 Flujo de validación de cédula y proceso de pago con tarjeta y facturación (ADR-013 y ADR-014)

### 10.15.1 Validación algorítmica de cédula (Módulo 10)

```mermaid
sequenceDiagram
  participant Cliente as Cliente (SPA/Usuario)
  participant Handler as AuthHandler
  participant Val as NationalIDValidator (Módulo 10)
  participant RC as IdentityProvider (Registro Civil)

  Cliente->>Handler: POST /validaciones/cedula o POST /auth/registro/validar-identidad
  Handler->>Val: ValidateCedula(cedula)
  alt Cédula no cumple longitud (10 dígitos) o provincia (01-24, 30) o Módulo 10
    Val-->>Handler: Error (VAL_CEDULA_INVALIDA)
    Handler-->>Cliente: 422 Unprocessable Entity (Fail-fast sin gastar llamadas de red)
  else Cédula matemáticamente válida
    Val-->>Handler: OK (Válida)
    alt Flujo de autenticación con Registro Civil
      Handler->>RC: Validate(cedula, tipoValidacion, valor)
      RC-->>Handler: IdentityResult
      Handler-->>Cliente: 200 OK con datos biográficos
    else Validación sintáctica aislada
      Handler-->>Cliente: 200 OK (valida: true)
    end
  end
```

### 10.15.2 Transacción de pago con tarjeta y emisión de factura

> **Retirado (03/11/2026).** La postulación al Directorio del BIESS es gratuita: no se cobra arancel ni tasa y no se emiten facturas. Esta sección se conserva solo como registro. Ver `db/migrations/002_eliminar_pagos.sql` y SRP-010 en `docs/TICKETS.md`.

```mermaid
sequenceDiagram
  participant SPA as SPA (Postulante)
  participant API as PagoHandler
  participant Svc as PagoService
  participant Luhn as Validador Luhn
  participant PG as PaymentGateway (Mock/Real)
  participant Tx as TxManager (PostgreSQL)
  participant Inv as InvoicingService (SRI/Mock)
  participant Out as OutboxNotificaciones

  SPA->>API: POST /postulaciones/{id}/pagos (Datos de tarjeta)
  API->>Svc: ProcesarPago(postulacionID, tarjeta, monto)
  Svc->>Luhn: ValidarPAN(numeroTarjeta, exp, cvv)
  alt Tarjeta inválida según Luhn o expirada
    Luhn-->>Svc: Error (TARJETA_INVALIDA)
    Svc-->>API: Error validación
    API-->>SPA: 422 Unprocessable Entity
  else Tarjeta sintácticamente válida
    Svc->>PG: AuthorizeAndCapture(tarjeta, monto)
    alt Pago Declinado / Fondos Insuficientes
      PG-->>Svc: PaymentResult (RECHAZADO, motivo)
      Svc->>Tx: WithTx -> INSERT INTO transaccion_pago (estado='RECHAZADO')
      Svc-->>API: Error (PAGO_DECLINADO)
      API-->>SPA: 402 Payment Required
    else Pago Aprobado
      PG-->>Svc: PaymentResult (APROBADO, authCode, txnID)
      Svc->>Tx: WithTx (BEGIN)
      Note over Svc,Tx: Inserción atómica multi-tabla
      Svc->>Tx: INSERT INTO transaccion_pago (APROBADO)
      Svc->>Inv: GenerateInvoice(datosCliente, monto, txnID)
      Inv-->>Svc: InvoiceResult (numeroFactura, claveAcceso, PDF/XML)
      Svc->>Tx: INSERT INTO factura (EMITIDA)
      Svc->>Tx: UPDATE postulacion SET estado_arancel='PAGADO'
      Svc->>Out: INSERT INTO notificacion (tipo='FACTURA_PAGO', adjuntos=[factura.pdf])
      Tx-->>Svc: COMMIT exitoso
      Svc-->>API: Pago y Factura confirmados
      API-->>SPA: 201 Created (transaccionId, factura, estado)
    end
  end
```


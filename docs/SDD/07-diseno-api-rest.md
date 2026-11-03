# 07. Diseño de API REST — SDD

## 7.1 Convenciones

| Aspecto | Convención |
|---|---|
| Base URL | `/api/v1` |
| Formato | JSON UTF-8 |
| Fechas | ISO 8601 (`2026-10-31T16:42:00-05:00`) |
| Autenticación | `Authorization: Bearer <access_token>` |
| Refresh | Cookie `HttpOnly`, `Secure`, `SameSite=Strict`, `Path=/api/v1/auth` |
| Errores | RFC 7807 + `code`, `traceId`, `errors` |
| Paginación | `page`, `pageSize`, `total`, `items` |
| Idempotencia | `Idempotency-Key` en `POST /enviar` |
| Versión | Prefijo `/v1`; cambios incompatibles → `/v2` |

### Cabeceras recomendadas

```http
Authorization: Bearer <jwt>
Content-Type: application/json
Accept: application/json
X-Request-Id: <uuid opcional>
Idempotency-Key: <uuid opcional>
```

### Respuesta de error

```json
{
  "type": "https://api.biess.gob.ec/errors/validation",
  "title": "Error de validación",
  "status": 422,
  "code": "VALIDATION_ERROR",
  "detail": "La solicitud contiene campos inválidos.",
  "instance": "/api/v1/postulaciones/550e.../domicilio",
  "traceId": "8f3c2a1e-9b7d-4c2a-8f1e-123456789abc",
  "errors": [
    { "field": "telefonoCelular", "message": "El teléfono no es válido." }
  ]
}
```

---

## 7.2 Autenticación y sesión

### 7.2.1 Flujo

1. `POST /auth/login` valida cédula y contraseña.
2. Devuelve `accessToken` (15 min) en el body.
3. Establece `refresh_token` en cookie `HttpOnly` (8 h).
4. La SPA guarda el access token en memoria.
5. Al expirar, `POST /auth/refresh` renueva el access token y rota el refresh.
6. `POST /auth/logout` revoca el refresh y agrega el `jti` a denylist.

### 7.2.2 Claims del JWT

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "cedula": "0102030405",
  "role": "POSTULANTE",
  "jti": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "iss": "srp-biess",
  "aud": "srp-biess-spa",
  "iat": 1761946920,
  "exp": 1761947820
}
```

### 7.2.3 Reglas

- El access token no se almacena en `localStorage`.
- El refresh token es opaco, aleatorio y se guarda hasheado en Redis.
- La cookie no es accesible por JavaScript.
- `POST /auth/refresh` valida `Origin` y `SameSite`.
- Un refresh usado se invalida (rotación).
- Logout revoca el refresh y agrega el `jti` a denylist.

---

## 7.3 Catálogo de endpoints

### 7.3.1 Salud

| Método | Ruta | Auth | Descripción |
|---|---|---|---|
| GET | `/health/live` | No | Proceso vivo. |
| GET | `/health/ready` | No | Dependencias listas. |

### 7.3.2 Catálogos

| Método | Ruta | Auth | Descripción |
|---|---|---|---|
| GET | `/catalogos/paises` | No | Países activos. |
| GET | `/catalogos/paises/{paisId}/provincias` | No | Provincias del país. |
| GET | `/catalogos/provincias/{provinciaId}/cantones` | No | Cantones de la provincia. |

### 7.3.3 Autenticación

| Método | Ruta | Auth | Descripción |
|---|---|---|---|
| POST | `/auth/registro/validar-identidad` | No | Valida cédula + dato. |
| POST | `/auth/registro` | No | Crea la cuenta. |
| POST | `/auth/login` | No | Inicia sesión. |
| POST | `/auth/refresh` | Cookie | Renueva access token. |
| POST | `/auth/logout` | Sí | Cierra sesión. |
| POST | `/auth/recuperar-password` | No | Solicita recuperación. |
| POST | `/auth/restablecer-password` | No | Cambia contraseña. |
| GET | `/me` | Sí | Perfil del postulante. |

### 7.3.4 Postulación

| Método | Ruta | Auth | Descripción |
|---|---|---|---|
| GET | `/postulaciones/actual` | Sí | Postulación vigente o 404. |
| POST | `/postulaciones` | Sí | Crea borrador. |
| GET | `/postulaciones/{id}` | Sí | Consulta completa. |
| PATCH | `/postulaciones/{id}/informacion-personal` | Sí | Guarda datos personales. |
| PUT | `/postulaciones/{id}/domicilio` | Sí | Guarda domicilio. |
| GET | `/postulaciones/{id}/familiares` | Sí | Lista familiares. |
| POST | `/postulaciones/{id}/familiares` | Sí | Agrega familiar. |
| PUT | `/postulaciones/{id}/familiares/{familiarId}` | Sí | Edita familiar. |
| DELETE | `/postulaciones/{id}/familiares/{familiarId}` | Sí | Elimina familiar. |
| PUT | `/postulaciones/{id}/cargo` | Sí | Guarda y valida cargo. |
| GET | `/postulaciones/{id}/formacion-academica` | Sí | Lista formación. |
| POST | `/postulaciones/{id}/formacion-academica` | Sí | Agrega formación. |
| PUT | `/postulaciones/{id}/formacion-academica/{itemId}` | Sí | Edita formación. |
| DELETE | `/postulaciones/{id}/formacion-academica/{itemId}` | Sí | Elimina formación. |
| GET | `/postulaciones/{id}/experiencia` | Sí | Lista experiencia. |
| POST | `/postulaciones/{id}/experiencia` | Sí | Agrega experiencia. |
| PUT | `/postulaciones/{id}/experiencia/{itemId}` | Sí | Edita experiencia. |
| DELETE | `/postulaciones/{id}/experiencia/{itemId}` | Sí | Elimina experiencia. |
| GET | `/postulaciones/{id}/publicaciones` | Sí | Lista publicaciones. |
| POST | `/postulaciones/{id}/publicaciones` | Sí | Agrega publicación. |
| PUT | `/postulaciones/{id}/publicaciones/{itemId}` | Sí | Edita publicación. |
| DELETE | `/postulaciones/{id}/publicaciones/{itemId}` | Sí | Elimina publicación. |
| GET | `/postulaciones/{id}/capacitaciones` | Sí | Lista capacitación. |
| POST | `/postulaciones/{id}/capacitaciones` | Sí | Agrega capacitación. |
| PUT | `/postulaciones/{id}/capacitaciones/{itemId}` | Sí | Edita capacitación. |
| DELETE | `/postulaciones/{id}/capacitaciones/{itemId}` | Sí | Elimina capacitación. |
| PUT | `/postulaciones/{id}/inhabilidades` | Sí | Guarda respuestas. |
| POST | `/postulaciones/{id}/documentos` | Sí | Carga PDF. |
| DELETE | `/postulaciones/{id}/documentos/{docId}` | Sí | Elimina PDF. |
| POST | `/postulaciones/{id}/guardar` | Sí | Valida y guarda. |
| POST | `/postulaciones/{id}/enviar` | Sí | Envía postulación. |
| POST | `/postulaciones/{id}/pagos` | Sí | **Retirado (03/11/2026).** Procesa pago con tarjeta (Luhn + Gateway). |
| GET | `/postulaciones/{id}/factura` | Sí | **Retirado (03/11/2026).** Consulta y descarga de factura electrónica. |
| POST | `/validaciones/cedula` | No | Validación algorítmica de cédula (Módulo 10). |
| GET | `/postulaciones/{id}/pdf` | Sí | Descarga PDF. |
| GET | `/postulaciones/{id}/resumen` | Sí | Resumen y estado. |

---

## 7.4 Contratos de autenticación

### 7.4.1 Validar identidad

`POST /auth/registro/validar-identidad`

```json
{
  "cedula": "0102030405",
  "tipoValidacion": "FECHA_EXPEDICION",
  "valor": "15/06/2018"
}
```

**200 OK**

```json
{
  "identidadValidada": true,
  "tokenValidacion": "eyJ...",
  "datos": {
    "nombres": "MARÍA FERNANDA",
    "apellidos": "PÉREZ LÓPEZ",
    "genero": "FEMENINO",
    "fechaNacimiento": "1990-05-20",
    "nacionalidad": "ECUADOR"
  }
}
```

**Errores:** `AUTH_IDENTIDAD_INVALIDA` (422), `AUTH_REGISTRO_BLOQUEADO` (423), `INTEGRACION_NO_DISPONIBLE` (503).

### 7.4.2 Registro

`POST /auth/registro`

```json
{
  "tokenValidacion": "eyJ...",
  "correo": "maria.perez@example.com",
  "password": "MiClaveSegura#2026"
}
```

**201 Created**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "cedula": "0102030405",
  "correo": "maria.perez@example.com",
  "estado": "ACTIVO"
}
```

**Errores:** `AUTH_CEDULA_DUPLICADA` (409), `AUTH_CORREO_DUPLICADO` (409), `VALIDATION_ERROR` (422).

### 7.4.3 Login

`POST /auth/login`

```json
{
  "cedula": "0102030405",
  "password": "MiClaveSegura#2026"
}
```

**200 OK**

```json
{
  "accessToken": "eyJ...",
  "tokenType": "Bearer",
  "expiresIn": 900,
  "usuario": {
    "id": "550e...",
    "cedula": "0102030405",
    "nombres": "MARÍA FERNANDA",
    "correo": "maria.perez@example.com"
  }
}
```

**Errores:** `AUTH_CREDENCIALES_INVALIDAS` (401), `AUTH_LOGIN_BLOQUEADO` (423), `AUTH_CUENTA_INACTIVA` (403).

### 7.4.4 Refresh y logout

`POST /auth/refresh` no requiere body; usa la cookie. Devuelve un nuevo `accessToken` y rota el refresh.

`POST /auth/logout` revoca el refresh y devuelve `204 No Content`.

---

## 7.5 Contratos de postulación

### 7.5.1 Crear borrador

`POST /postulaciones`

**201 Created**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "estado": "BORRADOR",
  "seccionReanudacion": "personal",
  "porcentajeAvance": 0
}
```

**Errores:** `POS_FASE_CERRADA` (409), `POS_YA_EXISTE` (409).

### 7.5.2 Guardar información personal

`PATCH /postulaciones/{id}/informacion-personal`

```json
{
  "estadoCivil": "CASADO",
  "pareja": {
    "nombresApellidos": "JUAN CARLOS GÓMEZ",
    "cedula": "0912345678"
  },
  "version": 3
}
```

**200 OK**

```json
{
  "id": "550e...",
  "estado": "BORRADOR",
  "version": 4,
  "porcentajeAvance": 10
}
```

**Errores:** `POS_NACIONALIDAD_NO_ECUATORIANA` (422), `POS_CONFLICTO_VERSION` (409), `VALIDATION_ERROR` (422).

### 7.5.3 Guardar domicilio

`PUT /postulaciones/{id}/domicilio`

```json
{
  "paisId": 1,
  "provinciaId": 17,
  "cantonId": 1,
  "telefonoCelular": "+593991234567",
  "correoAlterno": "maria.alt@example.com",
  "callePrincipal": "Av. Amazonas",
  "numero": "N34-120",
  "calleSecundaria": "Av. Naciones Unidas",
  "sector": "La Carolina",
  "referencia": "Frente al parque",
  "telefonoDomicilio": "022345678",
  "version": 4
}
```

**200 OK** con `version` incrementada.

### 7.5.4 Guardar cargo

`PUT /postulaciones/{id}/cargo`

```json
{
  "cargoId": 1,
  "confirmacion": true,
  "version": 8
}
```

**200 OK (elegible)**

```json
{
  "elegible": true,
  "mensaje": "Postulante afiliado al IESS",
  "requiereRespaldo": true,
  "version": 9
}
```

**422 (no elegible)**

```json
{
  "code": "POS_NO_ELEGIBLE_AFI",
  "detail": "Su postulación ... NO puede continuar ...",
  "estado": "NO_ELEGIBLE"
}
```

### 7.5.5 Cargar documento

`POST /postulaciones/{id}/documentos` con `multipart/form-data`:

| Campo | Tipo | Obligatorio |
|---|---|---|
| `archivo` | File PDF | Sí |
| `tipo` | Enum | Sí |
| `itemId` | UUID | No (según sección) |

**201 Created**

```json
{
  "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "nombreOriginal": "titulo.pdf",
  "tamanoBytes": 1048576,
  "hashSha256": "...",
  "urlDescarga": "/api/v1/postulaciones/.../documentos/.../download"
}
```

**Errores:** `DOC_TIPO_INVALIDO` (415), `DOC_TAMANO_EXCEDIDO` (413), `DOC_TOTAL_EXCEDIDO` (413).

### 7.5.6 Guardar postulación

`POST /postulaciones/{id}/guardar`

```json
{
  "aceptaConsentimiento": true,
  "versionTerminos": "2026-10-31"
}
```

**200 OK**

```json
{
  "estado": "LISTA_PARA_ENVIO",
  "porcentajeAvance": 100,
  "advertencias": ["Existen períodos de experiencia superpuestos."]
}
```

**422** si faltan secciones, documentos, inhabilidades o consentimiento.

### 7.5.7 Enviar postulación

`POST /postulaciones/{id}/enviar` con `Idempotency-Key`.

```json
{
  "aceptaConsentimiento": true,
  "versionTerminos": "2026-10-31"
}
```

**200 OK**

```json
{
  "codigoUnico": "BIESS-2026-AFI-000123",
  "estado": "ENVIADA",
  "fechaEnvio": "2026-10-31T16:42:00-05:00",
  "pdfUrl": "/api/v1/postulaciones/550e.../pdf"
}
```

**Errores:** `POS_INCOMPLETA` (422), `POS_CERRADA` (409), `POS_YA_ENVIADA` (200 idempotente).

### 7.5.8 Descargar PDF

`GET /postulaciones/{id}/pdf`

- Valida pertenencia.
- Devuelve `application/pdf` con `Content-Disposition`.
- Si el PDF está en storage, redirige a URL firmada o lo transmite.
- Nombre: `Postulacion-BIESS-{codigo}.pdf`.

---

## 7.6 Códigos de estado

| Código | Uso |
|---|---|
| 200 | Operación exitosa. |
| 201 | Recurso creado. |
| 204 | Sin contenido (logout, delete). |
| 400 | Request malformado. |
| 401 | No autenticado o token expirado (`AUTH_TOKEN_EXPIRADO`). |
| 403 | No autorizado / recurso ajeno. |
| 404 | No encontrado. |
| 409 | Conflicto de estado o versión. |
| 413 | Archivo demasiado grande. |
| 415 | Tipo de archivo no soportado. |
| 422 | Validación o regla de negocio. |
| 423 | Bloqueo temporal. |
| 429 | Rate limit. |
| 500 | Error interno. |
| 503 | Dependencia no disponible. |

---

## 7.6 Contratos de validación de cédula, pagos y facturación (ADR-013 y ADR-014)

### 7.6.1 Validación algorítmica de cédula

`POST /validaciones/cedula`

```json
{
  "cedula": "1710034065"
}
```

**200 OK**
```json
{
  "valida": true,
  "provincia": "PICHINCHA",
  "codigoProvincia": "17",
  "tipo": "PERSONA_NATURAL",
  "algoritmo": "MODULO_10"
}
```

**422 Unprocessable Entity**
```json
{
  "codigo": "VAL_CEDULA_INVALIDA",
  "mensaje": "El número de cédula ingresado no cumple con el algoritmo Módulo 10 ecuatoriano",
  "detalles": [
    { "campo": "cedula", "error": "Dígito verificador incorrecto" }
  ]
}
```

### 7.6.2 Procesamiento de pago con tarjeta

> **Retirado (03/11/2026).** La postulación al Directorio del BIESS es gratuita: no se cobra arancel ni tasa y no se emiten facturas. Esta sección se conserva solo como registro. Ver `db/migrations/002_eliminar_pagos.sql` y SRP-010 en `docs/TICKETS.md`.

`POST /postulaciones/{id}/pagos`

```json
{
  "numeroTarjeta": "4532015112830366",
  "nombreTitular": "JUAN PEREZ",
  "mesExpiracion": 12,
  "anioExpiracion": 2028,
  "cvv": "123",
  "monto": 25.00,
  "moneda": "USD",
  "descripcion": "Tasa administrativa de postulación Concurso BIESS"
}
```

**201 Created**
```json
{
  "transaccionId": "TXN-BIESS-2026-884920",
  "estado": "APROBADO",
  "monto": 25.00,
  "moneda": "USD",
  "franquicia": "VISA",
  "ultimosCuatro": "0366",
  "autorizacion": "AUTH-MOCK-99412",
  "fecha": "2026-10-31T22:30:00-05:00",
  "factura": {
    "numeroFactura": "001-001-000004521",
    "claveAcceso": "3110202601176000132000120010010000045211234567819",
    "estado": "EMITIDA",
    "enviadaCorreo": true
  }
}
```

### 7.6.3 Consulta de factura electrónica

> **Retirado (03/11/2026).** La postulación al Directorio del BIESS es gratuita: no se cobra arancel ni tasa y no se emiten facturas. Esta sección se conserva solo como registro. Ver `db/migrations/002_eliminar_pagos.sql` y SRP-010 en `docs/TICKETS.md`.

`GET /postulaciones/{id}/factura`

**200 OK**
```json
{
  "facturaId": "f9a4e321-7299-4a0b-9311-20816223a412",
  "numeroFactura": "001-001-000004521",
  "fechaEmision": "2026-10-31T22:30:00-05:00",
  "claveAcceso": "3110202601176000132000120010010000045211234567819",
  "emisor": {
    "ruc": "1760001320001",
    "razonSocial": "BANCO DEL INSTITUTO ECUATORIANO DE SEGURIDAD SOCIAL BIESS"
  },
  "cliente": {
    "identificacion": "1710034065",
    "nombre": "JUAN CARLOS PEREZ",
    "email": "juan.perez@example.com"
  },
  "subtotal": 22.32,
  "iva": 2.68,
  "total": 25.00,
  "pdfUrl": "/api/v1/documentos/f9a4e321-factura.pdf",
  "xmlUrl": "/api/v1/documentos/f9a4e321-factura.xml"
}
```

---

## 7.7 Idempotencia y rate limiting

### Idempotencia

- `POST /postulaciones/{id}/enviar` requiere `Idempotency-Key`.
- Se guarda la respuesta en Redis por 24 h.
- Si se repite, se devuelve la misma respuesta sin duplicar.
- Si no se envía la cabecera, el servidor usa `postulacionId` como clave.

### Rate limiting

| Scope | Límite | Ventana |
|---|---|---|
| Login por IP | 20 req | 1 min |
| Login por cédula | 3 fallos | 10 min |
| Registro por cédula | 3 fallos | 10 min |
| Validación de identidad por IP | 30 req | 1 min |
| Catálogos por IP | 120 req | 1 min |
| Envío por postulación | 3 req | 1 min |

Respuesta `429`:

```http
Retry-After: 60
```

---

## 7.8 CORS y cabeceras

```http
Access-Control-Allow-Origin: https://concurso.biess.gob.ec
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Authorization, Content-Type, Idempotency-Key, X-Request-Id
```

Cabeceras de seguridad:

- `Content-Security-Policy`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Strict-Transport-Security` (producción)

---

## 7.9 OpenAPI

Se recomienda generar un `openapi.yaml` a partir de los DTO y mantenerlo versionado. El contrato debe cubrir:

- Esquemas de request/response.
- Códigos de error.
- Seguridad `bearerAuth` y cookie de refresh.
- Ejemplos.
- Validaciones.

El archivo puede ubicarse en `docs/api/openapi.yaml` del repositorio y publicarse con Swagger UI en un entorno de desarrollo.

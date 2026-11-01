# 06. Interfaces externas — SRS

## 6.1 Interfaces de usuario

La SPA debe exponer las siguientes pantallas o vistas. El detalle visual y de componentes se encuentra en `SDD/08-diseno-interfaz-usuario.md`.

| ID | Pantalla | Propósito | Actor |
|---|---|---|---|
| UI-01 | Inicio / bienvenida | Presentar el concurso y las opciones de acceso. | Postulante |
| UI-02 | Iniciar sesión | Autenticar con cédula y contraseña. | Postulante |
| UI-03 | Registro — validación de identidad | Solicitar cédula y fecha de expedición o código dactilar. | Postulante |
| UI-04 | Registro — credenciales | Crear correo y contraseña. | Postulante |
| UI-05 | Registro bloqueado | Informar el bloqueo de 10 minutos. | Postulante |
| UI-06 | Recuperar contraseña | Solicitar validación de identidad. | Postulante |
| UI-07 | Restablecer contraseña | Definir nueva contraseña. | Postulante |
| UI-08 | Formulario — información personal | Mostrar datos verificados y estado civil. | Postulante |
| UI-09 | Formulario — domicilio y contacto | Capturar ubicación y contacto. | Postulante |
| UI-10 | Formulario — familiares | Declarar y gestionar familiares. | Postulante |
| UI-11 | Formulario — cargo | Seleccionar y confirmar cargo. | Postulante |
| UI-12 | Formulario — formación académica | Gestionar títulos. | Postulante |
| UI-13 | Formulario — experiencia | Gestionar experiencia y ver tiempo calculado. | Postulante |
| UI-14 | Formulario — publicaciones | Gestionar publicaciones. | Postulante |
| UI-15 | Formulario — capacitación | Gestionar cursos, seminarios y talleres. | Postulante |
| UI-16 | Formulario — inhabilidades | Responder el cuestionario legal. | Postulante |
| UI-17 | Formulario — resumen y consentimiento | Revisar, aceptar y guardar. | Postulante |
| UI-18 | Postulación enviada | Mostrar código, estado y descarga. | Postulante |
| UI-19 | Mi postulación | Consultar estado y PDF. | Postulante |
| UI-20 | Error / no autorizado | Informar errores y ofrecer recuperación. | Postulante |

### Requisitos de interfaz

- Diseño responsive (360 px a 1920 px).
- Navegación por teclado y foco visible.
- Mensajes de error junto al campo y resumen al inicio del formulario.
- Indicadores de carga, éxito y error en todas las operaciones asíncronas.
- Confirmación antes de acciones irreversibles (envío, eliminación, cierre).
- Advertencia al salir con cambios sin guardar.
- Descarga de PDF con nombre legible: `Postulacion-BIESS-{codigo}.pdf`.

---

## 6.2 Interfaces de software

### 6.2.1 Registro Civil del Ecuador

| Aspecto | Especificación |
|---|---|
| **Propósito** | Validar identidad del postulante. |
| **Dirección** | API → Registro Civil. |
| **Datos enviados** | Cédula, tipo de validación (`FECHA_EXPEDICION` o `CODIGO_DACTILAR`), valor. |
| **Datos recibidos** | Coincidencia, nombres, apellidos, género, fecha de nacimiento, nacionalidad, estado civil (si disponible). |
| **Protocolo** | REST/JSON sobre HTTPS. |
| **Frecuencia** | Por intento de registro o recuperación. |
| **Timeout** | 3 s. |
| **Reintentos** | 1 reintento con backoff; no consumir intento si es error técnico. |
| **Errores** | Identidad no coincide, datos inválidos, servicio no disponible. |
| **Simulación** | Adaptador `MockRegistroCivil` con datos deterministas para la demo. |

### 6.2.2 Historial Laboral del IESS

| Aspecto | Especificación |
|---|---|
| **Propósito** | Validar calidad de afiliado, jubilado o empleador. |
| **Dirección** | API → IESS. |
| **Datos enviados** | Cédula, sector consultado. |
| **Datos recibidos** | `esAfiliado`, `esJubilado`, `esEmpleador`, mensaje. |
| **Protocolo** | REST/JSON sobre HTTPS. |
| **Frecuencia** | Al confirmar el cargo. |
| **Timeout** | 3 s. |
| **Reintentos** | 1; si falla, no se marca `NO_ELEGIBLE`. |
| **Simulación** | Adaptador `MockHistorialLaboral` con casos de prueba por cédula. |

### 6.2.3 Servicio de correo (SMTP)

| Aspecto | Especificación |
|---|---|
| **Propósito** | Enviar recuperación de contraseña y PDF de postulación. |
| **Dirección** | API → SMTP. |
| **Protocolo** | SMTP sobre TLS. |
| **Remitente** | `no-reply@biess.gob.ec` (configurable). |
| **Plantillas** | Recuperación, postulación enviada, alerta de fallo. |
| **Adjuntos** | PDF de postulación. |
| **Timeout** | 10 s. |
| **Reintentos** | 5 con backoff exponencial. |
| **Simulación** | Mailpit en local. |

### 6.2.4 PostgreSQL

| Aspecto | Especificación |
|---|---|
| **Propósito** | Persistencia transaccional. |
| **Dirección** | API → PostgreSQL. |
| **Protocolo** | TCP/SSL. |
| **Versión** | 16+. |
| **Acceso** | Usuario de aplicación con permisos mínimos; migraciones con usuario separado. |
| **Transacciones** | `BEGIN`, `COMMIT`, `ROLLBACK` en operaciones multi-tabla. |
| **Respaldo** | Diario completo + WAL continuo (producción). |

### 6.2.5 Redis

| Aspecto | Especificación |
|---|---|
| **Propósito** | Caché de catálogos, bloqueo de intentos, rate limiting, tokens temporales e idempotencia. |
| **Dirección** | API → Redis. |
| **Protocolo** | RESP sobre TCP/TLS. |
| **Versión** | 7+. |
| **Persistencia** | AOF recomendado para claves de seguridad; caché puede ser efímera. |
| **Bases lógicas** | DB 0: caché; DB 1: seguridad y tokens. |
| **Timeout** | 500 ms. |

### 6.2.6 Almacenamiento de archivos

| Aspecto | Especificación |
|---|---|
| **Propósito** | Guardar PDF de respaldo y formularios generados. |
| **Dirección** | API → almacenamiento. |
| **Desarrollo** | Volumen local fuera del directorio público. |
| **Producción** | S3 compatible o volumen cifrado. |
| **Acceso** | URLs firmadas con expiración de 5 min. |
| **Nombres** | UUID + `.pdf`; nunca el nombre original. |

---

## 6.3 Interfaces de comunicación

| Interfaz | Protocolo | Formato | Seguridad |
|---|---|---|---|
| SPA ↔ API | HTTPS | JSON | JWT en `Authorization`; refresh en cookie `HttpOnly`. |
| API ↔ PostgreSQL | TCP | SQL | TLS; credenciales por variables de entorno. |
| API ↔ Redis | TCP | RESP | TLS; contraseña; red privada. |
| API ↔ Registro Civil | HTTPS | JSON | mTLS o token institucional. |
| API ↔ IESS | HTTPS | JSON | mTLS o token institucional. |
| API ↔ SMTP | SMTP/TLS | MIME | Credenciales SMTP. |
| API ↔ Storage | HTTPS | Binario/API | Credenciales temporales. |

### Convenciones de API

- Base: `/api/v1`.
- Content-Type: `application/json` salvo carga de archivos (`multipart/form-data`).
- Fechas: ISO 8601 en API (`2026-10-31T15:30:00-05:00`).
- Errores: RFC 7807 con `code`, `detail`, `traceId` y `errors`.
- Paginación: `page`, `pageSize`, `total`, `items`.
- Idempotencia: cabecera `Idempotency-Key` en envío.

---

## 6.4 Interfaces de hardware

El sistema no requiere hardware especializado. El postulante utiliza un dispositivo con navegador y conexión a Internet. El servidor requiere CPU, memoria, disco y red según el dimensionamiento del SDD.

---

## 6.5 Resumen de endpoints

| Método | Ruta | Propósito |
|---|---|---|
| GET | `/health/live` | Salud del proceso. |
| GET | `/health/ready` | Disponibilidad de dependencias. |
| GET | `/catalogos/paises` | Listar países. |
| GET | `/catalogos/paises/{id}/provincias` | Listar provincias. |
| GET | `/catalogos/provincias/{id}/cantones` | Listar cantones. |
| POST | `/auth/registro/validar-identidad` | Validar identidad. |
| POST | `/auth/registro` | Crear cuenta. |
| POST | `/auth/login` | Iniciar sesión. |
| POST | `/auth/refresh` | Renovar acceso. |
| POST | `/auth/logout` | Cerrar sesión. |
| POST | `/auth/recuperar-password` | Solicitar recuperación. |
| POST | `/auth/restablecer-password` | Cambiar contraseña. |
| GET | `/me` | Perfil del postulante. |
| GET | `/postulaciones/actual` | Postulación vigente. |
| POST | `/postulaciones` | Crear borrador. |
| GET | `/postulaciones/{id}` | Consultar postulación. |
| PATCH | `/postulaciones/{id}/informacion-personal` | Guardar datos personales. |
| PUT | `/postulaciones/{id}/domicilio` | Guardar domicilio. |
| GET/POST/PUT/DELETE | `/postulaciones/{id}/familiares` | CRUD familiares. |
| PUT | `/postulaciones/{id}/cargo` | Guardar cargo. |
| GET/POST/PUT/DELETE | `/postulaciones/{id}/formacion-academica` | CRUD formación. |
| GET/POST/PUT/DELETE | `/postulaciones/{id}/experiencia` | CRUD experiencia. |
| GET/POST/PUT/DELETE | `/postulaciones/{id}/publicaciones` | CRUD publicaciones. |
| GET/POST/PUT/DELETE | `/postulaciones/{id}/capacitaciones` | CRUD capacitación. |
| PUT | `/postulaciones/{id}/inhabilidades` | Guardar inhabilidades. |
| POST | `/postulaciones/{id}/documentos` | Cargar documento. |
| DELETE | `/postulaciones/{id}/documentos/{docId}` | Eliminar documento. |
| POST | `/postulaciones/{id}/guardar` | Validar y guardar. |
| POST | `/postulaciones/{id}/enviar` | Enviar postulación. |
| GET | `/postulaciones/{id}/pdf` | Descargar PDF. |
| GET | `/postulaciones/{id}/resumen` | Resumen y estado. |

El contrato detallado se encuentra en `SDD/07-diseno-api-rest.md`.

---

## 6.6 Manejo de errores y tiempos

| Situación | Comportamiento |
|---|---|
| Error de validación | HTTP 422; errores por campo; no se pierde lo ingresado. |
| No autenticado | HTTP 401; redirigir al login. |
| No autorizado | HTTP 403; mensaje genérico. |
| Recurso no encontrado | HTTP 404. |
| Conflicto de estado | HTTP 409; mensaje y acción sugerida. |
| Bloqueo temporal | HTTP 423; `retryAfterSeconds`. |
| Rate limit | HTTP 429; `Retry-After`. |
| Error externo | HTTP 503; mensaje de reintento; no consumir intento. |
| Error interno | HTTP 500; `traceId`; log con detalle. |

---

## 6.7 Requisitos de seguridad de interfaces

- TLS 1.2+ en todas las comunicaciones externas.
- CORS restringido al dominio de la SPA.
- Cabeceras de seguridad: `Content-Security-Policy`, `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`.
- No exponer stack traces ni consultas SQL.
- URLs de descarga firmadas y de corta duración.
- Validar `Origin` en operaciones con cookie de refresh.

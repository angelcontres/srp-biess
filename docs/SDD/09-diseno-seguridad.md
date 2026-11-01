# 09. Diseño de seguridad — SDD

## 9.1 Modelo de amenazas (STRIDE)

| Amenaza | Escenario | Control |
|---|---|---|
| **Spoofing** | Suplantar a un postulante. | Validación de identidad, Argon2id, JWT firmado. |
| **Tampering** | Alterar una postulación enviada. | Inmutabilidad, control de estado, auditoría. |
| **Repudiation** | Negar una acción. | Auditoría con actor, IP, fecha y `trace_id`. |
| **Information disclosure** | Acceder a datos de otro postulante. | Autorización por pertenencia, URLs firmadas, logs redactados. |
| **Denial of service** | Fuerza bruta o floods. | Rate limiting, bloqueo, límites de body, timeouts. |
| **Elevation of privilege** | Acceder a funciones administrativas. | RBAC, `RequireRole`, rutas no expuestas. |

### Superficie de ataque

- Formularios públicos de registro y recuperación.
- Endpoints autenticados de postulación.
- Carga de archivos.
- Integraciones externas.
- Redis y PostgreSQL (red privada).
- Correo (posible phishing; se mitiga con enlaces firmados y plantillas).

---

## 9.2 Autenticación

### 9.2.1 Validación de identidad

- Cédula validada con algoritmo ecuatoriano antes de llamar al proveedor.
- Fecha de expedición o código dactilar validados por Registro Civil.
- Token de validación de un solo uso, TTL 10 min, almacenado en Redis.
- No se crea la cuenta hasta que la identidad esté validada.

### 9.2.2 Contraseñas

- Hash **Argon2id**: memoria 64 MB, iteraciones 3, paralelismo 2, salt 16 bytes, key 32 bytes.
- Comparación en tiempo constante.
- Política: mínimo 10, máximo 128, tres de cuatro clases.
- No se permiten contraseñas iguales a la cédula o al correo.
- No se registra la contraseña en logs ni auditoría.

### 9.2.3 Mensajes

- Login fallido: “Sus credenciales de ingreso son incorrectas.”
- Registro fallido: “La información ingresada es incorrecta.”
- Recuperación: mensaje genérico, sin revelar si la cuenta existe.

---

## 9.3 Autorización

### Roles

| Rol | Alcance |
|---|---|
| `POSTULANTE` | Su propia cuenta, su postulación, sus documentos. |
| `SISTEMA` | Procesos internos. |
| `ADMIN` | Futuro EF-13; no implementado. |

### Reglas

- Cada endpoint valida `sub` del JWT contra `postulante_id` del recurso.
- Un postulante no puede leer, editar ni eliminar recursos de otro.
- Los documentos se descargan con URL firmada y validación de pertenencia.
- Las rutas administrativas no existen en esta fase.
- `403` para recurso ajeno; no `404` si se quiere evitar enumeración.

### Matriz de permisos

| Recurso | POSTULANTE | SISTEMA | ADMIN |
|---|---|---|---|
| Perfil propio | CRUD | R | R |
| Postulación propia | CRUD | R/W | R |
| Documentos propios | CRUD | R | R |
| Catálogos | R | R | R |
| Auditoría | — | W | R (futuro) |

---

## 9.4 Sesiones y tokens

| Token | Formato | Duración | Almacenamiento |
|---|---|---|---|
| Access | JWT RS256 | 15 min | Memoria de la SPA. |
| Refresh | Opaco 256 bits | 8 h | Cookie `HttpOnly` + hash en Redis. |
| Validación de identidad | JWT corto | 10 min | Redis. |
| Recuperación | Opaco | 30 min | Redis (hash). |
| Idempotencia | UUID | 24 h | Redis. |

### Reglas

- El refresh rota en cada uso.
- Un refresh reutilizado se invalida y se alerta.
- Logout revoca el refresh y agrega el `jti` a denylist.
- El access token no se guarda en `localStorage`.
- La cookie usa `Secure`, `HttpOnly`, `SameSite=Strict`, `Path=/api/v1/auth`.
- Se valida `Origin` en refresh/logout.

---

## 9.5 Bloqueo y rate limiting

### Registro

- 3 fallos consecutivos → bloqueo 10 min.
- Clave: `srp:{env}:auth:reg:attempts:{cedula}` y `srp:{env}:auth:reg:lock:{cedula}`.
- Script Lua atómico.
- Respuesta `423` con `retryAfterSeconds`.

### Login

- 3 fallos consecutivos → bloqueo 10 min.
- Clave: `srp:{env}:auth:login:attempts:{cedula}` y `srp:{env}:auth:login:lock:{cedula}`.
- Un login exitoso reinicia el contador.

### Rate limiting

- Por IP y por cédula.
- Límites en `07-diseno-api-rest.md` §7.7.
- Respuesta `429` con `Retry-After`.

### Fail-closed

Si Redis no está disponible:

- Los catálogos se leen de PostgreSQL.
- El bloqueo de autenticación **no permite intentos ilimitados**: se rechaza temporalmente con `503` y se alerta.

---

## 9.6 Validación de entrada

| Capa | Validación |
|---|---|
| HTTP | Content-Type, tamaño, JSON válido, DTO. |
| Aplicación | Reglas de negocio, permisos, estado. |
| Dominio | Invariantes (fechas, nacionalidad, elegibilidad). |
| Base de datos | `UNIQUE`, `FOREIGN KEY`, `CHECK`. |

### Prevención

- **SQL injection:** consultas parametrizadas con pgx; nunca concatenar SQL.
- **XSS:** React escapa por defecto; no usar `dangerouslySetInnerHTML` con datos de usuario; CSP.
- **CSRF:** access token en header; refresh con `SameSite=Strict` y validación de `Origin`.
- **Mass assignment:** DTO explícitos; no bind directo a entidades.
- **Open redirect:** validar `redirect` contra una lista de rutas internas.
- **Enumeración:** mensajes genéricos y respuestas uniformes.

---

## 9.7 Carga de archivos

### Controles

1. Validar `Content-Type` y extensión `.pdf`.
2. Validar **magic bytes** `%PDF-`.
3. Validar tamaño por tipo y total.
4. Sanitizar nombre original.
5. Guardar con UUID fuera del directorio público.
6. Calcular SHA-256.
7. Registrar metadatos y auditoría.
8. Servir con URL firmada de 5 min.
9. Opcional: escanear con ClamAV y poner en cuarentena.

### Riesgos

| Riesgo | Mitigación |
|---|---|
| PDF malicioso | Antivirus, tipo real, no ejecutar. |
| Path traversal | Nombre UUID, no usar nombre original. |
| Sobrescritura | UUID único. |
| Enlace público | Storage privado + URL firmada. |
| Denegación por tamaño | Límites y body limit. |

---

## 9.8 Protección de datos personales

### Principios LOPDP

- **Legalidad y transparencia:** aviso de privacidad.
- **Finalidad:** solo para el concurso.
- **Minimización:** solo datos exigidos.
- **Conservación:** plazo definido.
- **Seguridad:** cifrado y control de acceso.
- **Confidencialidad:** acceso por necesidad.

### Datos sensibles

- Cédula, datos de pareja, inhabilidades, documentos.
- Se cifran en tránsito (TLS) y en reposo (disco/DB gestionada).
- Se redactan en logs.
- Se accede con permisos mínimos.
- Se auditan las consultas administrativas futuras.

### Consentimiento

- Cinco textos del RF.
- Versión de términos.
- Fecha, IP y user agent.
- Se almacena en `consentimiento`.
- Se incluye en el PDF.

---

## 9.9 Auditoría

Se auditan:

- Registro, validación de identidad, login, bloqueo, recuperación.
- Creación y cambios de postulación.
- Carga/eliminación de documentos.
- Cambio de cargo y elegibilidad.
- Consentimiento y envío.
- Accesos denegados.

Cada evento incluye `trace_id`, actor, IP, user agent y datos antes/después redactados.

---

## 9.10 Infraestructura

- PostgreSQL y Redis en red privada, sin exposición a Internet.
- TLS en todas las comunicaciones.
- Secretos en variables de entorno o gestor de secretos.
- Contenedores con usuario no root.
- Escaneo de imágenes y dependencias.
- Backups cifrados.
- Logs centralizados con retención.
- Firewall y mínimos privilegios.

---

## 9.11 Mapeo OWASP ASVS

| Control | Implementación |
|---|---|
| V2 Autenticación | Argon2id, bloqueo, mensajes genéricos. |
| V3 Sesión | JWT corto, refresh rotativo, logout. |
| V4 Control de acceso | Pertenencia, RBAC, 403. |
| V5 Validación | DTO, Zod, reglas de dominio. |
| V6 Criptografía | TLS, Argon2id, tokens aleatorios. |
| V7 Errores/logs | RFC 7807, sin stack traces, logs redactados. |
| V8 Datos | Clasificación, cifrado, retención. |
| V9 Comunicaciones | HTTPS, CORS, HSTS. |
| V10 Archivos | Magic bytes, tamaño, storage privado. |
| V11 Lógica | Estado, idempotencia, transacciones. |
| V12 API | Rate limit, content-type, métodos. |
| V13 Configuración | Variables de entorno, headers. |
| V14 Dependencias | Escaneo y actualización. |

---

## 9.12 Pruebas de seguridad

| Prueba | Herramienta |
|---|---|
| Fuerza bruta y bloqueo | Script automatizado. |
| Autorización horizontal | Pruebas con dos usuarios. |
| Upload malicioso | Archivos falsos y PDF corruptos. |
| SQL injection | Payloads en campos. |
| XSS | Payloads en textos. |
| CSRF | Peticiones cruzadas. |
| Rate limit | k6. |
| Dependencias | `govulncheck`, `npm audit`. |
| Configuración | OWASP ZAP baseline. |

---

## 9.13 Respuesta a incidentes

1. Detectar (alerta o reporte).
2. Contener (bloquear cuenta, revocar tokens, aislar servicio).
3. Erradicar (corregir causa).
4. Recuperar (restaurar servicio).
5. Documentar (post-mortem).
6. Notificar si hay datos personales afectados, según LOPDP.

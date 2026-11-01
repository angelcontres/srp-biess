# 10. Anexos — SRS

## 10.1 Catálogo de mensajes

| Código | Mensaje al usuario | HTTP | Origen |
|---|---|---|---|
| `AUTH_IDENTIDAD_INVALIDA` | La información ingresada es incorrecta. | 422 | EF-11 |
| `AUTH_REGISTRO_BLOQUEADO` | Registro fallido, inténtelo nuevamente en 10 minutos. | 423 | EF-11 |
| `AUTH_CREDENCIALES_INVALIDAS` | Sus credenciales de ingreso son incorrectas. | 401 | EF-11 |
| `AUTH_LOGIN_BLOQUEADO` | Cuenta bloqueada temporalmente. Intente nuevamente en 10 minutos. | 423 | EF-11 |
| `AUTH_CUENTA_INACTIVA` | Su cuenta no se encuentra habilitada. | 403 | EF-11 |
| `AUTH_REQUEST_INVALIDO` | La solicitud no tiene el formato esperado. | 400 | EF-11 |
| `VALIDATION_ERROR` | La solicitud contiene campos inválidos. | 422 | Transversal |
| `AUTH_TOKEN_EXPIRADO` | Su sesión expiró. Inicie sesión nuevamente. | 401 | EF-11 |
| `AUTH_RECUPERACION_EXPIRADA` | El enlace de recuperación expiró. Solicite uno nuevo. | 410 | EF-11 |
| `AUTH_CORREO_DUPLICADO` | El correo electrónico ya se encuentra registrado. | 409 | EF-11 |
| `AUTH_CEDULA_DUPLICADA` | La cédula ya se encuentra registrada. | 409 | EF-11 |
| `POS_FASE_CERRADA` | La fase de postulación no se encuentra habilitada. | 409 | EF-12 |
| `POS_YA_EXISTE` | Ya existe una postulación vigente para esta convocatoria. | 409 | EF-12 |
| `POS_NACIONALIDAD_NO_ECUATORIANA` | Su postulación no puede continuar porque la nacionalidad registrada no es ecuatoriana. | 422 | EF-12 |
| `POS_NO_ELEGIBLE_AFI` | Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS. | 422 | EF-12 |
| `POS_NO_ELEGIBLE_JUB` | Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación. | 422 | EF-12 |
| `POS_NO_ELEGIBLE_EMP` | Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador. | 422 | EF-12 |
| `POS_INCOMPLETA` | La postulación tiene campos pendientes. Revise las secciones marcadas. | 422 | EF-12 |
| `POS_CERRADA` | La postulación ya fue enviada y no puede modificarse. | 409 | EF-12 |
| `POS_CONFLICTO_VERSION` | La postulación fue modificada en otra sesión. Recargue la página. | 409 | EF-12 |
| `POS_YA_ENVIADA` | La postulación ya fue enviada. | 200 | EF-12 |
| `DOC_TIPO_INVALIDO` | Solo se permiten archivos PDF. | 415 | Documentos |
| `DOC_TAMANO_EXCEDIDO` | El archivo supera el tamaño máximo permitido. | 413 | Documentos |
| `DOC_TOTAL_EXCEDIDO` | Ha superado el límite total de 50 MB. | 413 | Documentos |
| `DOC_NO_ENCONTRADO` | El documento solicitado no existe. | 404 | Documentos |
| `INTEGRACION_NO_DISPONIBLE` | No es posible validar la información en este momento. Intente nuevamente. | 503 | Integraciones |
| `SISTEMA_ERROR` | Ocurrió un error inesperado. Contacte a soporte con el código de referencia. | 500 | Transversal |

---

## 10.2 Ejemplos de payloads

### 10.2.1 Validar identidad

```json
{
  "cedula": "0102030405",
  "tipoValidacion": "FECHA_EXPEDICION",
  "valor": "15/06/2018"
}
```

Respuesta exitosa:

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

### 10.2.2 Error de validación

```json
{
  "type": "https://api.biess.gob.ec/errors/validation",
  "title": "Error de validación",
  "status": 422,
  "code": "POS_INCOMPLETA",
  "detail": "La postulación tiene campos pendientes.",
  "traceId": "8f3c2a1e-9b7d-4c2a-8f1e-123456789abc",
  "errors": [
    { "field": "inhabilidades", "message": "Debe responder las 13 preguntas." },
    { "field": "formacionAcademica", "message": "Debe registrar al menos una formación." }
  ]
}
```

### 10.2.3 Enviar postulación

```json
{
  "aceptaConsentimiento": true,
  "versionTerminos": "2026-10-31"
}
```

Respuesta:

```json
{
  "codigoUnico": "BIESS-2026-AFI-000123",
  "estado": "ENVIADA",
  "fechaEnvio": "2026-10-31T16:42:00-05:00",
  "pdfUrl": "/api/v1/postulaciones/550e8400-e29b-41d4-a716-446655440000/pdf"
}
```

---

## 10.3 Decisiones pendientes

| ID | Tema | Decisión requerida | Responsable |
|---|---|---|---|
| PEND-01 | Política de contraseña definitiva | Confirmar longitud y complejidad. | SDNGTH / TI |
| PEND-02 | Retención de datos | Confirmar plazo legal y procedimiento de eliminación. | Asesoría Legal |
| PEND-03 | Efecto de inhabilidades 1.9.2 a 1.9.13 | Definir si alguna es bloqueante además de 1.9.1. | SDNGTH / Legal |
| PEND-04 | Respaldo organizacional negativo | Confirmar si la falta de respaldo impide la postulación. | SDNGTH |
| PEND-05 | Cargo único | Confirmar si un postulante puede postular a más de un cargo. | SDNGTH |
| PEND-06 | Firma electrónica | Confirmar si el PDF firmado manuscrito se carga después. | SDNGTH |
| PEND-07 | Notificaciones adicionales | Confirmar correos de bienvenida y cambios de estado. | SDNGTH |
| PEND-08 | Integraciones reales | Confirmar endpoints, convenios y esquemas de Registro Civil e IESS. | TI |

---

## 10.4 Checklist de revisión del SRS

- [ ] Todos los requisitos tienen ID único.
- [ ] Todos los `Must` tienen criterio de aceptación.
- [ ] Las reglas de negocio están vinculadas a requisitos.
- [ ] Los requisitos no funcionales son medibles.
- [ ] Las interfaces externas están identificadas.
- [ ] Los datos personales están clasificados.
- [ ] El alcance excluido está explícito.
- [ ] La trazabilidad está completa.
- [ ] Los mensajes al usuario están centralizados.
- [ ] Las decisiones pendientes están registradas.

---

## 10.5 Glosario complementario

| Término | Definición |
|---|---|
| **Backoff exponencial** | Estrategia de reintento con espera creciente. |
| **Cache-aside** | Patrón en el que la aplicación consulta caché y, si falla, consulta la base y llena la caché. |
| **Circuit breaker** | Patrón que corta llamadas a un servicio que falla repetidamente. |
| **Idempotencia** | Propiedad por la cual repetir una operación no cambia el resultado. |
| **Outbox** | Patrón de encolado de eventos en la misma transacción de negocio. |
| **Rate limiting** | Limitación de solicitudes por unidad de tiempo. |
| **Refresh token** | Credencial de larga duración para renovar el acceso. |
| **Rollback** | Reversión de una transacción. |
| **TTL** | Tiempo de vida de una clave o token. |
| **Magic bytes** | Firma binaria que identifica el tipo real de un archivo. |

---

## 10.6 Referencias del anexo

- Requerimiento Funcional v1.4 — IESS.
- Resolución C.D. 701.
- ISO/IEC/IEEE 29148:2018.
- OWASP ASVS 4.0.3.
- RFC 7807.
- Ley Orgánica de Protección de Datos Personales (Ecuador).

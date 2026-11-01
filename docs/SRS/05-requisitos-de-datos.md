# 05. Requisitos de datos — SRS

## 5.1 Propósito

Este capítulo define **qué datos necesita el sistema**, su significado de negocio, sus reglas de calidad, su clasificación, su ciclo de vida y las condiciones de auditoría y retención. El modelo físico, los tipos, índices y DDL se detallan en `SDD/04-diseno-de-datos.md` y `SDD/16-anexos.md`.

---

## 5.2 Entidades de negocio

| Entidad | Descripción | Datos clave | Sensibilidad |
|---|---|---|---|
| **Postulante** | Persona registrada en el concurso. | Cédula, nombres, apellidos, correo, hash de contraseña, estado. | Alta |
| **Pareja del postulante** | Datos de cónyuge o conviviente cuando aplica. | Nombres, apellidos, cédula. | Alta |
| **Domicilio** | Ubicación y contacto del postulante. | País, provincia, cantón, dirección, teléfonos, correos. | Alta |
| **Familiar en IESS/BIESS** | Declaración de vínculo familiar. | Nombres, parentesco, institución, área, cargo. | Alta |
| **Convocatoria** | Proceso concursal y su fase. | Código, fechas, bandera de postulación. | Media |
| **Cargo** | Puesto al que se postula. | Código, nombre, sector. | Pública |
| **Postulación** | Expediente del postulante para un cargo. | Código único, estado, fechas, consentimiento. | Alta |
| **Respaldo organizacional** | Declaración de respaldo y documento. | Tiene respaldo, nombre, documento. | Alta |
| **Formación académica** | Títulos y registros SENESCYT. | Nivel, institución, título, área, registro. | Alta |
| **Experiencia profesional** | Trayectoria laboral. | Cargo, institución, fechas, tiempo calculado. | Alta |
| **Publicación** | Producción académica. | Título, tipo, medio, fecha. | Media |
| **Capacitación** | Cursos, seminarios y talleres. | Tipo, nombre, institución, horas. | Media |
| **Inhabilidad** | Respuesta a pregunta legal. | Pregunta, respuesta, datos de remoción. | Muy alta |
| **Documento** | PDF de respaldo. | Tipo, nombre, hash, tamaño, ruta. | Alta |
| **Notificación** | Correo encolado/enviado. | Destinatario, plantilla, estado, intentos. | Media |
| **Auditoría** | Evidencia de eventos y cambios. | Actor, acción, entidad, antes/después. | Alta |
| **Catálogo geográfico** | Países, provincias y cantones. | Código, nombre, relación padre. | Pública |

---

## 5.3 Diccionario de datos de negocio

### 5.3.1 Postulante

| Atributo | Tipo lógico | Obligatorio | Regla |
|---|---|---|---|
| Cédula | Texto de 10 dígitos | Sí | Única; validada con algoritmo ecuatoriano; no editable. |
| Nombres | Texto | Sí | Proveniente de Registro Civil. |
| Apellidos | Texto | Sí | Proveniente de Registro Civil. |
| Género | Catálogo | Sí | Proveniente de Registro Civil. |
| Fecha de nacimiento | Fecha | Sí | Válida, no futura. |
| Nacionalidad | Catálogo de países | Sí | Debe ser Ecuador para continuar. |
| Estado civil | Catálogo | Sí | Determina datos de pareja. |
| Correo | Texto | Sí | Único, formato válido. |
| Contraseña | Secreto | Sí | Nunca se almacena en claro; hash Argon2id. |
| Estado | Catálogo | Sí | `REGISTRADO`, `ACTIVO`, `BLOQUEADO`, `INACTIVO`. |

### 5.3.2 Postulación

| Atributo | Tipo lógico | Obligatorio | Regla |
|---|---|---|---|
| Código único | Texto | Al enviar | Formato `BIESS-{AÑO}-{SECTOR}-{SECUENCIA}`; único. |
| Postulante | Referencia | Sí | Dueño de la postulación. |
| Convocatoria | Referencia | Sí | Convocatoria vigente. |
| Cargo | Referencia | Sí | Uno de los tres cargos activos. |
| Estado | Catálogo | Sí | `BORRADOR`, `LISTA_PARA_ENVIO`, `ENVIADA`, `EN_REVISION`, `NO_ELEGIBLE`, `CERRADA`. |
| Fecha de creación | Fecha-hora | Sí | Zona horaria institucional. |
| Fecha de envío | Fecha-hora | Al enviar | No puede ser anterior a la creación. |
| Consentimiento | Booleano + metadatos | Al enviar | Debe ser verdadero; versión, fecha, IP. |

### 5.3.3 Documento

| Atributo | Tipo lógico | Obligatorio | Regla |
|---|---|---|---|
| Tipo | Catálogo | Sí | Respaldo organizacional, título, experiencia, publicación, capacitación, formulario. |
| Nombre original | Texto | Sí | Se conserva para el usuario. |
| Nombre almacenado | UUID + `.pdf` | Sí | No se usa el nombre original en disco. |
| MIME | Texto | Sí | `application/pdf`. |
| Tamaño | Entero | Sí | ≤ 4 MB o ≤ 2 MB según tipo; total ≤ 50 MB. |
| Hash SHA-256 | Texto | Sí | Integridad y deduplicación. |
| Ruta | Texto | Sí | Fuera del directorio público. |
| Estado | Catálogo | Sí | `PENDIENTE`, `VALIDADO`, `RECHAZADO`. |

### 5.3.4 Inhabilidad

| Atributo | Tipo lógico | Obligatorio | Regla |
|---|---|---|---|
| Pregunta | Referencia | Sí | Catálogo de 13 preguntas. |
| Respuesta | Booleano | Sí | Sí/No. |
| Institución de remoción | Texto | Condicional | Obligatorio si la pregunta 1.9.7 es “Sí”. |
| Cargo de remoción | Texto | Condicional | Obligatorio si la pregunta 1.9.7 es “Sí”. |
| Fecha de finalización | Fecha | Condicional | Obligatorio si la pregunta 1.9.7 es “Sí”. |
| Motivo | Texto | Condicional | Obligatorio si la pregunta 1.9.7 es “Sí”. |

---

## 5.4 Reglas de calidad de datos

| ID | Regla | Verificación |
|---|---|---|
| RCD-01 | La cédula debe ser única en `postulante`. | Restricción única + prueba. |
| RCD-02 | El correo debe ser único y válido. | Restricción única + validación. |
| RCD-03 | Un cantón debe pertenecer a la provincia seleccionada. | FK compuesta o validación de servicio. |
| RCD-04 | Las fechas de experiencia deben cumplir inicio ≤ fin. | Validación de dominio. |
| RCD-05 | El tiempo calculado debe ser reproducible. | Prueba con casos borde. |
| RCD-06 | Los documentos deben ser PDF reales, no solo extensión `.pdf`. | Validación de magic bytes. |
| RCD-07 | El total de documentos por postulante no debe superar 50 MB. | Suma transaccional. |
| RCD-08 | Una postulación enviada no puede modificarse. | Regla de estado + prueba. |
| RCD-09 | Los catálogos no deben tener duplicados por padre. | Restricción única. |
| RCD-10 | Los eventos de auditoría no deben poder editarse ni eliminarse. | Permisos de base de datos. |

---

## 5.5 Clasificación de la información

| Clase | Ejemplos | Controles |
|---|---|---|
| **Pública** | Catálogos de países, provincias, cantones; nombres de cargos. | Integridad; sin restricción de lectura. |
| **Interna** | Configuración, métricas, logs no sensibles. | Acceso restringido a operadores. |
| **Confidencial** | Correo, teléfono, dirección, formación, experiencia. | Cifrado en tránsito; acceso por dueño y roles autorizados. |
| **Sensible** | Cédula, datos de pareja, inhabilidades, documentos de identidad. | Cifrado en reposo; minimización; auditoría; acceso estricto. |

---

## 5.6 Volumen y crecimiento estimado

| Entidad | Volumen inicial | Crecimiento | Observación |
|---|---|---|---|
| Postulantes | 1 000 | 1 000 por convocatoria | Depende de la convocatoria. |
| Postulaciones | 1 000 | 1 por postulante | Una vigente por convocatoria. |
| Documentos | 5 000 | 5–10 por postulación | PDF de 0.5–4 MB. |
| Formación | 3 000 | 1–5 por postulación | — |
| Experiencia | 5 000 | 1–10 por postulación | — |
| Auditoría | 50 000 | Alto | Crecimiento continuo; particionar por año si es necesario. |
| Catálogos | ~250 países, 24 provincias, ~220 cantones | Bajo | Cacheables. |

---

## 5.7 Retención y ciclo de vida

| Dato | Retención propuesta | Acción al vencer |
|---|---|---|
| Postulación y documentos | 5 años después del cierre del concurso. | Archivo histórico o eliminación segura según política. |
| Datos de cuenta | Mientras la cuenta esté activa y hasta 1 año después. | Anonimización o eliminación. |
| Tokens de recuperación | 30 minutos. | Eliminación automática. |
| Tokens de refresh | 8 horas o hasta logout. | Revocación. |
| Logs de aplicación | 90 días. | Rotación y eliminación. |
| Auditoría | 7 años o según normativa. | Archivo inmutable. |
| Borradores abandonados | 90 días sin actividad. | Notificación y eliminación. |

> **PENDIENTE:** la SDNGTH debe confirmar los plazos definitivos de retención y eliminación. Los valores anteriores son una propuesta técnica alineada con buenas prácticas y deben validarse con asesoría legal.

---

## 5.8 Auditoría y trazabilidad

Se debe auditar como mínimo:

- Registro, validación de identidad, login exitoso/fallido, bloqueo y recuperación.
- Creación de borrador y cada guardado de sección.
- Creación, edición y eliminación de ítems repetibles.
- Carga, reemplazo y eliminación de documentos.
- Cambio de cargo y resultado de elegibilidad.
- Aceptación de consentimiento.
- Envío, generación de código y PDF.
- Envío de correo y reintentos.
- Accesos denegados y errores de autorización.

Cada evento debe incluir: `actor_id`, `actor_tipo`, `accion`, `entidad`, `entidad_id`, `datos_antes`, `datos_despues`, `ip`, `user_agent`, `trace_id`, `created_at`.

---

## 5.9 Migraciones y datos semilla

- Las migraciones deben versionarse y ejecutarse en orden.
- Los catálogos de países, provincias y cantones se cargan con un script semilla idempotente.
- Las preguntas de inhabilidades se cargan con su texto, orden, efecto y vigencia.
- Los cargos se cargan con sus códigos `AFI`, `JUB`, `EMP`.
- Los datos de prueba deben ser ficticios y no contener información real de personas.

---

## 5.10 Privacidad y protección de datos

1. **Minimización:** solo se solicita lo exigido por el RF.
2. **Finalidad:** los datos se usan exclusivamente para el concurso.
3. **Consentimiento:** se registra antes del envío, con versión de términos.
4. **Seguridad:** cifrado en tránsito; control de acceso; auditoría.
5. **Derechos:** el postulante puede solicitar acceso, rectificación, eliminación u oposición según la LOPDP.
6. **Transferencias:** las integraciones con Registro Civil e IESS deben estar amparadas en convenios y bases legales.
7. **Retención:** se aplica la política institucional validada.

El detalle de controles se encuentra en `SDD/09-diseno-seguridad.md`.

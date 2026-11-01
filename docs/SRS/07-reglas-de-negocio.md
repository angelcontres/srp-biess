# 07. Reglas de negocio — SRS

## 7.1 Convenciones

Cada regla tiene un identificador `RN-xx`, una declaración verificable, su origen y el lugar donde se aplica. El origen se clasifica así:

| Tipo | Significado |
|---|---|
| **Fuente** | Deriva directamente del RF v1.4 o de la rúbrica. |
| **Derivada** | Necesaria para seguridad, integridad o consistencia del diseño. |
| **Propuesta** | Decisión razonable del equipo que debe validarse con la SDNGTH. |

Las reglas **Fuente** son obligatorias. Las **Derivadas** se justifican en el SDD. Las **Propuestas** se marcan como `[PENDIENTE]` y no deben presentarse como normativa definitiva.

---

## 7.2 Catálogo de reglas

| ID | Regla | Tipo | Aplicación |
|---|---|---|---|
| RN-01 | La nacionalidad del postulante debe ser ecuatoriana; si no lo es, la postulación no puede continuar y se cierra. | Fuente | Dominio / API / UI |
| RN-02 | La identidad debe validarse contra Registro Civil con cédula + fecha de expedición o código dactilar. | Fuente | Servicio de autenticación |
| RN-03 | La cédula debe ser única por postulante y no puede modificarse después del registro. | Fuente / Derivada | Base de datos |
| RN-04 | El correo electrónico debe ser único y válido. | Derivada | Base de datos / validación |
| RN-05 | La contraseña debe tener mínimo 10 caracteres, máximo 128, con al menos tres de: mayúscula, minúscula, número y símbolo. | Propuesta | Servicio de autenticación |
| RN-06 | Tras 3 intentos fallidos consecutivos de validación de identidad, el registro se bloquea 10 minutos. | Fuente | Redis / servicio |
| RN-07 | Tras 3 intentos fallidos consecutivos de login, la cuenta se bloquea 10 minutos. | Derivada | Redis / servicio |
| RN-08 | Los mensajes de bloqueo deben indicar el tiempo restante y no revelar datos sensibles. | Fuente / Derivada | API / UI |
| RN-09 | El token de recuperación dura 30 minutos, es de un solo uso y, al usarse, revoca las sesiones previas. | Derivada | Redis / servicio |
| RN-10 | El token de acceso dura 15 minutos; el refresh dura 8 horas y es rotativo. | Derivada | Servicio de sesión |
| RN-11 | El registro y el formulario solo se habilitan si la fase de postulación está activa. | Fuente | Configuración / API |
| RN-12 | Un postulante solo puede tener un borrador vigente por convocatoria. | Derivada | Base de datos |
| RN-13 | Un postulante solo puede tener una postulación vigente; el cargo puede cambiarse mientras esté en borrador y no haya fallado la elegibilidad. | Propuesta | Dominio |
| RN-14 | Cada sección se guarda de forma independiente y transaccional. | Fuente | Servicio / UI |
| RN-15 | Si el estado civil es “Casado/a” o “Unión de hecho”, se exigen nombres, apellidos y cédula de la pareja. | Fuente | Dominio / UI |
| RN-16 | Para Ecuador, provincia y cantón se seleccionan de catálogo; para el extranjero, se ingresan manualmente. | Fuente | Dominio / UI |
| RN-17 | El correo principal se precarga y no es editable; el correo secundario es opcional y no puede ser igual al principal. | Derivada | Dominio / UI |
| RN-18 | Si el postulante declara familiares en IESS/BIESS, debe registrar al menos uno con todos los campos. | Fuente | Dominio / UI |
| RN-19 | El cargo se selecciona una sola vez por postulación y requiere confirmación explícita. | Fuente / Derivada | Dominio / UI |
| RN-20 | La elegibilidad se valida según el sector: afiliado, jubilado o empleador. Si no cumple, la postulación se cierra. | Fuente | Servicio de elegibilidad |
| RN-21 | Si el postulante declara respaldo organizacional, debe indicar el nombre y adjuntar un PDF de máximo 4 MB. | Fuente | Dominio / documentos |
| RN-22 | La formación académica debe usar los niveles y áreas del catálogo; cada registro requiere PDF de máximo 2 MB. | Fuente | Dominio / documentos |
| RN-23 | La experiencia debe tener fechas coherentes; el sistema calcula años, meses y días por registro y muestra el total. | Fuente | Servicio de cálculo |
| RN-24 | Si declara publicaciones, debe registrar al menos una con PDF de máximo 2 MB. | Fuente | Dominio / documentos |
| RN-25 | La capacitación es opcional; si se registra, todos los campos y el PDF de máximo 2 MB son obligatorios. | Fuente | Dominio / documentos |
| RN-26 | Las 13 preguntas de inhabilidades son obligatorias; la pregunta 1.9.1 bloquea si se responde “No”; las demás generan alerta de revisión. | Fuente / Propuesta | Dominio |
| RN-27 | Solo se aceptan PDF reales; los límites son 4 MB para respaldo organizacional y 2 MB para los demás documentos. | Fuente | Servicio de documentos |
| RN-28 | El total de documentos por postulante no puede superar 50 MB. | Fuente | Servicio de documentos |
| RN-29 | Para guardar la postulación, todas las secciones obligatorias deben estar completas y sin bloqueos. | Fuente | Servicio de postulación |
| RN-30 | El envío requiere la aceptación de los cinco consentimientos informados. | Fuente | Dominio / UI |
| RN-31 | Una postulación enviada es inmutable; no se puede editar ni eliminar. | Fuente | Dominio / API |
| RN-32 | El código único tiene formato `BIESS-{AÑO}-{SECTOR}-{SECUENCIA}` y el PDF incluye todos los datos, consentimiento y espacio de firma. | Fuente | Servicio de envío |
| RN-33 | Los eventos críticos y cambios deben auditarse con actor, fecha, IP y datos antes/después. | Derivada | Servicio de auditoría |
| RN-34 | Las fechas de interfaz se muestran `DD/MM/AAAA`; la API y la base usan ISO 8601. | Derivada | UI / API / DB |
| RN-35 | El envío es idempotente: un segundo intento devuelve el mismo código y no duplica la postulación. | Derivada | Servicio / Redis |
| RN-36 | Las escrituras usan control de concurrencia optimista con `version`; un conflicto devuelve HTTP 409. | Derivada | Repositorio / API |
| RN-37 | Los correos se reintentan hasta 5 veces con backoff; un fallo de correo no revierte la postulación. | Derivada | Worker / outbox |
| RN-38 | Todo acceso a una postulación, ítem o documento debe validar que pertenece al postulante autenticado. | Derivada | Servicio / API |
| RN-39 | Los catálogos se cachean con TTL de 24 h y se invalidan por versión. | Derivada | Redis |
| RN-40 | La retención de datos sigue la política institucional; propuesta: 5 años después del cierre del concurso. | Propuesta | Operación |
| RN-41 | Toda cédula ecuatoriana debe cumplir validación de 10 dígitos, provincia (01-24 o 30), tercer dígito < 6 y residuo de algoritmo Módulo 10 antes de consultar Registro Civil. | Derivada | Validador / UI / API |
| RN-42 | El pago de arancel con tarjeta valida algoritmo de Luhn, fecha futura y CVV; la pasarela opera mediante interfaz mockeable pero preparada para producción. | Derivada | Servicio de Pago / Gateway |
| RN-43 | Aprobado el pago, se emite factura con clave de acceso de 49 dígitos del SRI y se envía automáticamente por correo con su PDF adjunto. | Derivada | Facturación / Outbox |

---

## 7.3 Tabla de decisión — Elegibilidad por cargo

| Cargo | Condición consultada | Resultado | Acción |
|---|---|---|---|
| Afiliados (`AFI`) | `esAfiliado = true` | Elegible | Mostrar pregunta de respaldo y continuar. |
| Afiliados (`AFI`) | `esAfiliado = false` | No elegible | Mensaje de cierre; estado `NO_ELEGIBLE`; auditoría. |
| Jubilados (`JUB`) | `esJubilado = true` | Elegible | Mostrar pregunta de respaldo y continuar. |
| Jubilados (`JUB`) | `esJubilado = false` | No elegible | Mensaje de cierre; estado `NO_ELEGIBLE`; auditoría. |
| Empleadores (`EMP`) | `esEmpleador = true` | Elegible | Mostrar pregunta de respaldo y continuar. |
| Empleadores (`EMP`) | `esEmpleador = false` | No elegible | Mensaje de cierre; estado `NO_ELEGIBLE`; auditoría. |
| Cualquiera | Servicio externo no disponible | Indeterminado | HTTP 503; no marcar `NO_ELEGIBLE`; permitir reintento. |

### Mensajes institucionales

| Caso | Mensaje |
|---|---|
| AFI no elegible | “Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de Afiliados ante el Directorio del BIESS, NO puede continuar, por cuanto no mantiene la calidad de afiliado al IESS.” |
| JUB no elegible | “Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación.” |
| EMP no elegible | “Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador.” |

---

## 7.4 Tabla de decisión — Inhabilidades

| Pregunta | Respuesta | Efecto | Acción |
|---|---|---|---|
| 1.9.1 ¿Es ecuatoriano/a y está en pleno goce de los derechos de participación política? | No | **Bloqueante** | No permite guardar ni enviar; muestra mensaje. |
| 1.9.1 | Sí | Sin alerta | Continúa. |
| 1.9.2 ¿Inhabilitado para ejercer el comercio? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.3 ¿En mora con el Estado o instituciones bajo control de la Superintendencia de Bancos? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.4 ¿Vínculo con sistema financiero, seguros, bolsas, casas de valores o fondos? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.5 ¿Funcionario o empleado del IESS-BIESS? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.6 ¿Interés propio o representación de terceros en aseguradoras o seguridad social? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.7 ¿Removido por Organismo de Control en los últimos 5 años? | Sí | Alerta + campos | Exige institución, cargo, fecha y motivo; marca revisión. |
| 1.9.8 ¿Obligaciones pendientes con el SRI? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.9 ¿Castigo de obligaciones por institución financiera en los últimos 5 años? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.10 ¿Litigando contra el IESS-BIESS? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.11 ¿Procesado, enjuiciado o condenado por corrupción o crimen organizado? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.12 ¿Responsabilidades en firme de la Contraloría? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |
| 1.9.13 ¿Registrado en la base de la UAFE? | Sí | Alerta | Marca `requiere_revision`; permite continuar. |

> **PENDIENTE:** el efecto de las preguntas 1.9.2 a 1.9.13 debe ser validado por la SDNGTH y Asesoría Legal. La propuesta del equipo es **no rechazar automáticamente** salvo la pregunta 1.9.1, y dejar las demás como alertas para revisión humana. Esto evita decisiones automáticas sobre situaciones que requieren análisis legal.

---

## 7.5 Reglas de cálculo de experiencia

### Fórmula por registro

1. Si `fecha_fin` es nula y el postulante marcó “Actualmente”, se usa la fecha actual.
2. Si `fecha_fin < fecha_inicio`, se rechaza.
3. Se calcula la diferencia de calendario: años, meses y días, tomando como referencia el día de inicio.
4. Ejemplo: del `15/01/2020` al `10/03/2023` = **3 años, 1 mes y 23 días**.
5. El resultado se almacena en `tiempo_anios`, `tiempo_meses`, `tiempo_dias`.

### Total acumulado

- Se suman los días de cada período y se convierten a años/meses/días con la convención **365 días/año y 30 días/mes**, solo para visualización.
- Si hay períodos superpuestos, se muestra una advertencia “Existen períodos que se superponen. Verifique la información”.
- El total es **referencial**; la SDNGTH valida la experiencia según la normativa del concurso.

### Casos borde

| Caso | Resultado |
|---|---|
| 29/02/2020 → 28/02/2021 | 1 año, 0 meses, 0 días (ajuste de bisiesto). |
| 31/01/2023 → 28/02/2023 | 0 años, 1 mes, 0 días (ajuste de fin de mes). |
| Inicio = fin | 0 años, 0 meses, 0 días; se permite si la experiencia es válida. |
| Fecha futura | Error de validación. |

---

## 7.6 Reglas de documentos

| Tipo | Tamaño máximo | Obligatorio |
|---|---|---|
| Respaldo organizacional | 4 MB | Solo si declara respaldo “Sí”. |
| Título / formación | 2 MB | Sí, por cada formación. |
| Experiencia | 2 MB | Sí, por cada experiencia. |
| Publicación | 2 MB | Sí, por cada publicación. |
| Capacitación | 2 MB | Sí, por cada capacitación. |
| Formulario generado | Sin límite práctico | Generado por el sistema. |
| **Total por postulante** | **50 MB** | Siempre. |

Validaciones:

- Extensión `.pdf` y MIME real `application/pdf` (magic bytes `%PDF-`).
- Nombre original saneado; nombre en disco UUID.
- Hash SHA-256 para integridad.
- Almacenamiento fuera del directorio público.
- Descarga mediante URL firmada.

---

## 7.7 Reglas de mensajes y errores

| Código | Mensaje | HTTP | Acción UI |
|---|---|---|---|
| `AUTH_IDENTIDAD_INVALIDA` | “La información ingresada es incorrecta.” | 422 | Limpiar dato de validación; mantener cédula. |
| `AUTH_REGISTRO_BLOQUEADO` | “Registro fallido, inténtelo nuevamente en 10 minutos.” | 423 | Mostrar cuenta regresiva; deshabilitar formulario. |
| `AUTH_CREDENCIALES_INVALIDAS` | “Sus credenciales de ingreso son incorrectas.” | 401 | Limpiar contraseña; mantener cédula. |
| `AUTH_LOGIN_BLOQUEADO` | “Cuenta bloqueada temporalmente. Intente nuevamente en 10 minutos.” | 423 | Mostrar tiempo restante. |
| `POS_NACIONALIDAD_NO_ECUATORIANA` | “Su postulación no puede continuar porque la nacionalidad registrada no es ecuatoriana.” | 422 | Bloquear y cerrar postulación. |
| `POS_NO_ELEGIBLE_AFI` | Mensaje institucional AFI. | 422 | Cerrar postulación. |
| `POS_NO_ELEGIBLE_JUB` | Mensaje institucional JUB. | 422 | Cerrar postulación. |
| `POS_NO_ELEGIBLE_EMP` | Mensaje institucional EMP. | 422 | Cerrar postulación. |
| `POS_INCOMPLETA` | “La postulación tiene campos pendientes.” | 422 | Mostrar lista y enlaces. |
| `POS_CERRADA` | “La postulación ya fue enviada y no puede modificarse.” | 409 | Modo lectura. |
| `DOC_TIPO_INVALIDO` | “Solo se permiten archivos PDF.” | 415 | Rechazar archivo. |
| `DOC_TAMANO_EXCEDIDO` | “El archivo supera el tamaño máximo permitido.” | 413 | Rechazar archivo. |
| `DOC_TOTAL_EXCEDIDO` | “Ha superado el límite total de 50 MB.” | 413 | Rechazar archivo. |
| `INTEGRACION_NO_DISPONIBLE` | “No es posible validar la información en este momento. Intente nuevamente.” | 503 | Reintentar. |

El catálogo completo se encuentra en `10-anexos.md`.

---

## 7.8 Aplicación de las reglas

| Capa | Responsabilidad |
|---|---|
| **UI** | Validaciones de formato y feedback inmediato; nunca es la única defensa. |
| **Handler** | Validación del contrato, autenticación y autorización. |
| **Service** | Reglas de negocio, elegibilidad, cálculo, consentimiento y estado. |
| **Repository** | Restricciones, transacciones, unicidad e integridad referencial. |
| **Base de datos** | `UNIQUE`, `FOREIGN KEY`, `CHECK`, índices y triggers de auditoría si aplica. |
| **Redis** | Bloqueos, rate limiting, caché e idempotencia. |

La duplicación de validaciones es intencional: la UI mejora la experiencia; el servidor garantiza la integridad.

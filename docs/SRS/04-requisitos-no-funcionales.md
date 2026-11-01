# 04. Requisitos no funcionales — SRS

## 4.1 Categorías

Los requisitos no funcionales se clasifican según ISO/IEC 25010 y se complementan con las restricciones de la rúbrica del examen:

| Categoría | Prefijo | Enfoque |
|---|---|---|
| Arquitectura y mantenibilidad | `RNF-ARQ` | Capas, inyección de dependencias, modularidad. |
| Rendimiento y capacidad | `RNF-REN` | Latencia, concurrencia, throughput. |
| Disponibilidad y continuidad | `RNF-DIS` | Disponibilidad, respaldo, recuperación. |
| Seguridad | `RNF-SEG` | Autenticación, autorización, OWASP, datos. |
| Usabilidad y accesibilidad | `RNF-USA` | Responsive, WCAG, mensajes. |
| Compatibilidad | `RNF-COM` | Navegadores y dispositivos. |
| Observabilidad | `RNF-OBS` | Logs, métricas, trazas, salud. |
| Portabilidad y despliegue | `RNF-POR` | Docker, variables de entorno, CI/CD. |
| Interoperabilidad | `RNF-INT` | Integraciones, timeouts, degradación. |
| Legalidad y privacidad | `RNF-LEG` | LOPDP, consentimiento, retención. |
| Calidad de datos | `RNF-DAT` | Validación, integridad, duplicados. |
| Escalabilidad | `RNF-ESC` | Crecimiento horizontal. |

---

## 4.2 Catálogo de requisitos no funcionales

### 4.2.1 Arquitectura y mantenibilidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-ARQ-01 | El backend debe organizarse en capas `handler → service → repository`, con dependencias invertidas e inyección explícita. | Revisión de código: ningún handler contiene lógica de negocio; los servicios dependen de interfaces. | Must |
| RNF-ARQ-02 | El frontend debe ser una SPA modular por funcionalidades (`features`), con componentes reutilizables y estado separado. | Revisión de estructura; componentes con una sola responsabilidad. | Must |
| RNF-ARQ-03 | La lógica de negocio debe ser independiente de Fiber, PostgreSQL y Redis. | Pruebas unitarias de servicios sin infraestructura. | Must |
| RNF-ARQ-04 | El código debe cumplir `gofmt`, `go vet`, ESLint y Prettier. | Pipeline de CI sin errores. | Must |
| RNF-ARQ-05 | La cobertura de pruebas debe ser ≥ 70 % en servicios backend y ≥ 60 % en componentes críticos del frontend. | Reporte de cobertura. | Should |
| RNF-ARQ-06 | Las decisiones de arquitectura deben documentarse como ADR. | `SDD/14` completo. | Should |

### 4.2.2 Rendimiento y capacidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-REN-01 | Las operaciones CRUD deben responder en p95 < 300 ms con carga normal. | Prueba de carga con k6. | Must |
| RNF-REN-02 | Las consultas de catálogos cacheadas deben responder en p95 < 100 ms. | Medición con Redis habilitado. | Must |
| RNF-REN-03 | El sistema debe soportar al menos 200 usuarios concurrentes y 50 req/s en el entorno de prueba. | Prueba de carga sostenida 10 min. | Must |
| RNF-REN-04 | La carga inicial de la SPA debe ser < 2.5 s en una conexión 4G simulada. | Lighthouse / WebPageTest. | Should |
| RNF-REN-05 | El envío de una postulación debe completarse en < 5 s sin contar el correo. | Prueba de integración. | Must |
| RNF-REN-06 | La generación del PDF debe completarse en < 3 s para una postulación completa. | Prueba de rendimiento. | Should |

### 4.2.3 Disponibilidad y continuidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-DIS-01 | La disponibilidad mensual del servicio debe ser ≥ 99.5 %. | Monitoreo de uptime. | Should |
| RNF-DIS-02 | El RPO debe ser ≤ 15 min y el RTO ≤ 2 h. | Prueba de restauración. | Should |
| RNF-DIS-03 | El sistema debe tener endpoints de salud `live` y `ready`. | Verificación con balanceador. | Must |
| RNF-DIS-04 | Si Redis no está disponible, los catálogos deben consultarse en PostgreSQL y el bloqueo de autenticación debe fallar de forma segura. | Prueba de caos controlada. | Must |
| RNF-DIS-05 | Si un servicio externo no responde, el sistema debe mostrar un error comprensible y no perder datos ya guardados. | Prueba de integración con mocks de fallo. | Must |

### 4.2.4 Seguridad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-SEG-01 | El sistema debe cumplir OWASP ASVS nivel 2 en los flujos críticos. | Revisión y pruebas de seguridad. | Must |
| RNF-SEG-02 | Las contraseñas deben almacenarse con Argon2id (o bcrypt con costo ≥ 12) y nunca en claro. | Inspección de base de datos y código. | Must |
| RNF-SEG-03 | Las comunicaciones deben usar HTTPS/TLS 1.2+ en producción. | Configuración del servidor. | Must |
| RNF-SEG-04 | Los tokens de acceso deben ser JWT firmados y de corta duración; el refresh debe ser opaco, rotativo y almacenado de forma segura. | Prueba de expiración y revocación. | Must |
| RNF-SEG-05 | El sistema debe aplicar bloqueo tras 3 intentos fallidos y rate limiting por IP y por cédula. | Prueba automatizada. | Must |
| RNF-SEG-06 | Los archivos deben validarse por MIME real, extensión y tamaño; no deben ejecutarse ni servirse desde el dominio de la API. | Prueba con archivos maliciosos. | Must |
| RNF-SEG-07 | Los logs no deben contener contraseñas, tokens, códigos dactilares ni datos sensibles completos. | Revisión de logs. | Must |
| RNF-SEG-08 | Las rutas deben validar pertenencia: un postulante no puede acceder a la postulación de otro. | Prueba de autorización horizontal. | Must |
| RNF-SEG-09 | Las dependencias deben escanearse y no presentar vulnerabilidades críticas conocidas. | `govulncheck`, `npm audit`. | Should |

### 4.2.5 Usabilidad y accesibilidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-USA-01 | La interfaz debe ser responsive entre 360 px y 1920 px. | Pruebas en 3 anchos. | Must |
| RNF-USA-02 | Los flujos críticos deben cumplir WCAG 2.1 AA: contraste, foco, etiquetas y navegación por teclado. | Auditoría con axe y prueba manual. | Must |
| RNF-USA-03 | Cada formulario debe mostrar errores junto al campo y un resumen al inicio. | Revisión de UI. | Must |
| RNF-USA-04 | Las operaciones asíncronas deben mostrar `loading`, éxito y error. | Prueba de interacción. | Must |
| RNF-USA-05 | El formulario debe advertir antes de salir con cambios sin guardar. | Prueba manual. | Should |
| RNF-USA-06 | Los textos deben estar en español y usar lenguaje claro, sin jerga técnica. | Revisión de contenido. | Must |

### 4.2.6 Compatibilidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-COM-01 | Soporte para Chrome, Edge, Firefox y Safari en sus últimas 2 versiones. | Matriz de pruebas. | Must |
| RNF-COM-02 | La descarga y visualización de PDF debe funcionar en los navegadores soportados. | Prueba manual. | Must |
| RNF-COM-03 | La interfaz debe funcionar en resoluciones móviles y tablets. | Prueba responsive. | Must |

### 4.2.7 Observabilidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-OBS-01 | Los logs deben ser estructurados (JSON) e incluir `trace_id`, ruta, método, estado y latencia. | Revisión de logs. | Must |
| RNF-OBS-02 | El sistema debe exponer métricas de requests, errores, latencia, caché y bloqueos. | Endpoint `/metrics` o equivalente. | Should |
| RNF-OBS-03 | Las alertas deben dispararse ante caída del servicio, tasa de error > 5 % o fallos de correo repetidos. | Configuración de monitoreo. | Should |
| RNF-OBS-04 | Cada error debe devolver un `traceId` para soporte. | Prueba de error. | Must |

### 4.2.8 Portabilidad y despliegue

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-POR-01 | El entorno local debe levantarse con `docker-compose` (PostgreSQL, Redis, Mailpit, API y SPA). | `docker compose up` funcional. | Must |
| RNF-POR-02 | La configuración debe provenir de variables de entorno; no debe haber secretos en el repositorio. | Revisión de `.env.example`. | Must |
| RNF-POR-03 | Las migraciones deben ser versionadas y ejecutables de forma idempotente. | `migrate up` en base limpia. | Must |
| RNF-POR-04 | El build de la SPA debe ser estático y servible por Nginx/CDN. | `npm run build`. | Must |

### 4.2.9 Interoperabilidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-INT-01 | Las llamadas a Registro Civil e Historial Laboral deben tener timeout ≤ 3 s, reintentos limitados y circuito de corte. | Prueba con mocks lentos. | Must |
| RNF-INT-02 | Los adaptadores externos deben implementar interfaces sustituibles por simuladores. | Prueba con dobles. | Must |
| RNF-INT-03 | Los errores de integración deben mapearse a mensajes de usuario y códigos internos. | Prueba de contrato. | Must |

### 4.2.10 Legalidad y privacidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-LEG-01 | El sistema debe registrar consentimiento informado antes del envío. | Prueba de envío sin consentimiento. | Must |
| RNF-LEG-02 | Solo se deben recolectar datos necesarios para el concurso. | Revisión del formulario. | Must |
| RNF-LEG-03 | Los datos personales deben protegerse en tránsito y en reposo. | Configuración TLS y cifrado de volumen. | Must |
| RNF-LEG-04 | La retención debe seguir la política institucional; propuesta inicial: 5 años después del cierre del concurso. | Revisión con SDNGTH. | Should |
| RNF-LEG-05 | El postulante debe poder solicitar acceso, rectificación o eliminación según la LOPDP. | Procedimiento documentado. | Should |

### 4.2.11 Calidad de datos

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-DAT-01 | La cédula debe validarse con algoritmo ecuatoriano y ser única. | Prueba de duplicados. | Must |
| RNF-DAT-02 | El correo debe validarse y ser único. | Prueba de duplicados. | Must |
| RNF-DAT-03 | Las fechas deben ser válidas y coherentes (inicio ≤ fin, no futuras cuando corresponda). | Pruebas de borde. | Must |
| RNF-DAT-04 | Los catálogos deben ser consistentes: un cantón no puede pertenecer a otra provincia. | Prueba de integridad referencial. | Must |
| RNF-DAT-05 | Las operaciones multi-tabla deben ser atómicas. | Prueba de rollback. | Must |

### 4.2.12 Escalabilidad

| ID | Requisito | Métrica / verificación | Prioridad |
|---|---|---|---|
| RNF-ESC-01 | La API debe ser stateless y escalar horizontalmente. | Prueba con 2 réplicas. | Must |
| RNF-ESC-02 | La caché y la base de datos deben ser servicios externos al contenedor de la API. | Revisión de arquitectura. | Must |
| RNF-ESC-03 | El almacenamiento de archivos debe poder migrar de volumen local a S3 sin cambiar la lógica de negocio. | Prueba con adaptador alternativo. | Should |

---

## 4.3 Escenarios de atributos de calidad

| ID | Atributo | Estímulo | Respuesta esperada | Medida |
|---|---|---|---|---|
| QA-01 | Rendimiento | 200 usuarios consultan catálogos simultáneamente. | Redis responde y la API mantiene latencia. | p95 < 100 ms; error < 1 %. |
| QA-02 | Seguridad | Atacante prueba 10 contraseñas para una cédula. | Bloqueo tras 3 fallos y rate limiting. | 100 % de bloqueos; sin fuga de información. |
| QA-03 | Integridad | Falla la inserción del PDF después de actualizar el estado. | Rollback completo. | 0 postulaciones inconsistentes. |
| QA-04 | Disponibilidad | Redis cae durante 5 min. | Catálogos desde PostgreSQL; auth falla de forma segura. | Sin pérdida de datos; error comprensible. |
| QA-05 | Usabilidad | Postulante pierde conexión a mitad del formulario. | Datos guardados previamente se conservan; al volver, reanuda. | 0 pérdida de secciones guardadas. |
| QA-06 | Mantenibilidad | Se cambia el proveedor de correo. | Solo se reemplaza el adaptador. | 0 cambios en servicios de dominio. |
| QA-07 | Interoperabilidad | Registro Civil tarda 10 s. | Timeout a los 3 s y mensaje de reintento. | Sin bloqueo de la UI; sin consumo de intento. |
| QA-08 | Legalidad | Postulante no acepta consentimiento. | No puede enviar. | 100 % de bloqueo. |

---

## 4.4 Restricciones no funcionales

| Código | Restricción |
|---|---|
| RNF-RES-01 | El backend debe usar Go y Fiber o Gin. |
| RNF-RES-02 | El frontend debe usar React, Vite y Tailwind CSS. |
| RNF-RES-03 | La base de datos debe ser PostgreSQL. |
| RNF-RES-04 | Redis debe usarse para bloqueo y caché. |
| RNF-RES-05 | Los documentos deben ser PDF y respetar los límites del RF. |
| RNF-RES-06 | El sistema debe operar en zona horaria `America/Guayaquil`. |
| RNF-RES-07 | La interfaz debe estar en español. |

---

## 4.5 Verificación

| Método | Aplicación |
|---|---|
| **Inspección** | Arquitectura, código, configuración, logs, documentación. |
| **Análisis** | Modelo de datos, índices, consultas, amenazas. |
| **Demostración** | Flujos de registro, postulación y envío. |
| **Prueba** | Unitarias, integración, E2E, carga, seguridad y accesibilidad. |

La matriz de verificación se detalla en `08-trazabilidad-y-verificacion.md` y en `SDD/15-plan-de-pruebas.md`.

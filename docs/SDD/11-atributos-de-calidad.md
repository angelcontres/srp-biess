# 11. Atributos de calidad — SDD

## 11.1 Escenarios de calidad

| ID | Atributo | Fuente | Estímulo | Respuesta | Medida |
|---|---|---|---|---|---|
| QA-01 | Rendimiento | Postulantes | 200 usuarios consultan catálogos. | Redis responde; API estable. | p95 < 100 ms; error < 1 %. |
| QA-02 | Seguridad | Atacante | 10 intentos de login. | Bloqueo tras 3; rate limit. | 100 % bloqueado. |
| QA-03 | Integridad | Falla de DB | Error al insertar PDF. | Rollback total. | 0 inconsistencias. |
| QA-04 | Disponibilidad | Infraestructura | Redis cae 5 min. | Catálogos desde DB; auth fail-closed. | Sin pérdida; error claro. |
| QA-05 | Usabilidad | Postulante | Pierde conexión. | Datos guardados se conservan. | 0 pérdida de secciones guardadas. |
| QA-06 | Mantenibilidad | Equipo | Cambia proveedor SMTP. | Solo cambia adaptador. | 0 cambios en dominio. |
| QA-07 | Interoperabilidad | Registro Civil | Timeout 10 s. | Corte a 3 s; reintento. | UI no se bloquea. |
| QA-08 | Legalidad | Postulante | No acepta consentimiento. | No puede enviar. | 100 % bloqueado. |
| QA-09 | Escalabilidad | Operaciones | Aumenta a 2 réplicas. | API stateless funciona. | Sin sesiones locales. |
| QA-10 | Observabilidad | Operaciones | Error 500. | Log con `trace_id`. | Trazable en < 5 min. |

---

## 11.2 Rendimiento y capacidad

### Objetivos

| Operación | p95 | p99 |
|---|---|---|
| Catálogos cacheados | < 100 ms | < 200 ms |
| Catálogos desde DB | < 300 ms | < 500 ms |
| Guardar sección | < 300 ms | < 600 ms |
| Login | < 400 ms | < 800 ms |
| Enviar postulación (sin correo) | < 5 s | < 8 s |
| Descargar PDF | < 2 s | < 4 s |

### Capacidad

- 200 usuarios concurrentes.
- 50 req/s sostenidos.
- 1 000 postulaciones por convocatoria.
- 5 000 documentos.
- Crecimiento horizontal de la API.

### Cuellos de botella

| Punto | Mitigación |
|---|---|
| PostgreSQL | Índices, pool, consultas parametrizadas. |
| Redis | TTL, DB separadas, monitoreo. |
| Generación de PDF | Worker o límite de concurrencia. |
| SMTP | Outbox y reintentos. |
| Storage | S3/volumen con IOPS adecuados. |

---

## 11.3 Caché

### Estrategia

- **Cache-aside** para catálogos.
- TTL 24 h + jitter.
- Versionado global.
- No cachear datos personales.
- No cachear respuestas autenticadas.

### Métricas

- Hit ratio > 90 % para catálogos.
- Latencia de Redis < 5 ms.
- Errores de Redis < 0.1 %.

### Invalidación

- Cambio de catálogo → `INCR cat:version`.
- Las claves antiguas expiran.
- No `KEYS *`.

---

## 11.4 Optimización de base de datos

- Índices en claves foráneas y columnas de búsqueda.
- `EXPLAIN ANALYZE` para consultas lentas.
- Paginación en listados.
- Evitar `SELECT *` en producción.
- Pool de conexiones dimensionado.
- Transacciones cortas.
- `FOR UPDATE` solo en envío.
- Partición de auditoría por año si crece.
- VACUUM/AUTOVACUUM monitoreado.

---

## 11.5 Rendimiento del frontend

- Code splitting por ruta.
- Lazy loading de secciones pesadas.
- `staleTime` de 24 h para catálogos.
- Evitar re-renderizados innecesarios.
- Imágenes optimizadas.
- Bundle inicial < 300 KB gzip.
- Lighthouse performance > 85 en desktop.
- Carga inicial < 2.5 s en 4G.

---

## 11.6 Disponibilidad y continuidad

| Aspecto | Objetivo |
|---|---|
| Disponibilidad | ≥ 99.5 % mensual. |
| RPO | ≤ 15 min. |
| RTO | ≤ 2 h. |
| Backups | Diario completo + WAL. |
| Restauración | Probada trimestralmente. |
| Degradación | Catálogos desde DB; auth fail-closed. |

### Arquitectura de despliegue

- 2+ réplicas de API.
- PostgreSQL con réplica de lectura opcional.
- Redis con persistencia AOF.
- Storage replicado.
- Balanceador con health checks.

---

## 11.7 Observabilidad

### Logs

```json
{
  "level": "info",
  "ts": "2026-10-31T16:42:00-05:00",
  "trace_id": "8f3c2a1e-...",
  "method": "POST",
  "path": "/api/v1/postulaciones/.../enviar",
  "status": 200,
  "latency_ms": 842,
  "user_id": "550e...",
  "msg": "postulacion_enviada"
}
```

- No incluir contraseñas, tokens, códigos dactilares ni PDFs.
- Redactar cédula parcialmente en logs de aplicación.

### Métricas

- `http_requests_total`
- `http_request_duration_seconds`
- `http_errors_total`
- `cache_hits_total` / `cache_misses_total`
- `auth_lockouts_total`
- `postulaciones_enviadas_total`
- `notificaciones_fallidas_total`
- `db_pool_in_use`

### Trazas

- `trace_id` en cada request.
- Propagación a logs y respuestas de error.
- OpenTelemetry opcional.

### Alertas

| Alerta | Umbral |
|---|---|
| Error rate | > 5 % en 5 min. |
| Latencia p95 | > 1 s en 5 min. |
| Redis caído | > 1 min. |
| DB pool agotado | > 80 % en 5 min. |
| Correos fallidos | > 10 en 15 min. |
| Disco | > 80 %. |

---

## 11.8 Escalabilidad

- API stateless: no guarda sesión en memoria.
- Redis y PostgreSQL externos.
- Storage abstraído.
- Configuración por variables de entorno.
- Escalado horizontal de réplicas.
- Separación futura de workers de PDF/correo.
- Posibilidad de extraer módulos a servicios si el volumen lo exige.

---

## 11.9 SLOs

| SLO | Objetivo | Medición |
|---|---|---|
| Disponibilidad | 99.5 % | Uptime mensual. |
| Latencia de catálogos | p95 < 100 ms | Métricas. |
| Latencia de guardado | p95 < 300 ms | Métricas. |
| Envío exitoso | > 99 % | Logs/outbox. |
| Correo entregado | > 98 % | SMTP/outbox. |
| Errores 5xx | < 1 % | Métricas. |

---

## 11.10 Trade-offs

| Decisión | Beneficio | Costo |
|---|---|---|
| Microservicio modular | Simplicidad operativa. | Menos aislamiento que microservicios. |
| Redis para bloqueo | Rápido y con TTL. | Dependencia de Redis; fail-closed. |
| JWT + refresh cookie | Escalable y seguro. | Complejidad de refresh/rotación. |
| PDF en servidor | Formato consistente. | Uso de CPU. |
| Outbox | No se pierden correos. | Worker y tabla adicional. |
| Cache-aside | Baja latencia. | Invalidación y consistencia eventual. |
| Transacciones locales | Atomicidad simple. | No cubre servicios externos. |

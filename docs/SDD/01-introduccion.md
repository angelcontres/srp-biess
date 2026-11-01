# 01. Introducción — SDD

## 1.1 Propósito

Este documento especifica el diseño del **Módulo de Registro y Postulación** del SRP-BIESS. Su propósito es transformar los requisitos del SRS en una arquitectura implementable, verificable y defendible técnicamente.

El SDD responde a:

- ¿Cómo se organiza el sistema?
- ¿Cómo se separan las responsabilidades?
- ¿Cómo se persisten y transaccionan los datos?
- ¿Cómo se protege la información?
- ¿Cómo se despliega y se prueba?

---

## 1.2 Alcance

| Incluido | Excluido |
|---|---|
| EF-11 y EF-12. | EF-13 y EF-14. |
| Arquitectura backend en capas con DI. | Microservicios distribuidos. |
| PostgreSQL, Redis, archivos y correo. | Infraestructura institucional real. |
| SPA React + Vite + Tailwind. | Aplicación móvil nativa. |
| Adaptadores simulados de Registro Civil e IESS. | Convenios e integraciones productivas. |
| Seguridad, auditoría y pruebas. | Certificación formal y auditoría externa. |

---

## 1.3 Audiencia

| Lector | Uso |
|---|---|
| Equipo de desarrollo | Guía de implementación. |
| Docente evaluador | Verificación de arquitectura y rúbrica. |
| Arquitecto / líder técnico | Revisión de decisiones y riesgos. |
| QA | Diseño de pruebas y criterios de aceptación. |
| Operaciones | Despliegue, monitoreo y runbook. |

---

## 1.4 Definiciones y referencias

Las definiciones de negocio están en `SRS/01-introduccion.md`. Para este documento se agregan:

| Término | Definición |
|---|---|
| **Handler** | Adaptador HTTP que traduce requests a casos de uso. |
| **Service** | Caso de uso que orquesta reglas de negocio. |
| **Repository** | Puerto de persistencia. |
| **Puerto** | Interfaz que define una dependencia del dominio. |
| **Adaptador** | Implementación concreta de un puerto (PostgreSQL, Redis, SMTP). |
| **TxManager** | Componente que ejecuta una función dentro de una transacción. |
| **DTO** | Objeto de transferencia entre capas o hacia la API. |
| **Outbox** | Tabla de eventos pendientes de envío. |
| **Cache-aside** | Patrón de caché consultada antes de la base. |
| **ADR** | Registro de decisión de arquitectura. |

**Referencias:** IEEE 1016-2009, ISO/IEC 25010, OWASP ASVS, RFC 7807, SRS del proyecto.

---

## 1.5 Objetivos de diseño

| ID | Objetivo | Criterio de éxito |
|---|---|---|
| OD-01 | Separar presentación, aplicación, dominio e infraestructura. | Ningún handler contiene SQL ni reglas de negocio. |
| OD-02 | Invertir dependencias mediante interfaces. | Los servicios dependen de puertos, no de implementaciones. |
| OD-03 | Garantizar atomicidad. | El envío de postulación es todo o nada. |
| OD-04 | Optimizar consultas frecuentes. | Catálogos con caché y TTL. |
| OD-05 | Proteger identidad y datos. | Argon2id, JWT corto, refresh seguro, auditoría. |
| OD-06 | Permitir pruebas sin infraestructura. | Servicios testeables con mocks. |
| OD-07 | Facilitar el despliegue local. | `docker compose up` funcional. |
| OD-08 | Ser demostrable en 10 minutos. | Flujo completo registro → envío. |

---

## 1.6 Principios de diseño

### 1.6.1 Separación de responsabilidades

Cada capa tiene una responsabilidad única:

- **Presentación:** HTTP, validación de contrato, serialización.
- **Aplicación:** casos de uso, transacciones, autorización.
- **Dominio:** entidades, reglas, cálculos, estados.
- **Infraestructura:** PostgreSQL, Redis, SMTP, archivos, integraciones.

### 1.6.2 Inversión de dependencias

El dominio define interfaces; la infraestructura las implementa. El flujo de dependencias apunta hacia el dominio:

```text
HTTP → Application → Domain ← Infrastructure
```

### 1.6.3 Transacciones explícitas

Las operaciones multi-tabla se ejecutan dentro de `TxManager.WithTx`. Si una inserción falla, se hace `ROLLBACK` y no quedan datos huérfanos.

### 1.6.4 Seguridad desde el diseño

- Validar identidad antes de crear la cuenta.
- Hash de contraseña con Argon2id.
- Tokens de acceso cortos y refresh rotativo.
- Validación de pertenencia en cada recurso.
- Archivos fuera del directorio público.
- Logs sin secretos.

### 1.6.5 API-first

La SPA consume contratos REST versionados. Los DTO se documentan y se validan en el servidor. Los errores usan un formato uniforme.

### 1.6.6 Cache-aside

Los catálogos se leen de Redis; si no están, se consultan en PostgreSQL y se cachean con TTL. La base sigue siendo la fuente de verdad.

### 1.6.7 Observabilidad

Cada request tiene `trace_id`; los logs son estructurados; los endpoints de salud permiten operar el servicio.

### 1.6.8 Simplicidad deliberada

Para el primer parcial se prefiere un **microservicio modular** antes que una constelación de servicios. La complejidad distribuida no aporta valor al alcance y sí aumenta el riesgo de fallo.

---

## 1.7 Restricciones

| Código | Restricción | Impacto |
|---|---|---|
| RES-01 | Go + Fiber/Gin. | Stack del backend. |
| RES-02 | React + Vite + Tailwind. | Stack del frontend. |
| RES-03 | PostgreSQL. | Modelo relacional y transacciones. |
| RES-04 | Redis. | Bloqueo y caché. |
| RES-05 | PDF ≤ 4 MB / 2 MB; total 50 MB. | Validación de documentos. |
| RES-06 | 3 intentos / 10 minutos. | Diseño de claves y TTL. |
| RES-07 | Nacionalidad ecuatoriana. | Regla bloqueante. |
| RES-08 | LOPDP. | Consentimiento y protección de datos. |

---

## 1.8 Relación con el SRS

| SRS | SDD |
|---|---|
| RF-EF11 | `03`, `05`, `07`, `09`, `10` |
| RF-EF12 | `04`, `05`, `06`, `07`, `08`, `10` |
| RNF | `09`, `11`, `13`, `15` |
| Reglas de negocio | `05`, `09`, `10` |
| Criterios de aceptación | `15` |

---

## 1.9 Visión general

El capítulo 2 describe el contexto y los impulsores. El capítulo 3 presenta la arquitectura. El capítulo 4 detalla los datos. Los capítulos 5 y 6 describen componentes backend y frontend. El capítulo 7 define la API. El capítulo 8 la interfaz. El capítulo 9 la seguridad. El capítulo 10 los procesos y transacciones. El capítulo 11 los atributos de calidad. El capítulo 12 la trazabilidad. El capítulo 13 el despliegue. El capítulo 14 las decisiones. El capítulo 15 las pruebas. El capítulo 16 los anexos.

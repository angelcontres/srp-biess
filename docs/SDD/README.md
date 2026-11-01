# SDD — Diseño de Software

**Proyecto:** Sistema de Registro y Postulación al Directorio del BIESS (SRP-BIESS)  
**Módulo diseñado:** Registro, autenticación y formulario de postulación (EF-11 y EF-12)  
**Versión:** 1.0  
**Fecha:** 31/10/2026  
**Estándar de referencia:** IEEE 1016-2009

---

## 1. Propósito

Este documento describe **cómo se construye** el módulo inicial del SRP-BIESS: arquitectura, datos, componentes, interfaces, seguridad, procesos, despliegue y pruebas. Es la respuesta técnica al SRS y la base para la implementación, la sustentación y la evaluación de la rúbrica.

El SDD no repite los requisitos; los referencia. Cada decisión de diseño indica qué requisito satisface y qué alternativa se descartó.

---

## 2. Resumen de la solución

| Capa | Tecnología | Responsabilidad |
|---|---|---|
| Presentación | React + Vite + TypeScript + Tailwind CSS | SPA responsiva, formularios, estado, rutas protegidas. |
| API | Go 1.23+ + Fiber v2 | Handlers, middleware, validación de contrato. |
| Aplicación | Servicios de caso de uso | Reglas de negocio, orquestación, transacciones. |
| Dominio | Entidades y puertos | Modelo independiente de infraestructura. |
| Persistencia | PostgreSQL 16 + pgx | Datos relacionales, transacciones, integridad. |
| Caché y seguridad | Redis 7 | Catálogos, bloqueo, rate limiting, tokens temporales. |
| Archivos | Volumen local / S3 | PDF de respaldo y formularios generados. |
| Correo | SMTP / Mailpit | Recuperación y envío de postulación. |
| Integraciones | Adaptadores REST / mocks | Registro Civil e Historial Laboral IESS. |

### Decisión central

Se implementa un **microservicio transaccional único** con arquitectura interna en capas y puertos/adaptadores. Esto cumple el desacoplamiento exigido, permite inyección de dependencias explícita y deja preparada la extracción futura de módulos (administración, evaluación) sin sobredimensionar el primer parcial.

---

## 3. Estructura del SDD

| Capítulo | Contenido |
|---|---|
| [00. Control de versiones](./00-control-de-versiones.md) | Metadatos e historial. |
| [01. Introducción](./01-introduccion.md) | Propósito, alcance, referencias y principios. |
| [02. Descripción general](./02-descripcion-general.md) | Contexto, actores, impulsores y restricciones. |
| [03. Arquitectura](./03-arquitectura-de-software.md) | Estilo, capas, C4, stack y decisiones. |
| [04. Diseño de datos](./04-diseno-de-datos.md) | Modelo, tablas, Redis, archivos y transacciones. |
| [05. Componentes backend](./05-diseno-componentes-backend.md) | Paquetes, interfaces, DI y servicios. |
| [06. Componentes frontend](./06-diseno-componentes-frontend.md) | Rutas, estado, formularios y componentes. |
| [07. API REST](./07-diseno-api-rest.md) | Contratos, endpoints, errores y seguridad. |
| [08. Interfaz de usuario](./08-diseno-interfaz-usuario.md) | Pantallas, navegación, diseño visual y accesibilidad. |
| [09. Seguridad](./09-diseno-seguridad.md) | Amenazas, autenticación, autorización, uploads y LOPDP. |
| [10. Procesos y transacciones](./10-diseno-procesos-y-transacciones.md) | Secuencias, estados, rollback e idempotencia. |
| [11. Atributos de calidad](./11-atributos-de-calidad.md) | Rendimiento, disponibilidad, observabilidad y escalabilidad. |
| [12. Trazabilidad](./12-trazabilidad.md) | RF/RNF → componentes → pruebas. |
| [13. Despliegue y operación](./13-despliegue-y-operacion.md) | Docker, entornos, CI/CD y runbook. |
| [14. Decisiones de arquitectura](./14-decisiones-de-arquitectura.md) | ADR-001 a ADR-014. |
| [15. Plan de pruebas](./15-plan-de-pruebas.md) | Estrategia, casos, cobertura y calidad. |
| [16. Anexos](./16-anexos.md) | DDL, variables, ejemplos y glosario. |

---

## 4. Decisiones clave

| ID | Decisión | Justificación |
|---|---|---|
| ADR-001 | Microservicio modular único, no microservicios distribuidos. | Menor complejidad operativa para el parcial; capas desacopladas. |
| ADR-002 | Fiber v2 como framework HTTP. | Alto rendimiento, middleware simple, buena integración con Go. |
| ADR-003 | Inyección de dependencias manual en `main.go`. | Explícita, didáctica y verificable; sin magia de framework. |
| ADR-004 | pgx + `TxManager` para transacciones. | Control claro de `BEGIN/COMMIT/ROLLBACK` y rendimiento. |
| ADR-005 | Redis cache-aside con TTL y versionado. | Cumple la rúbrica y reduce latencia de catálogos. |
| ADR-006 | JWT corto + refresh opaco en cookie `HttpOnly`. | Reduce exposición de tokens en el navegador. |
| ADR-007 | TanStack Query + Zustand. | Separa estado del servidor del estado de UI. |
| ADR-008 | React Hook Form + Zod. | Formularios performantes y validación declarativa. |
| ADR-009 | Almacenamiento abstracto local/S3. | Portabilidad y despliegue. |
| ADR-010 | Adaptadores simulados para Registro Civil e IESS. | Permite demo sin depender de convenios. |
| ADR-011 | Outbox de notificaciones. | Evita perder correos y desacopla el envío. |
| ADR-012 | PDF generado en servidor con plantilla. | Formulario imprimible y firmable. |
| ADR-013 | Validación algorítmica de cédula (Módulo 10). | Fail-fast offline antes de invocar servicios externos. |
| ADR-014 | Pagos con tarjeta y facturación con Mocks implementables. | Arquitectura hexagonal para pagos y emisión de factura por correo. |

---

## 5. Cómo usar este SDD

1. Leer `03-arquitectura-de-software.md` para entender la estructura.
2. Revisar `04-diseno-de-datos.md` y `16-anexos.md` antes de implementar.
3. Usar `05` y `06` como guía de paquetes y componentes.
4. Implementar los contratos de `07-diseno-api-rest.md`.
5. Verificar seguridad con `09` y procesos con `10`.
6. Ejecutar las pruebas de `15` y demostrar la trazabilidad de `12`.

---

## 6. Estado

Documento en **línea base inicial**. Las decisiones marcadas como `[PENDIENTE]` requieren validación con la SDNGTH o el docente. El diseño asume que los servicios externos pueden simularse para la demostración académica.

# SRS — Especificación de Requisitos de Software

**Proyecto:** Sistema de Registro y Postulación al Directorio del BIESS (SRP-BIESS)  
**Módulo documentado:** Registro, autenticación y formulario de postulación (EF-11 y EF-12)  
**Versión:** 1.0  
**Fecha:** 31/10/2026  
**Estándar de referencia:** ISO/IEC/IEEE 29148:2018

---

## 1. Propósito del SRS

Este documento especifica, de forma completa, verificable y trazable, **qué debe hacer** el módulo inicial del Sistema de Registro y Postulación al Directorio del BIESS, así como las restricciones de calidad, datos, interfaces y negocio que condicionan su implementación. Es la línea base para el diseño (SDD), la planificación de pruebas y la aceptación del producto.

El SRS no prescribe la tecnología de implementación; esas decisiones se documentan en el SDD. Sin embargo, cuando el requerimiento funcional institucional o la rúbrica del examen imponen una tecnología o restricción, esta se registra como restricción del producto.

---

## 2. Alcance resumido

| Incluido | Excluido de esta fase |
|---|---|
| EF-11: registro, validación de identidad, autenticación, bloqueo y recuperación. | EF-13: módulo de administración. |
| EF-12: formulario de postulación, secciones, documentos, elegibilidad, guardado y envío. | EF-14: banco de preguntas y evaluación teórica. |
| Catálogos geográficos y caché. | Calificación de méritos y oposición. |
| Auditoría, notificaciones y generación de PDF. | Reconsideraciones, listas de elegibles y reportes finales. |

---

## 3. Estructura del documento

| Capítulo | Contenido |
|---|---|
| [00. Control de versiones](./00-control-de-versiones.md) | Metadatos, historial y aprobaciones. |
| [01. Introducción](./01-introduccion.md) | Propósito, alcance, contexto, definiciones y referencias. |
| [02. Descripción general](./02-descripcion-general.md) | Perspectiva del producto, actores, entorno y restricciones. |
| [03. Requisitos funcionales](./03-requisitos-funcionales.md) | Catálogo y especificación detallada de RF. |
| [04. Requisitos no funcionales](./04-requisitos-no-funcionales.md) | Calidad, rendimiento, seguridad, usabilidad y operación. |
| [05. Requisitos de datos](./05-requisitos-de-datos.md) | Entidades, atributos, calidad, retención y auditoría. |
| [06. Interfaces externas](./06-interfaces-externas.md) | Interfaces de usuario, software, comunicaciones y hardware. |
| [07. Reglas de negocio](./07-reglas-de-negocio.md) | Reglas, tablas de decisión y mensajes. |
| [08. Trazabilidad y verificación](./08-trazabilidad-y-verificacion.md) | Matrices RF/RN/RNF y métodos de verificación. |
| [09. Criterios de aceptación](./09-criterios-de-aceptacion.md) | Escenarios Gherkin de aceptación. |
| [10. Anexos](./10-anexos.md) | Catálogo de mensajes, ejemplos y pendientes. |

---

## 4. Cómo leer los requisitos

Cada requisito funcional tiene un identificador único:

- `RF-EF11-xx`: requisitos del módulo de registro y autenticación.
- `RF-EF12-xx`: requisitos del formulario de postulación.
- `RF-CAT-xx`: requisitos de catálogos geográficos.
- `RNF-xxx-xx`: requisitos no funcionales por categoría.
- `RN-xx`: reglas de negocio que restringen uno o varios RF.

La prioridad se expresa con MoSCoW:

| Prioridad | Significado |
|---|---|
| **Must** | Obligatorio para aceptar la fase. |
| **Should** | Importante; puede postergarse con justificación. |
| **Could** | Deseable; no compromete la aceptación. |
| **Won't** | Explícitamente fuera de esta fase. |

---

## 5. Relación con el SDD

| SRS | SDD relacionado |
|---|---|
| Requisitos funcionales | Arquitectura, componentes, API, procesos. |
| Requisitos no funcionales | Atributos de calidad, seguridad, despliegue. |
| Requisitos de datos | Diseño de datos y DDL. |
| Reglas de negocio | Servicios de dominio, validaciones y transacciones. |
| Criterios de aceptación | Plan de pruebas y trazabilidad. |

---

## 6. Estado

Documento en **línea base inicial**. Las decisiones marcadas como `[PENDIENTE]` requieren validación con la SDNGTH o con el docente antes de considerarse cerradas.

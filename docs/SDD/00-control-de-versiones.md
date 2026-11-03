# 00. Control de versiones — SDD

| Campo | Valor |
|---|---|
| **Documento** | Diseño de Software (SDD) |
| **Proyecto** | Sistema de Registro y Postulación al Directorio del BIESS |
| **Sigla** | SRP-BIESS |
| **Versión** | 1.1 |
| **Fecha** | 31/10/2026 |
| **Estado** | Versión ampliada y verificada |
| **Autor(es)** | Paulo Rafael Orrala Arriaga, Carlos Fernando Patiño Garcia, Jose Daniel Gonzalez Cochea, Angel Alejandro Villon Panimboza |
| **Revisor(es)** | `[Revisor / docente]` |
| **Aprobador(es)** | `[Aprobador]` |
| **Documento fuente** | `Docs/SRS/` versión 1.1 |

---

## 1. Historial de versiones

| Versión | Fecha | Autor | Descripción |
|---|---|---|---|
| 0.1 | 25/10/2026 | Equipo | Esquema inicial de arquitectura y modelo de datos. |
| 0.2 | 28/10/2026 | Equipo | Componentes backend/frontend, API y procesos. |
| 0.9 | 30/10/2026 | Equipo | Revisión de seguridad, transacciones y trazabilidad. |
| 1.0 | 31/10/2026 | Equipo | Línea base para el primer parcial. |
| 1.1 | 31/10/2026 | Equipo | Incorporación de ADR-013 (Validación Cédula Módulo 10) y ADR-014 (Pago con Tarjeta y Facturación con Mocks implementables), endpoints y DDL asociado. |
| 1.2 | 03/11/2026 | Equipo | Se retira el ADR-014 con sus endpoints, tablas y flujos: la postulación es gratuita y no hay pagos ni facturación. |

---

## 2. Control de cambios

1. Registrar el cambio y su motivo.
2. Identificar los requisitos del SRS afectados.
3. Actualizar los diagramas y la matriz de trazabilidad.
4. Revisar el impacto en pruebas y despliegue.
5. Aprobar la nueva versión.

---

## 3. Documentos relacionados

| Documento | Ubicación | Relación |
|---|---|---|
| SRS | `Docs/SRS/` | Requisitos que este diseño satisface. |
| Requerimiento Funcional v1.4 | `requerimiento.pdf` | Fuente institucional. |
| Examen Práctico | `Examen Práctico del Primer Parcial.md` | Rúbrica. |
| Repositorio | [github.com/angelcontres/srp-biess](https://github.com/angelcontres/srp-biess) | Código y pruebas. |
| DDL | `16-anexos.md` | Script de referencia. |

---

## 4. Aprobaciones

| Rol | Nombre | Firma | Fecha |
|---|---|---|---|
| Autor | Paulo Rafael Orrala Arriaga | | |
| Autor | Carlos Fernando Patiño Garcia | | |
| Autor | Jose Daniel Gonzalez Cochea | | |
| Autor | Angel Alejandro Villon Panimboza | | |
| Revisor | `[Revisor / Docente]` | | |
| Aprobador | `[Aprobador]` | | |

> Los nombres y el repositorio son marcadores que el equipo debe completar antes de la entrega.

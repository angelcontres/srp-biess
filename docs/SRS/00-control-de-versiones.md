# 00. Control de versiones — SRS

| Campo | Valor |
|---|---|
| **Documento** | Especificación de Requisitos de Software (SRS) |
| **Proyecto** | Sistema de Registro y Postulación al Directorio del BIESS |
| **Sigla** | SRP-BIESS |
| **Versión** | 1.1 |
| **Fecha** | 31/10/2026 |
| **Estado** | Versión ampliada y verificada |
| **Autor(es)** | Paulo Rafael Orrala Arriaga, Carlos Fernando Patiño Garcia, Jose Daniel Gonzalez Cochea, Angel Alejandro Villon Panimboza |
| **Revisor(es)** | `[Revisor / docente]` |
| **Aprobador(es)** | `[Aprobador]` |
| **Clasificación** | Documento académico — uso interno del equipo |

---

## 1. Historial de versiones

| Versión | Fecha | Autor | Descripción del cambio |
|---|---|---|---|
| 0.1 | 24/10/2026 | Equipo | Esquema inicial y extracción de requisitos del RF v1.4. |
| 0.2 | 28/10/2026 | Equipo | Requisitos funcionales EF-11 y EF-12, reglas de negocio y criterios de aceptación. |
| 0.9 | 30/10/2026 | Equipo | Revisión de consistencia con el SDD y la rúbrica. |
| 1.0 | 31/10/2026 | Equipo | Línea base para la entrega del primer parcial. |
| 1.1 | 31/10/2026 | Equipo | Incorporación de requisitos de validación algorítmica de cédula (RF-VAL-01), procesamiento de pago con tarjeta (RF-PAG-01) y facturación electrónica (RF-PAG-02). |
| 1.2 | 03/11/2026 | Equipo | Se retiran RF-PAG-01, RF-PAG-02, RN-42 y RN-43: la postulación es gratuita y no hay pagos ni facturación. |

---

## 2. Control de cambios

Todo cambio posterior a la versión 1.0 debe:

1. Registrar el motivo en la tabla anterior.
2. Identificar los requisitos afectados (`RF`, `RNF`, `RN`).
3. Actualizar la matriz de trazabilidad (`08-trazabilidad-y-verificacion.md`).
4. Revisar el impacto en el SDD y en los casos de prueba.
5. Ser aprobado por el responsable del documento.

---

## 3. Documentos relacionados

| Documento | Ubicación | Relación |
|---|---|---|
| Requerimiento Funcional v1.4 | `requerimiento.pdf` | Fuente institucional de EF-11, EF-12 y restricciones. |
| Examen Práctico del Primer Parcial | `Examen Práctico del Primer Parcial.md` | Rúbrica y alcance técnico de la entrega. |
| SRS | `Docs/SRS/` | Este documento. |
| SDD | `Docs/SDD/` | Diseño derivado de este SRS. |
| Repositorio de código | [github.com/angelcontres/srp-biess](https://github.com/angelcontres/srp-biess) | Implementación y pruebas. |

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

> **Nota académica:** las firmas se completan al momento de la entrega. Los nombres y el repositorio son marcadores que el equipo debe reemplazar.

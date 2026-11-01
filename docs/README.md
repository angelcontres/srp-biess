# Documentación del Sistema de Registro y Postulación al Directorio del BIESS

| Campo | Valor |
|---|---|
| **Proyecto** | Sistema de Registro y Postulación al Directorio del BIESS |
| **Sigla** | SRP-BIESS |
| **Asignatura** | Frameworks de Programación Web |
| **Entrega** | Primer parcial — Módulo de Registro y Postulación (EF-11 y EF-12) |
| **Versión documental** | 1.1 |
| **Fecha** | 31/10/2026 |
| **Estado** | Versión ampliada y verificada |
| **Repositorio** | [github.com/angelcontres/srp-biess](https://github.com/angelcontres/srp-biess) |
| **Equipo** | Paulo Rafael Orrala Arriaga, Carlos Fernando Patiño Garcia, Jose Daniel Gonzalez Cochea, Angel Alejandro Villon Panimboza |

> **Propósito.** Reunir, en documentos Markdown separados y trazables, el **SRS** (requisitos) y el **SDD** (diseño) del módulo inicial del concurso público de méritos y oposición para la designación de los Miembros Principales y Alternos del Directorio del BIESS, en representación de los Afiliados, Jubilados y Empleadores.

---

## 1. Alcance de esta entrega

| Módulo | Código | ¿Incluido? | Descripción |
|---|---|---|---|
| Registro y autenticación | EF-11 | ✅ Sí | Validación de identidad, registro, inicio de sesión, bloqueo temporal y recuperación de contraseña. |
| Formulario de postulación | EF-12 | ✅ Sí | Formulario por secciones, catálogos geográficos, documentos PDF, elegibilidad, guardado de progreso y envío. |
| Validación de identidad y cédula | Soporte / ADR-013 | ✅ Sí | Validación algorítmica Módulo 10 ecuatoriano e interoperabilidad con Registro Civil. |
| Pagos y facturación electrónica | Soporte / ADR-014 | ✅ Sí | Validación de tarjeta (Luhn), pasarela mockeable y emisión/envío de factura con clave de acceso SRI. |
| Catálogos geográficos | Soporte | ✅ Sí | Países, provincias y cantones, con caché en Redis. |
| Gestión documental | Soporte | ✅ Sí | Carga, validación, almacenamiento y descarga de PDF. |
| Notificaciones | Soporte | ✅ Sí | Recuperación de contraseña, factura electrónica y envío de la postulación en PDF. |
| Auditoría | Soporte | ✅ Sí | Eventos críticos y cambios sobre la postulación. |
| Módulo de administración | EF-13 | ❌ Futuro | Se documenta solo como interfaz futura. |
| Banco de preguntas y evaluación | EF-14 | ❌ Futuro | Fuera del primer parcial. |

---

## 2. Estructura documental

```text
Docs/
├── README.md                 ← este índice general
├── PLANNING.md               ← Plan de ejecución, WBS, RACI y guion de sustentación
├── TICKETS.md                ← Backlog de tickets de implementación detallados
├── SRS/                      ← Especificación de Requisitos de Software
│   ├── README.md
│   ├── 00-control-de-versiones.md
│   ├── 01-introduccion.md
│   ├── 02-descripcion-general.md
│   ├── 03-requisitos-funcionales.md
│   ├── 04-requisitos-no-funcionales.md
│   ├── 05-requisitos-de-datos.md
│   ├── 06-interfaces-externas.md
│   ├── 07-reglas-de-negocio.md
│   ├── 08-trazabilidad-y-verificacion.md
│   ├── 09-criterios-de-aceptacion.md
│   └── 10-anexos.md
└── SDD/                      ← Diseño de Software
    ├── README.md
    ├── 00-control-de-versiones.md
    ├── 01-introduccion.md
    ├── 02-descripcion-general.md
    ├── 03-arquitectura-de-software.md
    ├── 04-diseno-de-datos.md
    ├── 05-diseno-componentes-backend.md
    ├── 06-diseno-componentes-frontend.md
    ├── 07-diseno-api-rest.md
    ├── 08-diseno-interfaz-usuario.md
    ├── 09-diseno-seguridad.md
    ├── 10-diseno-procesos-y-transacciones.md
    ├── 11-atributos-de-calidad.md
    ├── 12-trazabilidad.md
    ├── 13-despliegue-y-operacion.md
    ├── 14-decisiones-de-arquitectura.md
    ├── 15-plan-de-pruebas.md
    └── 16-anexos.md
```

---

## 3. Ruta de lectura recomendada

1. **SRS — Introducción y descripción general:** entiende el problema, alcance y actores.
2. **SRS — Requisitos funcionales y reglas de negocio:** qué debe hacer el sistema y con qué restricciones.
3. **SRS — Requisitos no funcionales y criterios de aceptación:** cómo se verifica la calidad.
4. **SDD — Arquitectura y datos:** cómo se organiza la solución y cómo se persiste.
5. **SDD — Componentes, API e interfaz:** cómo se implementa cada capa.
6. **SDD — Seguridad, procesos, calidad, despliegue y pruebas:** cómo se opera y se defiende.

---

## 4. Convenciones

- **RF-EF11-xx / RF-EF12-xx:** requisitos funcionales.
- **RNF-xx:** requisitos no funcionales.
- **RN-xx:** reglas de negocio.
- **CU-xx:** casos de uso.
- **ADR-xxx:** decisión de arquitectura.
- **Prioridad MoSCoW:** `Must`, `Should`, `Could`, `Won't` (para esta fase).
- **Fechas:** `DD/MM/AAAA` en interfaz; ISO 8601 en API y base de datos.
- **Zona horaria:** `America/Guayaquil` (UTC−05:00).
- **Idioma:** español (Ecuador).

---

## 5. Cumplimiento de la rúbrica

| Criterio de la rúbrica | Dónde se evidencia |
|---|---|
| 1. Arquitectura backend e inyección de dependencias | `SDD/03`, `SDD/05`, `SDD/14` |
| 2. Persistencia relacional y transacciones | `SDD/04`, `SDD/10`, `SDD/16` |
| 3. Redis: bloqueo, TTL y caché | `SDD/04`, `SDD/09`, `SDD/10` |
| 4. Frontend SPA reactiva y modular | `SDD/06`, `SDD/08` |
| 5. Consumo de API y rutas protegidas | `SDD/06`, `SDD/07` |
| 6. Reglas de negocio EF-11 y EF-12 | `SRS/03`, `SRS/07`, `SRS/09` |

---

## 6. Exportación a PDF

Cada archivo puede exportarse con **VS Code + Markdown PDF** o con Pandoc:

```bash
pandoc Docs/SRS/01-introduccion.md -o introduccion.pdf --pdf-engine=xelatex
```

Para generar un documento consolidado, concatene los capítulos en orden numérico y exporte el resultado.

---

## 7. Estado y mantenimiento

- **Versión 1.0:** línea base del primer parcial.
- Los nombres del equipo, repositorio y datos institucionales están como marcadores `[...]`.
- Toda decisión pendiente se registra en `SDD/16-anexos.md` y `SRS/10-anexos.md`.
- Los cambios deben reflejarse primero en el SRS y luego en el SDD, manteniendo la trazabilidad.

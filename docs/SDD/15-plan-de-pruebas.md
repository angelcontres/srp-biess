# 15. Plan de pruebas — SDD

## 15.1 Estrategia

Se aplica una pirámide de pruebas con énfasis en la lógica de negocio y los flujos críticos:

```text
        ┌──────────────┐
        │   E2E (10%)  │  Playwright
        ├──────────────┤
        │ Integración  │  Testcontainers / httptest
        │    (30%)     │
        ├──────────────┤
        │  Unitarias   │  go test / Vitest
        │    (60%)     │
        └──────────────┘
```

### Objetivos

- Detectar regresiones en reglas de negocio.
- Verificar transacciones y rollback.
- Verificar bloqueo y caché.
- Validar los flujos completos de registro y envío.
- Medir rendimiento y seguridad.

---

## 15.2 Herramientas

| Nivel | Backend | Frontend |
|---|---|---|
| Unitarias | `testing`, `testify` | Vitest |
| Integración | `httptest`, Testcontainers | React Testing Library |
| E2E | Playwright | Playwright |
| Carga | k6 | Lighthouse |
| Seguridad | OWASP ZAP, `govulncheck` | `npm audit` |
| Accesibilidad | — | axe |
| Cobertura | `go test -cover` | `vitest --coverage` |

---

## 15.3 Pruebas unitarias

### Backend

| ID | Caso | Requisito |
|---|---|---|
| UT-01 | Cálculo de experiencia: bisiesto, fin de mes, mismo día. | RF-EF12-13 |
| UT-02 | Validación de cédula ecuatoriana. | RF-EF11-02 |
| UT-03 | Política de contraseña. | RN-05 |
| UT-04 | Generación de código único. | RF-EF12-20 |
| UT-05 | Evaluación de inhabilidades bloqueantes/alerta. | RF-EF12-16 |
| UT-06 | Validación de nacionalidad. | RF-EF12-04 |
| UT-07 | Regla de estado de postulación. | RF-EF12-22 |
| UT-08 | Cálculo de total de documentos. | RN-28 |
| UT-09 | Parseo de fechas `DD/MM/AAAA` e ISO. | RN-34 |
| UT-10 | Mapeo de errores de dominio a HTTP. | RF-EF12-24 |

### Frontend

| ID | Caso | Requisito |
|---|---|---|
| UT-F01 | Esquema Zod de teléfono. | RF-EF12-06 |
| UT-F02 | Esquema de contraseña. | RN-05 |
| UT-F03 | Formateo de fecha. | RN-34 |
| UT-F04 | Cálculo de tiempo en UI. | RF-EF12-13 |
| UT-F05 | Normalización de errores RFC 7807. | RF-EF12-24 |

---

## 15.4 Pruebas de integración

| ID | Caso | Infraestructura |
|---|---|---|
| IT-01 | Validar identidad con mock exitoso. | Mock provider. |
| IT-02 | Registro con cédula/correo duplicado. | PostgreSQL. |
| IT-03 | Bloqueo tras 3 intentos. | Redis. |
| IT-04 | Login y emisión de tokens. | PostgreSQL + Redis. |
| IT-05 | Bloqueo de login. | Redis. |
| IT-06 | Recuperación y envío de correo. | Mailpit/memory. |
| IT-07 | Restablecimiento y revocación. | Redis. |
| IT-08 | Auditoría de eventos. | PostgreSQL. |
| IT-09 | Creación única de borrador. | PostgreSQL. |
| IT-10 | Nacionalidad no ecuatoriana. | PostgreSQL. |
| IT-11 | Datos de pareja. | PostgreSQL. |
| IT-12 | Domicilio y cascada. | PostgreSQL. |
| IT-13 | Catálogos con caché. | Redis + PostgreSQL. |
| IT-14 | CRUD familiares. | PostgreSQL. |
| IT-15 | Elegibilidad AFI/JUB/EMP. | Mock IESS. |
| IT-16 | Respaldo organizacional. | Storage memory. |
| IT-17 | CRUD formación. | PostgreSQL. |
| IT-18 | Experiencia y tiempo. | PostgreSQL. |
| IT-19 | CRUD publicaciones. | PostgreSQL. |
| IT-20 | CRUD capacitación. | PostgreSQL. |
| IT-21 | Inhabilidades y efectos. | PostgreSQL. |
| IT-22 | Upload PDF válido/inválido. | Storage memory. |
| IT-23 | Guardar postulación completa. | PostgreSQL. |
| IT-24 | Envío transaccional. | PostgreSQL + storage. |
| IT-25 | Código y PDF. | PostgreSQL + PDF. |
| IT-26 | Inmutabilidad post-envío. | PostgreSQL. |
| IT-27 | Outbox y reintentos. | PostgreSQL + memory. |
| IT-28 | Validación servidor. | API. |
| IT-29 | Auditoría de cambios. | PostgreSQL. |
| IT-30 | Ownership de ítems. | API. |
| IT-31 | Países. | PostgreSQL. |
| IT-32 | Provincias. | PostgreSQL. |
| IT-33 | Cantones. | PostgreSQL. |
| IT-34 | Cache hit/miss. | Redis. |
| IT-35 | Invalidación por versión. | Redis. |

---

## 15.5 Pruebas E2E

| ID | Flujo | Criterio |
|---|---|---|
| E2E-01 | Landing → login/registro. | Navegación y textos. |
| E2E-02 | Registro con identidad válida. | Cuenta creada. |
| E2E-03 | Registro con correo duplicado. | Error claro. |
| E2E-04 | Bloqueo de registro. | Cuenta regresiva. |
| E2E-05 | Login exitoso. | Redirección al formulario. |
| E2E-06 | Recuperación de contraseña. | Correo y cambio. |
| E2E-07 | Ruta protegida sin sesión. | Redirección al login. |
| E2E-08 | Primer ingreso. | Borrador creado. |
| E2E-09 | Guardar y reanudar sección. | Datos persistentes. |
| E2E-10 | Información personal. | Datos verificados. |
| E2E-11 | Nacionalidad no ecuatoriana. | Bloqueo. |
| E2E-12 | Cascada geográfica. | Provincias/cantones correctos. |
| E2E-13 | Selección de cargo. | Confirmación. |
| E2E-14 | Elegibilidad. | Mensaje institucional. |
| E2E-15 | Inhabilidades. | 13 preguntas. |
| E2E-16 | Upload inválido. | Rechazo. |
| E2E-17 | Guardar postulación. | Estado LISTA_PARA_ENVIO. |
| E2E-18 | Enviar postulación. | Código y PDF. |
| E2E-19 | Mi postulación. | Estado y descarga. |

---

## 15.6 Matriz de pruebas críticas

| Escenario | Tipo | Prioridad |
|---|---|---|
| Registro exitoso | E2E | Alta |
| Bloqueo 3/10 | Integración/E2E | Alta |
| Login y refresh | Integración | Alta |
| Nacionalidad | Integración/E2E | Alta |
| Elegibilidad | Integración/E2E | Alta |
| Upload PDF | Integración | Alta |
| Envío transaccional | Integración | Alta |
| Rollback | Integración | Alta |
| Inmutabilidad | Integración | Alta |
| Caché | Integración | Media |
| Accesibilidad | E2E | Media |

---

## 15.7 Pruebas no funcionales

### Carga (k6)

```js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },
    { duration: '3m', target: 200 },
    { duration: '1m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const res = http.get(`${__ENV.BASE_URL}/api/v1/catalogos/paises`);
  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(1);
}
```

### Seguridad

- Fuerza bruta en login/registro.
- Autorización horizontal con dos usuarios.
- Upload de archivo falso.
- SQL injection y XSS.
- Rate limit.
- Dependencias vulnerables.

### Accesibilidad

- axe en landing, login, registro y formulario.
- Navegación por teclado.
- Contraste.
- Lectores de pantalla.

---

## 15.8 Datos de prueba

| Dato | Uso |
|---|---|
| Cédula válida `0102030405` | Registro exitoso. |
| Cédula válida `0912345678` | Pareja. |
| Cédula no elegible | Prueba AFI/JUB/EMP. |
| Correo `test@example.com` | Registro. |
| PDF válido de 1 MB | Upload. |
| PDF falso `.pdf` | Rechazo. |
| Fechas borde | Cálculo de experiencia. |

> Todos los datos deben ser ficticios. No usar información real de personas.

---

## 15.9 Cobertura

| Componente | Objetivo |
|---|---|
| Servicios backend | ≥ 70 % |
| Dominio (cálculos/reglas) | ≥ 80 % |
| Handlers | ≥ 60 % |
| Frontend crítico | ≥ 60 % |
| E2E flujos críticos | 100 % de los flujos Must |

---

## 15.10 Puertas de calidad

- `gofmt` y `go vet` sin errores.
- `go test ./... -race` verde.
- `npm run lint` verde.
- `npm run test` verde.
- Cobertura mínima.
- `govulncheck` sin críticos.
- `npm audit` sin críticos.
- Build de producción exitoso.
- Smoke test E2E.

---

## 15.11 Definición de terminado

Un requisito está terminado cuando:

1. Tiene código implementado.
2. Tiene prueba unitaria o de integración.
3. Tiene criterio de aceptación verificado.
4. Está auditado si aplica.
5. Está documentado en el SRS/SDD.
6. Pasa el pipeline.
7. Fue demostrado en la sustentación.

---

## 15.12 Riesgos de prueba

| Riesgo | Mitigación |
|---|---|
| Mocks no representan la realidad. | Contratos documentados y pruebas de contrato. |
| Redis/DB inestables en CI. | Testcontainers y health checks. |
| PDF depende de fuentes. | Imagen con fuentes instaladas. |
| Tiempo limitado. | Priorizar flujos Must y automatizar lo crítico. |
| Datos sensibles en pruebas. | Datos ficticios y anonimización. |

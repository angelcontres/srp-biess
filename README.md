# Sistema de Registro y Postulación al Directorio del BIESS (SRP-BIESS)

Plataforma web transaccional desacoplada para el **Concurso Público de Méritos y Oposición** del Banco del Instituto Ecuatoriano de Seguridad Social (BIESS), conforme a la Resolución No. C.D. 701 y las especificaciones institucionales EF-11 y EF-12.

| Campo | Detalle |
|---|---|
| **Asignatura** | Frameworks de Programación Web |
| **Evaluación** | Examen Práctico del Primer Parcial (40 Puntos) |
| **Integrantes del Equipo** | • Paulo Rafael Orrala Arriaga<br>• Carlos Fernando Patiño Garcia<br>• Jose Daniel Gonzalez Cochea<br>• Angel Alejandro Villon Panimboza |
| **Repositorio GitHub** | [https://github.com/angelcontres/srp-biess](https://github.com/angelcontres/srp-biess) |

---

## 1. Prerrequisitos del Sistema

| Software | Versión Recomendada | Propósito |
|---|---|---|
| **Docker & Docker Compose** | Docker 24+ / Compose v2+ | Contenerización de PostgreSQL 16, Redis 7 y Mailpit. |
| **Go** | 1.22+ / 1.23+ | Backend API de alto rendimiento (Fiber v2). |
| **Node.js & npm** | Node 20+ (LTS) / npm 10+ | Frontend Single Page Application (React + Vite + Tailwind). |
| **PostgreSQL (Cliente)** | 16+ | Base de datos relacional y transaccional. |
| **Redis (Cliente)** | 7+ | Caché en memoria y control de intentos/bloqueo. |

---

## 2. Panel de Control Interactivo (Consola de Gestión)

Para evitar configuraciones manuales, se incluye un menú interactivo en consola tanto para Windows (`start.bat`) como para Linux/macOS (`start.sh`):

### En Windows:
```cmd
start.bat
```

### En Linux / macOS / Git Bash:
```bash
chmod +x start.sh stop.sh
./start.sh
```

### Opciones del Menú en Consola:
* `[1]` **Iniciar en Modo Desarrollo:** Levanta PostgreSQL 16, Redis 7 y Mailpit Web UI.
* `[2]` **Iniciar en Modo Producción:** Modo productivo con remoción de huérfanos.
* `[3]` **Detener Todos los Servicios:** `docker compose down`.
* `[4]` **Reiniciar Servicios:** Reinicio ágil de los contenedores.
* `[5]` **Ver Estado y Salud:** Muestra `docker compose ps` con el healthcheck.
* `[6]` **Ver Logs en Tiempo Real:** Salida de logs combinada (`--tail=100 -f`).
* `[7]` **Resetear Base de Datos y Semillas:** Borra volúmenes y recarga `db/init.sql` desde cero.
* `[8]` **Consola PostgreSQL (psql):** Abre sesión interactiva directa en la base `srp_biess`.
* `[9]` **Consola Redis (redis-cli):** Abre sesión interactiva directa en Redis.
* `[0]` **Salir**.

---

## 3. Servicios Disponibles en Local

Una vez iniciado Docker Compose:

| Servicio | Puerto Local | Detalle / Credenciales por Defecto |
|---|---|---|
| **PostgreSQL 16** | `localhost:5432` | BD: `srp_biess` \| Usuario: `srp_user` \| Pass: `srp_secret_password` |
| **Redis 7** | `localhost:6379` | Sin contraseña en desarrollo \| DB: `0` |
| **Mailpit (Web UI)** | `http://localhost:8025` | **Bandeja de entrada web** para ver los correos de verificación y de envío de la postulación |
| **Mailpit (SMTP)** | `localhost:1025` | Servidor SMTP local para la API |

---

## 4. Base de Datos y Datos Semilla (DDL)

El script SQL completo con DDL normalizado, restricciones de integridad, transacciones y datos semilla (división política del Ecuador, cargos, preguntas de inhabilidad) se encuentra en:
* [`db/init.sql`](./db/init.sql)

Se monta automáticamente en `/docker-entrypoint-initdb.d/01-init.sql`, ejecutándose en el primer arranque del contenedor.

Si deseas conectarte directamente vía CLI o DBeaver:
```bash
psql -h localhost -p 5432 -U srp_user -d srp_biess
```

---

## 5. Variables de Entorno

El archivo modelo con todos los parámetros documentados es:
* [`.env.example`](./.env.example)

Los scripts `start.bat` y `start.sh` crean automáticamente el archivo `.env` local a partir de este modelo si no existe.

---

## 6. Estructura de Documentación del Proyecto

Toda la documentación técnica exigida por estándares de ingeniería de software se encuentra organizada en el directorio `docs/`:

* **[docs/README.md](./docs/README.md):** Índice general y visión global de la entrega.
* **[docs/PLANNING.md](./docs/PLANNING.md):** Plan de proyecto, WBS, Matriz RACI para 4 integrantes y Guion cronometrado para la sustentación oral de 10 minutos.
* **[docs/TICKETS.md](./docs/TICKETS.md):** Catálogo de 16 tickets de implementación organizados por épicas y asignados por integrante.
* **[docs/SRS/](./docs/SRS/README.md):** Especificación de Requisitos de Software (ISO/IEC/IEEE 29148:2018).
* **[docs/SDD/](./docs/SDD/README.md):** Diseño de Software (IEEE 1016-2009) con arquitectura en capas, inyección de dependencias y ADRs (incluyendo validación Módulo 10 y pagos/facturación con Mocks implementables).

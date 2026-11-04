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
| **Go** | 1.27.1+ | Versión declarada en `go.mod`; backend API con Fiber v2. |
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
* `[2]` **Iniciar en Modo Producción:** Modo productivo completo con Docker Compose (`docker-compose.prod.yml`).
* `[3]` **Detener Todos los Servicios:** `docker compose down`.
* `[4]` **Reiniciar Servicios:** Reinicio ágil de los contenedores.
* `[5]` **Ver Estado y Salud:** Muestra `docker compose ps` con el healthcheck.
* `[6]` **Ver Logs en Tiempo Real:** Salida de logs combinada (`--tail=100 -f`).
* `[7]` **Resetear Base de Datos y Semillas:** Borra volúmenes y recarga `db/init.sql` desde cero.
* `[8]` **Consola PostgreSQL (psql):** Abre sesión interactiva directa en la base `srp_biess`.
* `[9]` **Consola Redis (redis-cli):** Abre sesión interactiva directa en Redis.
* `[0]` **Salir**.

---

## 3. Flujo completo de instalación y ejecución

### 3.1 Preparar el entorno

1. Clonar el repositorio y entrar en su directorio:

```bash
git clone https://github.com/angelcontres/srp-biess.git
cd srp-biess
```

2. Crear las variables locales. Los scripts lo hacen automáticamente, pero también puede
   ejecutarse de forma manual:

```bash
# Linux/macOS/Git Bash
cp .env.example .env

# Windows CMD
copy .env.example .env
```

3. Levantar PostgreSQL 16, Redis 7 y Mailpit:

```bash
# Linux/macOS/Git Bash
./start.sh
# Seleccionar la opción 1

# Windows
start.bat
# Seleccionar la opción 1
```

También puede ejecutarse directamente:

```bash
docker compose up -d
docker compose ps
```

### 3.2 Crear o reiniciar la base de datos

La opción 7 de `start.sh` o `start.bat` elimina los volúmenes y vuelve a ejecutar
`db/init.sql`. Esta operación destruye los datos locales existentes.

Para hacerlo manualmente:

```bash
docker compose down -v
docker compose up -d postgres redis mailpit
```

`db/init.sql` contiene el DDL completo, catálogos geográficos, cargos, convocatoria y
preguntas de inhabilidades. Se ejecuta automáticamente cuando PostgreSQL inicia con un
volumen nuevo.

### 3.3 Aplicar migraciones

Las migraciones deben ejecutarse después de crear la base y en orden numérico:

```bash
for migration in db/migrations/*.sql; do
  docker exec -i srp-biess-postgres \
    psql -v ON_ERROR_STOP=1 -U srp_user -d srp_biess < "$migration"
done
```

En Windows PowerShell:

```powershell
Get-ChildItem db/migrations/*.sql | Sort-Object Name | ForEach-Object {
  Get-Content $_.FullName | docker exec -i srp-biess-postgres `
    psql -v ON_ERROR_STOP=1 -U srp_user -d srp_biess
}
```

### 3.4 Ejecutar los seeders

El seed de catálogos ya forma parte de `db/init.sql`. El usuario de prueba para desarrollo
se carga por separado:

```bash
docker exec -i srp-biess-postgres \
  psql -v ON_ERROR_STOP=1 -U srp_user -d srp_biess \
  < db/seeds/postulante_prueba.sql
```

Credenciales del usuario de prueba:

```text
Cédula:     0926687856
Contraseña: BiessPostulante2026*
Correo:     postulante.test@biess.gob.ec
```

### 3.5 Encender el backend

Con Docker y PostgreSQL/Redis activos, desde la raíz del proyecto:

```bash
go run ./cmd/api
```

La API queda disponible en `http://localhost:8080`. Para compilarla:

```bash
go build ./cmd/api
```

### 3.6 Encender el frontend

En otra terminal:

```bash
cd frontend
npm install
npm run dev
```

La SPA queda disponible normalmente en `http://localhost:5173`.

Comandos frontend disponibles:

```bash
npm run typecheck
npm run test -- --run
npm run build
```

### 3.7 Detener el entorno

```bash
# Linux/macOS/Git Bash
./stop.sh

# Windows
stop.bat
```

## 4. Servicios disponibles en local

Una vez iniciado Docker Compose:

| Servicio | Puerto Local | Detalle / Credenciales por Defecto |
|---|---|---|
| **PostgreSQL 16** | `localhost:5432` | BD: `srp_biess` \| Usuario: `srp_user` \| Pass: `srp_secret_password` |
| **Redis 7** | `localhost:6379` | Sin contraseña en desarrollo \| DB: `0` |
| **Mailpit (Web UI)** | `http://localhost:8025` | **Bandeja de entrada web** para ver los correos de verificación y de envío de la postulación |
| **Mailpit (SMTP)** | `localhost:1025` | Servidor SMTP local para la API |

---

## 5. Base de datos y datos semilla (DDL)

El script SQL completo con DDL normalizado, restricciones de integridad, transacciones y datos semilla (división política del Ecuador, cargos, preguntas de inhabilidad) se encuentra en:
* [`db/init.sql`](./db/init.sql)

Se monta automáticamente en `/docker-entrypoint-initdb.d/01-init.sql`, ejecutándose en el primer arranque del contenedor.

Si deseas conectarte directamente vía CLI o DBeaver:
```bash
psql -h localhost -p 5432 -U srp_user -d srp_biess
```

---

## 6. Variables de entorno

El archivo modelo con todos los parámetros documentados es:
* [`.env.example`](./.env.example)

Los scripts `start.bat` y `start.sh` crean automáticamente el archivo `.env` local a partir de este modelo si no existe.

---

## 7. Historial Git y ramas

Repositorio:

* [https://github.com/angelcontres/srp-biess](https://github.com/angelcontres/srp-biess)

Consultar el historial con commits, autores y ramas:

```bash
git log --oneline --decorate --graph --all
```

Listar las ramas locales ordenadas por fecha del último commit:

```bash
git branch --sort=-committerdate -vv
```

Listar también las ramas remotas:

```bash
git branch --all --sort=-committerdate
```

Ramas principales:

* `main`: referencia de producción.
* `develop`: integración y pruebas.
* `feature/*`, `fix/*` y `style/*`: trabajo aislado por funcionalidad, corrección o interfaz.

## 8. Pipeline de CI/CD (GitHub Actions & Runner Self-Hosted)

El repositorio cuenta con integración continua y despliegue automatizado en [`.github/workflows/deploy.yml`](./.github/workflows/deploy.yml) ejecutándose en un runner propio Linux (`self-hosted`):

### Fases del Pipeline:
1. **CI (Integración Continua):** Se activa automáticamente ante cualquier `push` o `pull_request` hacia `main` o `develop`.
   - **Arquitectura Limpia (ADR-003):** Verifica que `internal/domain` no importe dependencias prohibidas (SQL, Fiber, Redis ni infraestructura).
   - **Backend API (Go):** Análisis estático con `go vet`, pruebas unitarias con `-race` y compilación de `cmd/api`.
   - **Frontend SPA (React):** Verificación de tipos TypeScript (`tsc --noEmit`), suite de pruebas unitarias (`vitest run`) y empaquetado optimizado de producción (`vite build`).
2. **CD (Despliegue Continuo):** Se ejecuta automáticamente al hacer `push` a la rama `main` o de forma manual mediante el disparador `workflow_dispatch` en la pestaña Actions de GitHub.
   - Orquesta el stack contenerizado completo mediante [`docker-compose.prod.yml`](./docker-compose.prod.yml).
   - Genera imágenes Docker multi-stage optimizadas para la API de Go (`Dockerfile`) y para la SPA con Nginx (`frontend/Dockerfile`).
   - Mantiene aislamiento de red seguro para PostgreSQL y Redis dentro de la red interna de Docker (`srp-network`), evitando colisiones en el host.
   - Valida la disponibilidad operativa de los servicios mediante health checks HTTP en `/health` y `/healthz`.

---

## 9. Estructura de documentación del proyecto

Toda la documentación técnica exigida por estándares de ingeniería de software se encuentra organizada en el directorio `docs/`:

* **[docs/README.md](./docs/README.md):** Índice general y visión global de la entrega.
* **[docs/PLANNING.md](./docs/PLANNING.md):** Plan de proyecto, WBS, Matriz RACI para 4 integrantes y Guion cronometrado para la sustentación oral de 10 minutos.
* **[docs/TICKETS.md](./docs/TICKETS.md):** Catálogo de 16 tickets de implementación organizados por épicas y asignados por integrante.
* **[docs/SRS/](./docs/SRS/README.md):** Especificación de Requisitos de Software (ISO/IEC/IEEE 29148:2018).
* **[docs/SDD/](./docs/SDD/README.md):** Diseño de Software (IEEE 1016-2009) con arquitectura en capas, inyección de dependencias y ADRs (incluyendo validación Módulo 10 y pagos/facturación con Mocks implementables).

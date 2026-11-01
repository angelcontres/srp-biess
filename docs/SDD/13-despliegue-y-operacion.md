# 13. Despliegue y operación — SDD

## 13.1 Entornos

| Entorno | Propósito | Datos | Servicios |
|---|---|---|---|
| **Local** | Desarrollo y demo. | Ficticios. | Docker Compose. |
| **Test** | Integración y QA. | Ficticios. | Contenedores + mocks. |
| **Producción** | Operación real. | Reales. | Infraestructura institucional. |

---

## 13.2 Docker Compose local

```yaml
services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: srp_biess
      POSTGRES_USER: srp
      POSTGRES_PASSWORD: srp_dev
    ports: ["5432:5432"]
    volumes: ["pgdata:/var/lib/postgresql/data"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U srp -d srp_biess"]
      interval: 5s
      timeout: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    command: ["redis-server", "--appendonly", "yes"]
    ports: ["6379:6379"]
    volumes: ["redisdata:/data"]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 10

  mailpit:
    image: axllent/mailpit:latest
    ports: ["1025:1025", "8025:8025"]

  api:
    build: ./backend
    env_file: .env
    depends_on:
      postgres: { condition: service_healthy }
      redis: { condition: service_healthy }
    ports: ["8080:8080"]

  web:
    build: ./frontend
    depends_on: [api]
    ports: ["5173:80"]

volumes:
  pgdata:
  redisdata:
```

### Uso

```bash
cp .env.example .env
docker compose up --build
```

- API: `http://localhost:8080`
- SPA: `http://localhost:5173`
- Mailpit: `http://localhost:8025`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

---

## 13.3 Variables de entorno

| Variable | Ejemplo | Secreto |
|---|---|---|
| `APP_ENV` | `development` | No |
| `HTTP_ADDR` | `:8080` | No |
| `DATABASE_URL` | `postgres://srp:srp_dev@postgres:5432/srp_biess?sslmode=disable` | Sí |
| `REDIS_ADDR` | `redis:6379` | No |
| `REDIS_PASSWORD` | `***` | Sí |
| `JWT_PRIVATE_KEY` | PEM | Sí |
| `JWT_PUBLIC_KEY` | PEM | Sí |
| `JWT_ISSUER` | `srp-biess` | No |
| `ACCESS_TOKEN_TTL` | `15m` | No |
| `REFRESH_TOKEN_TTL` | `8h` | No |
| `REGISTRO_CIVIL_MODE` | `mock` | No |
| `REGISTRO_CIVIL_URL` | `https://...` | Sí |
| `IESS_MODE` | `mock` | No |
| `IESS_URL` | `https://...` | Sí |
| `SMTP_HOST` | `mailpit` | No |
| `SMTP_PORT` | `1025` | No |
| `SMTP_FROM` | `no-reply@biess.gob.ec` | No |
| `STORAGE_MODE` | `local` | No |
| `STORAGE_PATH` | `./storage` | No |
| `MAX_TOTAL_UPLOAD_MB` | `50` | No |
| `CORS_ORIGINS` | `http://localhost:5173` | No |

> `.env.example` debe contener valores de ejemplo, nunca secretos reales.

---

## 13.4 Build y ejecución

### Backend

```bash
cd backend
go mod download
go build -o bin/api ./cmd/api
./bin/api
```

### Frontend

```bash
cd frontend
npm ci
npm run dev
npm run build
npm run preview
```

### Calidad

```bash
gofmt -w .
go vet ./...
go test ./... -race -cover
npm run lint
npm run test
npm run build
```

---

## 13.5 Migraciones y seed

```bash
migrate -path migrations -database "$DATABASE_URL" up
psql "$DATABASE_URL" -f seeds/001_catalogos.sql
psql "$DATABASE_URL" -f seeds/002_inhabilidades.sql
psql "$DATABASE_URL" -f seeds/003_cargos.sql
```

### Reglas

- Las migraciones son inmutables una vez aplicadas.
- El seed es idempotente (`ON CONFLICT DO NOTHING`).
- En producción, las migraciones se ejecutan con un usuario dedicado.
- Se prueba en una base limpia antes de desplegar.

---

## 13.6 Reverse proxy y TLS

```nginx
server {
  listen 443 ssl http2;
  server_name concurso.biess.gob.ec;

  ssl_certificate /etc/ssl/certs/biess.crt;
  ssl_certificate_key /etc/ssl/private/biess.key;
  ssl_protocols TLSv1.2 TLSv1.3;

  root /usr/share/nginx/html;
  index index.html;

  location / {
    try_files $uri /index.html;
  }

  location /api/ {
    proxy_pass http://api:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-Proto $scheme;
  }
}
```

- HSTS en producción.
- CSP.
- Compresión gzip/brotli.
- Límite de body para uploads.

---

## 13.7 CI/CD

```yaml
stages:
  - lint
  - test
  - build
  - security
  - deploy

lint:
  script:
    - gofmt -l .
    - go vet ./...
    - npm run lint

test:
  script:
    - go test ./... -race -cover
    - npm run test

build:
  script:
    - go build ./cmd/api
    - npm run build

security:
  script:
    - govulncheck ./...
    - npm audit --audit-level=high

deploy:
  when: manual
  script:
    - docker compose pull
    - docker compose up -d
```

### Quality gates

- Sin errores de lint.
- Tests verdes.
- Cobertura mínima.
- Sin vulnerabilidades críticas.
- Build exitoso.

---

## 13.8 Monitoreo

| Señal | Herramienta | Alerta |
|---|---|---|
| Uptime | Health checks | Caída > 1 min. |
| Errores 5xx | Métricas | > 5 % en 5 min. |
| Latencia | Métricas | p95 > 1 s. |
| Redis | Redis exporter | Caída > 1 min. |
| PostgreSQL | Postgres exporter | Pool > 80 %. |
| Correos | Outbox | > 10 fallidos en 15 min. |
| Disco | Node exporter | > 80 %. |

---

## 13.9 Backup y restauración

### PostgreSQL

- Backup completo diario.
- WAL continuo.
- Retención 30 días.
- Restauración probada trimestralmente.

### Redis

- AOF habilitado.
- Snapshot diario.
- No es fuente de verdad para datos de negocio.

### Storage

- Versionado o replicación.
- Backup de PDF según retención.

### Restauración

1. Detener escrituras.
2. Restaurar backup.
3. Aplicar WAL.
4. Verificar integridad.
5. Reanudar servicio.
6. Documentar.

---

## 13.10 Runbook

| Síntoma | Diagnóstico | Acción |
|---|---|---|
| API no responde | Health check, logs. | Reiniciar réplica; revisar DB/Redis. |
| Error 500 masivo | Logs con `trace_id`. | Identificar causa; rollback si es deploy. |
| Redis caído | Ping, métricas. | Catálogos desde DB; auth fail-closed; restaurar Redis. |
| PostgreSQL caído | Conexiones, logs. | Failover; restaurar; verificar integridad. |
| Correos no salen | Outbox, SMTP. | Reintentar; revisar credenciales. |
| Disco lleno | `df`, storage. | Limpiar temporales; ampliar volumen. |
| Upload falla | Tamaño, MIME, storage. | Revisar límites y permisos. |
| PDF no genera | Logs del generador. | Reintentar; revisar plantilla. |

---

## 13.11 Hardening

- Contenedores con usuario no root.
- Imágenes mínimas (alpine/distroless).
- Sin puertos de DB/Redis expuestos.
- Secretos fuera del repositorio.
- TLS obligatorio.
- Firewall y listas de allowlist.
- Escaneo de imágenes.
- Actualizaciones de dependencias.
- Logs sin datos sensibles.
- Backups cifrados.

---

## 13.12 Checklist de despliegue

- [ ] Variables de entorno configuradas.
- [ ] Migraciones aplicadas.
- [ ] Seed ejecutado.
- [ ] Health checks verdes.
- [ ] TLS activo.
- [ ] CORS restringido.
- [ ] Storage con permisos correctos.
- [ ] SMTP probado.
- [ ] Backups configurados.
- [ ] Alertas activas.
- [ ] Smoke test de registro y envío.
- [ ] Rollback documentado.

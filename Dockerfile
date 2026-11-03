# =====================================================================
# SRP-BIESS · Backend API Dockerfile (Go 1.22+ / Fiber v2)
# Multi-stage build para imagen mínima, segura y de alto rendimiento
# =====================================================================

# ---------------------------------------------------------------------
# Etapa 1: Compilación del binario estático
# ---------------------------------------------------------------------
FROM golang:alpine AS builder

WORKDIR /app

# Dependencias necesarias para la compilación estática
RUN apk add --no-cache git ca-certificates tzdata

# Descarga de módulos (aprovecha la caché de capas de Docker)
COPY go.mod go.sum ./
RUN go mod download

# Copia de código fuente
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Compilación de binario estático sin CGO y con símbolos podados (-w -s)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/api ./cmd/api

# ---------------------------------------------------------------------
# Etapa 2: Imagen de ejecución mínima (Alpine Linux)
# ---------------------------------------------------------------------
FROM alpine:3.20 AS runner

WORKDIR /app

# Dependencias de ejecución: certificados TLS, zona horaria y utilidades
RUN apk add --no-cache ca-certificates tzdata wget

# Zona horaria de Ecuador (conforme a las reglas del proyecto)
ENV TZ=America/Guayaquil

# Directorios para uploads y almacenamiento de documentos PDF
RUN mkdir -p /app/uploads /app/storage

# Copia del binario compilado
COPY --from=builder /bin/api /app/api

EXPOSE 8080

# Health check contra el endpoint /health nativo de Fiber
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget -q --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/api"]

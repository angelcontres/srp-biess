# =====================================================================
# SRP-BIESS · Frontend SPA Dockerfile (React 18 + Vite + Tailwind CSS)
# Multi-stage build con servidor Nginx Alpine optimizado para producción
# =====================================================================

# ---------------------------------------------------------------------
# Etapa 1: Instalación y empaquetado de producción
# ---------------------------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Copia de manifiestos para caché eficiente de npm
COPY package.json package-lock.json ./

# Instalación limpia de dependencias
RUN npm ci

# Copia de código fuente de la SPA
COPY . .

# Verificación de tipos y generación del bundle estático en /app/dist
RUN npm run build

# ---------------------------------------------------------------------
# Etapa 2: Servidor web ligero Nginx
# ---------------------------------------------------------------------
FROM nginx:alpine AS runner

# Limpiar archivos estáticos por defecto
RUN rm -rf /usr/share/nginx/html/*

# Copiar bundle compilado desde la etapa builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copiar configuración de Nginx con reverse proxy y fallback de SPA
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

# Health check para el contenedor web
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD wget -q --spider http://localhost/healthz || exit 1

CMD ["nginx", "-g", "daemon off;"]

#!/usr/bin/env bash

# =====================================================================
# SRP-BIESS · Script de Parada de Entorno Local (Linux / macOS / Git Bash)
# =====================================================================

set -e

YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=====================================================================${NC}"
echo -e "${BLUE} SRP-BIESS | Deteniendo Entorno Local (Linux / macOS)${NC}"
echo -e "${BLUE}=====================================================================${NC}\n"

echo -e "${YELLOW}[*] Deteniendo contenedores de Docker Compose...${NC}"
docker compose down
docker compose -f docker-compose.prod.yml down >/dev/null 2>&1

echo -e "\n${GREEN}[OK] Contenedores detenidos correctamente.${NC}\n"


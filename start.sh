#!/usr/bin/env bash

# =====================================================================
# SRP-BIESS · Panel de Control Interactivo (Linux / macOS / Git Bash)
# =====================================================================

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}[ERROR] Docker no está instalado o no se encuentra en el PATH.${NC}"
        echo "Por favor inicia Docker antes de continuar."
        read -rp "Presiona [Enter] para volver al menú..."
        return 1
    fi
    return 0
}

check_env() {
    if [ ! -f .env ]; then
        echo -e "${YELLOW}[*] Creando .env a partir de .env.example...${NC}"
        cp .env.example .env
        echo -e "${GREEN}[OK] .env creado.${NC}"
    fi
    if [ ! -d uploads ]; then
        mkdir -p uploads
    fi
}

show_menu() {
    clear
    echo -e "${BLUE}=====================================================================${NC}"
    echo -e "${CYAN}           SRP-BIESS | PANEL DE CONTROL DEL ENTORNO${NC}"
    echo -e "${BLUE}=====================================================================${NC}"
    echo ""
    echo -e "  ${GREEN}[1]${NC} Iniciar Entorno de Desarrollo (Dev: Postgres, Redis, Mailpit)"
    echo -e "  ${GREEN}[2]${NC} Iniciar en Modo Producción (Prod)"
    echo -e "  ${GREEN}[3]${NC} Detener todos los Servicios (Down)"
    echo -e "  ${GREEN}[4]${NC} Reiniciar Servicios (Restart)"
    echo -e "  ${GREEN}[5]${NC} Ver Estado y Salud de Contenedores (Status)"
    echo -e "  ${GREEN}[6]${NC} Ver Logs en Tiempo Real (Logs)"
    echo -e "  ${YELLOW}[7]${NC} Resetear Base de Datos y Semillas (Recrear Volúmenes)"
    echo -e "  ${CYAN}[8]${NC} Abrir Consola PostgreSQL (psql CLI)"
    echo -e "  ${CYAN}[9]${NC} Abrir Consola Redis (redis-cli)"
    echo -e "  ${RED}[0]${NC} Salir"
    echo ""
    echo -e "${BLUE}=====================================================================${NC}"
}

while true; do
    show_menu
    read -rp "Selecciona una opción [0-9]: " OPCION
    case $OPCION in
        1)
            check_docker || continue
            check_env
            echo -e "\n${YELLOW}[*] Levantando entorno de desarrollo...${NC}"
            docker compose up -d
            echo -e "\n${GREEN}=====================================================================${NC}"
            echo -e "${GREEN} SERVICIOS ACTIVOS (MODO DESARROLLO)${NC}"
            echo -e "${GREEN}=====================================================================${NC}"
            echo -e " - PostgreSQL: localhost:5432 (BD: srp_biess, User: srp_user)"
            echo -e " - Redis:      localhost:6379"
            echo -e " - Mailpit UI: ${CYAN}http://localhost:8025${NC}"
            echo -e "${GREEN}=====================================================================${NC}\n"
            read -rp "Presiona [Enter] para continuar..."
            ;;
        2)
            check_docker || continue
            check_env
            echo -e "\n${YELLOW}[*] Levantando contenedores en modo producción...${NC}"
            docker compose -f docker-compose.prod.yml up -d --build --remove-orphans
            echo -e "\n${GREEN}[OK] Entorno iniciado en modo producción.${NC}"
            echo -e "Recuerda configurar tus credenciales reales en el archivo .env\n"
            read -rp "Presiona [Enter] para continuar..."
            ;;
        3)
            check_docker || continue
            echo -e "\n${YELLOW}[*] Deteniendo contenedores...${NC}"
            docker compose down
            echo -e "${GREEN}[OK] Servicios detenidos.${NC}\n"
            read -rp "Presiona [Enter] para continuar..."
            ;;
        4)
            check_docker || continue
            echo -e "\n${YELLOW}[*] Reiniciando contenedores...${NC}"
            docker compose restart
            echo -e "${GREEN}[OK] Servicios reiniciados.${NC}\n"
            read -rp "Presiona [Enter] para continuar..."
            ;;
        5)
            check_docker || continue
            echo -e "\n${BLUE}=====================================================================${NC}"
            echo -e "${CYAN} ESTADO ACTUAL DE LOS CONTENEDORES${NC}"
            echo -e "${BLUE}=====================================================================${NC}"
            docker compose ps
            echo -e "${BLUE}=====================================================================${NC}\n"
            read -rp "Presiona [Enter] para continuar..."
            ;;
        6)
            check_docker || continue
            echo -e "\n${YELLOW}[*] Mostrando logs en tiempo real (Presiona Ctrl + C para regresar)...${NC}\n"
            docker compose logs -f --tail=100
            ;;
        7)
            check_docker || continue
            echo -e "\n${RED}[ATENCIÓN] Esto eliminará todos los datos de la BD y la recreará con las semillas.${NC}"
            read -rp "¿Deseas continuar? (s/N): " CONFIRM
            if [[ "$CONFIRM" =~ ^[sS]$ ]]; then
                echo -e "${YELLOW}[*] Eliminando volúmenes y recreando contenedores...${NC}"
                docker compose down -v
                docker compose up -d
                echo -e "${GREEN}[OK] Base de datos reseteada con éxito.${NC}\n"
            else
                echo -e "${BLUE}[*] Operación cancelada.${NC}\n"
            fi
            read -rp "Presiona [Enter] para continuar..."
            ;;
        8)
            check_docker || continue
            echo -e "\n${CYAN}[*] Conectando a PostgreSQL CLI (escribe '\q' para salir)...${NC}"
            docker exec -it srp-biess-postgres psql -U srp_user -d srp_biess
            ;;
        9)
            check_docker || continue
            echo -e "\n${CYAN}[*] Conectando a Redis CLI (escribe 'exit' para salir)...${NC}"
            docker exec -it srp-biess-redis redis-cli
            ;;
        0)
            echo -e "\n${GREEN}Saliendo del panel. ¡Hasta luego!${NC}\n"
            exit 0
            ;;
        *)
            echo -e "\n${RED}[!] Opción inválida. Intenta nuevamente.${NC}"
            sleep 1
            ;;
    esac
done


@echo off
echo =====================================================================
echo  SRP-BIESS ^| Deteniendo Entorno Local (Windows)
echo =====================================================================
echo.
echo [*] Deteniendo contenedores de Docker Compose...
docker compose down
docker compose -f docker-compose.prod.yml down >nul 2>&1

echo.
echo [OK] Contenedores detenidos correctamente.
echo.
pause


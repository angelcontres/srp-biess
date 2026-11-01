@echo off
echo =====================================================================
echo  SRP-BIESS ^| Deteniendo Entorno Local (Windows)
echo =====================================================================
echo.
echo [*] Deteniendo contenedores de Docker Compose...
docker compose down

echo.
echo [OK] Contenedores detenidos correctamente.
echo.
pause

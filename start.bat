@echo off
setlocal enabledelayedexpansion
title SRP-BIESS - Panel de Control del Entorno

:MENU
cls
echo =====================================================================
echo           SRP-BIESS ^| PANEL DE CONTROL DEL ENTORNO
echo =====================================================================
echo.
echo   [1] Iniciar Entorno de Desarrollo (Dev: Postgres, Redis, Mailpit)
echo   [2] Iniciar en Modo Produccion (Prod)
echo   [3] Detener todos los Servicios (Down)
echo   [4] Reiniciar Servicios (Restart)
echo   [5] Ver Estado y Salud de Contenedores (Status)
echo   [6] Ver Logs en Tiempo Real (Logs)
echo   [7] Resetear Base de Datos y Semillas (Recrear Volumenes)
echo   [8] Abrir Consola PostgreSQL (psql CLI)
echo   [9] Abrir Consola Redis (redis-cli)
echo   [0] Salir
echo.
echo =====================================================================
set /p OPCION="Selecciona una opcion [0-9]: "

if "%OPCION%"=="1" goto DEV_START
if "%OPCION%"=="2" goto PROD_START
if "%OPCION%"=="3" goto STOP_ALL
if "%OPCION%"=="4" goto RESTART_ALL
if "%OPCION%"=="5" goto STATUS
if "%OPCION%"=="6" goto LOGS
if "%OPCION%"=="7" goto RESET_DB
if "%OPCION%"=="8" goto CLI_POSTGRES
if "%OPCION%"=="9" goto CLI_REDIS
if "%OPCION%"=="0" goto EXIT_SCRIPT

echo [!] Opcion invalida. Intenta nuevamente.
timeout /t 2 >nul
goto MENU

:CHECK_DOCKER
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker no esta disponible o no esta corriendo.
    echo Inicia Docker Desktop antes de continuar.
    pause
    goto MENU
)
exit /b 0

:CHECK_ENV
if not exist .env (
    echo [*] Creando .env a partir de .env.example...
    copy .env.example .env >nul
    echo [OK] .env creado.
)
if not exist uploads (
    mkdir uploads
)
exit /b 0

:DEV_START
call :CHECK_DOCKER
call :CHECK_ENV
echo.
echo [*] Levantando entorno de desarrollo...
docker compose up -d
echo.
echo =====================================================================
echo  SERVICIOS ACTIVOS (MODO DESARROLLO)
echo =====================================================================
echo  - PostgreSQL: localhost:5432 (BD: srp_biess, User: srp_user)
echo  - Redis:      localhost:6379
echo  - Mailpit UI: http://localhost:8025
echo =====================================================================
echo.
pause
goto MENU

:PROD_START
call :CHECK_DOCKER
call :CHECK_ENV
echo.
echo [*] Preparando variables para modo produccion...
echo [*] Levantando contenedores en modo produccion...
docker compose -f docker-compose.prod.yml up -d --build --remove-orphans
echo.
echo [OK] Entorno iniciado en modo produccion.
echo Recuerda configurar credenciales de produccion en el archivo .env
echo.
pause
goto MENU

:STOP_ALL
call :CHECK_DOCKER
echo.
echo [*] Deteniendo contenedores...
docker compose down
echo [OK] Servicios detenidos.
echo.
pause
goto MENU

:RESTART_ALL
call :CHECK_DOCKER
echo.
echo [*] Reiniciando contenedores...
docker compose restart
echo [OK] Servicios reiniciados.
echo.
pause
goto MENU

:STATUS
call :CHECK_DOCKER
echo.
echo =====================================================================
echo  ESTADO ACTUAL DE LOS CONTENEDORES
echo =====================================================================
docker compose ps
echo =====================================================================
echo.
pause
goto MENU

:LOGS
call :CHECK_DOCKER
echo.
echo [*] Mostrando logs en tiempo real (Presiona Ctrl + C para regresar)...
echo.
docker compose logs -f --tail=100
goto MENU

:RESET_DB
call :CHECK_DOCKER
echo.
echo [ATENCION] Esto eliminara todos los datos de la BD y la recreara con las semillas de db/init.sql.
set /p CONFIRM="Deseas continuar? (S/N): "
if /i "%CONFIRM%"=="S" (
    echo [*] Eliminando volumenes y recreando...
    docker compose down -v
    docker compose up -d
    echo [OK] Base de datos reseteada con exito.
) else (
    echo [*] Operacion cancelada.
)
echo.
pause
goto MENU

:CLI_POSTGRES
call :CHECK_DOCKER
echo.
echo [*] Conectando a PostgreSQL CLI (escribe '\q' para salir)...
docker exec -it srp-biess-postgres psql -U srp_user -d srp_biess
goto MENU

:CLI_REDIS
call :CHECK_DOCKER
echo.
echo [*] Conectando a Redis CLI (escribe 'exit' para salir)...
docker exec -it srp-biess-redis redis-cli
goto MENU

:EXIT_SCRIPT
echo.
echo Saliendo del panel. Hasta luego.
exit /b 0


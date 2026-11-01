@echo off
setlocal enabledelayedexpansion

echo ========================================================
echo   SRP-BIESS: ARNES DE VERIFICACION RAPIDA (TOKEN SAVER)
echo ========================================================

set HAS_ERR=0

echo [1/3] Verificando Arquitectura Limpia en internal\domain...
if exist internal\domain (
    dir /b internal\domain\*.go >nul 2>&1
    if !errorlevel! equ 0 (
        findstr /s /i /m "database/sql" internal\domain\*.go >nul 2>&1
        if !errorlevel! equ 0 (
            echo ERROR: Se detectaron imports de database/sql en internal/domain
            set HAS_ERR=1
        )
    )
)
if !HAS_ERR! equ 0 (
    echo OK: Dominio limpio y desacoplado
)

echo [2/3] Verificando compilacion de Go...
if exist go.mod (
    go build ./... > build_err.log 2>&1
    if !errorlevel! neq 0 (
        echo ERROR DE COMPILACION:
        findstr /i "error undefined cannot syntax" build_err.log
        if exist build_err.log del /f /q build_err.log
        set HAS_ERR=1
    ) else (
        if exist build_err.log del /f /q build_err.log
        echo OK: Backend compila sin errores
    )
) else (
    echo INFO: go.mod no inicializado todavia
)

echo [3/3] Verificando Docker...
docker compose ps > docker_status.log 2>&1
findstr /i "postgres" docker_status.log >nul 2>&1
if !errorlevel! equ 0 (
    echo OK: PostgreSQL activo en Docker
) else (
    echo AVISO: PostgreSQL no esta corriendo. Usa: docker compose up -d
)
if exist docker_status.log del /f /q docker_status.log

echo ========================================================
if !HAS_ERR! equ 0 (
    echo RESULTADO: VERIFICACION EXITOSA
) else (
    echo RESULTADO: FALLO DETECTADO - COPIA EL ERROR A TU IA
)
echo ========================================================
exit /b !HAS_ERR!

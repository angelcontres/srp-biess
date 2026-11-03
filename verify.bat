@echo off
setlocal enabledelayedexpansion

echo ========================================================
echo   SRP-BIESS: ARNES DE VERIFICACION RAPIDA (TOKEN SAVER)
echo ========================================================

set HAS_ERR=0
set MODE_FULL=0

if "%1"=="--full" set MODE_FULL=1
if "%1"=="--all" set MODE_FULL=1

REM ----------------------------------------------------------
REM [1/5] Verificacion de Arquitectura Limpia (ADR-003)
REM ----------------------------------------------------------
echo [1/5] Verificando Arquitectura Limpia en internal\domain...
if exist internal\domain (
    set FORBIDDEN_FOUND=0
    for /f "tokens=*" %%F in ('dir /b /s internal\domain\*.go 2^>nul') do (
        findstr /m /i "database/sql github.com/jackc/pgx github.com/gofiber/fiber github.com/redis/go-redis net/http srp-biess/internal/infrastructure srp-biess/internal/interfaces" "%%F" >nul 2>&1
        if !errorlevel! equ 0 (
            echo   ERROR ARQUITECTURA: %%F contiene dependencias prohibidas.
            set FORBIDDEN_FOUND=1
            set HAS_ERR=1
        )
    )
    if !FORBIDDEN_FOUND! equ 0 (
        echo   OK: Dominio puro y desacoplado - Sin SQL, HTTP ni frameworks
    )
) else (
    echo   AVISO: Directorio internal\domain no encontrado.
)

REM ----------------------------------------------------------
REM [2/5] Compilacion de Paquetes Go
REM ----------------------------------------------------------
echo [2/5] Verificando compilacion de Go...
if exist go.mod (
    go build ./... > build_err.log 2>&1
    if !errorlevel! neq 0 (
        echo   ERROR DE COMPILACION:
        findstr /i "error undefined cannot syntax" build_err.log
        set HAS_ERR=1
    ) else (
        go build -o nul ./cmd/api > build_err.log 2>&1
        if !errorlevel! neq 0 (
            echo   ERROR DE COMPILACION en cmd/api:
            findstr /i "error undefined cannot syntax" build_err.log
            set HAS_ERR=1
        ) else (
            echo   OK: Backend y Composition Root compilan sin errores
        )
    )
    if exist build_err.log del /f /q build_err.log 2>nul
) else (
    echo   AVISO: go.mod no encontrado.
)

REM ----------------------------------------------------------
REM [3/5] Analisis Estatico (go vet)
REM ----------------------------------------------------------
echo [3/5] Ejecutando analisis estatico con go vet...
if exist go.mod (
    go vet ./internal/... ./cmd/... > vet_err.log 2>&1
    if !errorlevel! neq 0 (
        echo   AVISO O ERROR EN GO VET:
        type vet_err.log
        set HAS_ERR=1
    ) else (
        echo   OK: Analisis estatico aprobado
    )
    if exist vet_err.log del /f /q vet_err.log 2>nul
)

REM ----------------------------------------------------------
REM [4/5] Pruebas Automatizadas
REM ----------------------------------------------------------
if !MODE_FULL! equ 1 (
    echo [4/5] Ejecutando suite completa de pruebas: go test ./...
    go test ./... > test_out.log 2>&1
    if !errorlevel! neq 0 (
        echo   ERROR EN PRUEBAS - SUITE COMPLETA:
        findstr /i "FAIL ---" test_out.log
        set HAS_ERR=1
    ) else (
        echo   OK: Todas las pruebas unitarias y de integracion pasaron exitosamente
    )
    if exist test_out.log del /f /q test_out.log 2>nul
) else (
    echo [4/5] Ejecutando pruebas rapidas de Dominio y Aplicacion...
    go test ./internal/domain/... ./internal/application/... > test_out.log 2>&1
    if !errorlevel! neq 0 (
        echo   ERROR EN PRUEBAS DE DOMINIO/APLICACION:
        findstr /i "FAIL ---" test_out.log
        set HAS_ERR=1
    ) else (
        echo   OK: Pruebas rapidas de logica de negocio y validacion aprobadas
    )
    if exist test_out.log del /f /q test_out.log 2>nul
)

REM ----------------------------------------------------------
REM [5/5] Estado de Infraestructura Docker
REM ----------------------------------------------------------
echo [5/5] Verificando servicios de infraestructura en Docker...
docker compose ps > docker_status.log 2>&1
if !errorlevel! equ 0 (
    findstr /i "postgres" docker_status.log >nul 2>&1
    if !errorlevel! equ 0 (
        echo   OK: PostgreSQL activo
    ) else (
        echo   AVISO: PostgreSQL no esta activo. Iniciar con: docker compose up -d postgres
    )

    findstr /i "redis" docker_status.log >nul 2>&1
    if !errorlevel! equ 0 (
        echo   OK: Redis activo
    ) else (
        echo   AVISO: Redis no esta activo. Iniciar con: docker compose up -d redis
    )

    findstr /i "mailpit" docker_status.log >nul 2>&1
    if !errorlevel! equ 0 (
        echo   OK: Mailpit activo
    ) else (
        echo   INFO: Mailpit no esta activo. Opcional: docker compose up -d mailpit
    )
) else (
    echo   AVISO: Docker daemon no disponible o docker compose fallo.
)
if exist docker_status.log del /f /q docker_status.log 2>nul

REM ----------------------------------------------------------
REM Frontend: Verificacion rapida de entorno
REM ----------------------------------------------------------
if exist frontend\package.json (
    if not exist frontend\node_modules (
        echo   INFO FRONTEND: node_modules ausente. Instalar con: npm --prefix frontend install
    )
)

echo ========================================================
if !HAS_ERR! equ 0 (
    echo   RESULTADO: VERIFICACION EXITOSA - LISTO PARA COMMIT
) else (
    echo   RESULTADO: FALLO DETECTADO - REVISA LOS ERRORES
)
echo ========================================================
exit /b !HAS_ERR!

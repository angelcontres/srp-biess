#!/usr/bin/env bash
# ========================================================
#   SRP-BIESS: ARNES DE VERIFICACION RAPIDA (TOKEN SAVER)
# ========================================================

set -o pipefail

echo "========================================================"
echo "  SRP-BIESS: ARNES DE VERIFICACION RAPIDA (TOKEN SAVER)"
echo "========================================================"

HAS_ERR=0
MODE_FULL=0

if [ "${1:-}" = "--full" ] || [ "${1:-}" = "--all" ]; then
    MODE_FULL=1
fi

# ----------------------------------------------------------
# [1/5] Verificación de Arquitectura Limpia (ADR-003)
# ----------------------------------------------------------
echo "[1/5] Verificando Arquitectura Limpia en internal/domain..."
if [ -d "internal/domain" ]; then
    FORBIDDEN_PATTERN="database/sql|github.com/jackc/pgx|github.com/gofiber/fiber|github.com/redis/go-redis|net/http|srp-biess/internal/infrastructure|srp-biess/internal/interfaces"
    OFFENDING_FILES=$(grep -rnE "$FORBIDDEN_PATTERN" internal/domain/ 2>/dev/null || true)
    if [ -n "$OFFENDING_FILES" ]; then
        echo "  [ERROR ARQUITECTURA]: Dependencias prohibidas encontradas en internal/domain/:"
        echo "$OFFENDING_FILES" | head -n 10
        HAS_ERR=1
    else
        echo "  [OK] Dominio puro y desacoplado - Sin SQL, HTTP ni frameworks"
    fi
else
    echo "  [AVISO] Directorio internal/domain no encontrado."
fi

# ----------------------------------------------------------
# [2/5] Compilación de Paquetes Go
# ----------------------------------------------------------
echo "[2/5] Verificando compilacion de Go..."
if [ -f "go.mod" ]; then
    if ! go build ./... 2> build_err.log; then
        echo "  [ERROR DE COMPILACION (./...)]:"
        grep -iE "error|undefined|cannot|syntax" build_err.log | head -n 15
        HAS_ERR=1
    elif ! go build -o /dev/null ./cmd/api 2> build_err.log; then
        echo "  [ERROR DE COMPILACION (cmd/api)]:"
        grep -iE "error|undefined|cannot|syntax" build_err.log | head -n 15
        HAS_ERR=1
    else
        echo "  [OK] Backend y Composition Root compilan sin errores"
    fi
    rm -f build_err.log
else
    echo "  [AVISO] go.mod no encontrado."
fi

# ----------------------------------------------------------
# [3/5] Análisis Estático (go vet)
# ----------------------------------------------------------
echo "[3/5] Ejecutando analisis estatico con go vet..."
if [ -f "go.mod" ]; then
    if ! go vet ./internal/... ./cmd/... 2> vet_err.log; then
        echo "  [AVISO O ERROR EN GO VET]:"
        head -n 15 vet_err.log
        HAS_ERR=1
    else
        echo "  [OK] Analisis estatico aprobado"
    fi
    rm -f vet_err.log
fi

# ----------------------------------------------------------
# [4/5] Pruebas Automatizadas
# ----------------------------------------------------------
if [ "$MODE_FULL" -eq 1 ]; then
    echo "[4/5] Ejecutando suite completa de pruebas: go test ./..."
    if ! go test ./... 2> test_err.log > test_out.log; then
        echo "  [ERROR EN PRUEBAS - SUITE COMPLETA]:"
        grep -iE "FAIL|---" test_out.log | head -n 15
        HAS_ERR=1
    else
        echo "  [OK] Todas las pruebas unitarias y de integracion pasaron exitosamente"
    fi
    rm -f test_err.log test_out.log
else
    echo "[4/5] Ejecutando pruebas rapidas de Dominio y Aplicacion..."
    if ! go test ./internal/domain/... ./internal/application/... 2> test_err.log > test_out.log; then
        echo "  [ERROR EN PRUEBAS DE DOMINIO/APLICACION]:"
        grep -iE "FAIL|---" test_out.log | head -n 15
        HAS_ERR=1
    else
        echo "  [OK] Pruebas rapidas de logica de negocio y validacion aprobadas"
    fi
    rm -f test_err.log test_out.log
fi

# ----------------------------------------------------------
# [5/5] Estado de Infraestructura Docker
# ----------------------------------------------------------
echo "[5/5] Verificando servicios de infraestructura en Docker..."
if docker compose ps > docker_status.log 2>&1; then
    if grep -iq "postgres" docker_status.log; then
        echo "  [OK] PostgreSQL activo"
    else
        echo "  [AVISO] PostgreSQL no esta activo. Iniciar con: docker compose up -d postgres"
    fi

    if grep -iq "redis" docker_status.log; then
        echo "  [OK] Redis activo"
    else
        echo "  [AVISO] Redis no esta activo. Iniciar con: docker compose up -d redis"
    fi

    if grep -iq "mailpit" docker_status.log; then
        echo "  [OK] Mailpit activo"
    else
        echo "  [INFO] Mailpit no esta activo. Opcional: docker compose up -d mailpit"
    fi
else
    echo "  [AVISO] Docker daemon no disponible o docker compose fallo."
fi
rm -f docker_status.log

# ----------------------------------------------------------
# Frontend: Verificación rápida de entorno
# ----------------------------------------------------------
if [ -f "frontend/package.json" ]; then
    if [ ! -d "frontend/node_modules" ]; then
        echo "  [INFO FRONTEND] node_modules ausente. Instalar con: npm --prefix frontend install"
    fi
fi

echo "========================================================"
if [ "$HAS_ERR" -eq 0 ]; then
    echo "  [RESULTADO]: VERIFICACION EXITOSA - LISTO PARA COMMIT"
else
    echo "  [RESULTADO]: FALLO DETECTADO - REVISA LOS ERRORES"
fi
echo "========================================================"
exit "$HAS_ERR"

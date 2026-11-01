#!/usr/bin/env bash

echo "========================================================"
echo "  SRP-BIESS: ARNES DE VERIFICACION RAPIDA (TOKEN SAVER)"
echo "========================================================"

HAS_ERROR=0

# 1. Arquitectura Limpia
echo "[1/3] Verificando Arquitectura Limpia en internal/domain..."
if [ -d "internal/domain" ]; then
    if grep -rn "database/sql" internal/domain/*.go > /dev/null 2>&1; then
        echo "[ERROR ARQUITECTURA]: Imports de database/sql en internal/domain/ (ADR-003)."
        HAS_ERROR=1
    fi
    if grep -rn "github.com/gofiber/fiber" internal/domain/*.go > /dev/null 2>&1; then
        echo "[ERROR ARQUITECTURA]: Imports de Fiber en internal/domain/ (ADR-003)."
        HAS_ERROR=1
    fi
fi
if [ $HAS_ERROR -eq 0 ]; then
    echo "[OK] Dominio puro y desacoplado."
fi

# 2. Compilación Go
echo "[2/3] Verificando compilacion de Go..."
if [ -f "go.mod" ]; then
    if ! go build ./... 2> build_err.log; then
        echo "[ERROR DE COMPILACION]:"
        grep -iE "error|undefined|cannot|syntax" build_err.log | head -n 15
        rm -f build_err.log
        HAS_ERROR=1
    else
        rm -f build_err.log
        echo "[OK] Backend compila perfectamente."
    fi
else
    echo "[INFO] go.mod aun no inicializado. Pendiente Ticket SRP-001."
fi

# 3. Servicios Docker
echo "[3/3] Verificando estado de contenedores Docker..."
if docker compose ps --services --filter "status=running" 2>/dev/null | grep -q "postgres"; then
    echo "[OK] PostgreSQL activo en Docker."
else
    echo "[AVISO] PostgreSQL no esta corriendo. Puedes levantarlo con: docker compose up -d"
fi

echo "========================================================"
if [ $HAS_ERROR -eq 0 ]; then
    echo "  [RESULTADO]: VERIFICACION EXITOSA - LISTO PARA COMMIT"
else
    echo "  [RESULTADO]: FALLO DETECTADO - COPIA EL ERROR A TU IA"
fi
echo "========================================================"
exit $HAS_ERROR

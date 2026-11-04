@echo off
echo =====================================================================
echo  SRP-BIESS ^| Aplicando Semilla Completa de Pruebas
echo =====================================================================
docker compose ps | findstr /i "postgres" >nul 2>&1
if errorlevel 1 (
    echo [ERROR] El contenedor de PostgreSQL no esta corriendo.
    echo Inicia Docker primero con start.bat
    exit /b 1
)

type db\seeds\semilla_completa.sql | docker exec -i srp-biess-postgres psql -U srp_user -d srp_biess
if errorlevel 1 (
    echo.
    echo [ERROR] Fallo al aplicar la semilla.
    exit /b 1
)

echo.
echo [OK] Semilla completa aplicada con exito.
echo.
echo Cuentas listas para Login (clave: BiessPostulante2026*):
echo   - 0926687856 ^| postulante.test@biess.gob.ec (Afiliada - Borrador listo)
echo   - 1710034065 ^| jubilado.test@biess.gob.ec   (Jubilado - Postulacion enviada)
echo   - 0912345675 ^| empleador.test@biess.gob.ec  (Empleador - Cuenta limpia)
echo   - 1304567892 ^| noelegible.test@biess.gob.ec (No Elegible)
echo.
echo Cuentas listas para probar Registro desde cero:
echo   - 0110500006 ^| Diego Cuenca ^| Exp 2022-04-10 ^| Dactilar A1234A5678 (Afiliado)
echo   - 1810500007 ^| Elena Morales ^| Exp 2019-10-15 ^| Dactilar T5678T1234 (Jubilada)
echo =====================================================================

-- Postulante de prueba SOLO para desarrollo local (SRP-013). No ejecutar en producción.
-- Credenciales: cédula 0926687856, contraseña BiessPostulante2026*
-- Uso: docker exec -i srp-biess-postgres psql -U srp_user -d srp_biess < db/seeds/postulante_prueba.sql
-- Idempotente: si la cédula ya existe, conserva la contraseña y solo completa los datos del
-- Registro Civil que estén vacíos.
-- Los datos verificados coinciden con el mock del Registro Civil (registro_civil_mock.go).
INSERT INTO postulante (cedula, nombres, apellidos, nacionalidad, genero, fecha_nacimiento, estado_civil, correo, password_hash)
VALUES (
  '0926687856',
  'MARIA JOSE',
  'PEREZ LOPEZ',
  'ECUATORIANA',
  'FEMENINO',
  '1985-03-21',
  'CASADO',
  'carlitoscharlie2005@gmail.com',
  '$argon2id$v=19$m=65536,t=1,p=2$cSUqFj2dE9xCIZOic6W+Mg$uy8ChFmsSiczO7utxsT+RSrnsOI1XS4C4x0mEw0kThQ'
)
ON CONFLICT (cedula) DO UPDATE
SET genero           = COALESCE(postulante.genero, EXCLUDED.genero),
    fecha_nacimiento = COALESCE(postulante.fecha_nacimiento, EXCLUDED.fecha_nacimiento),
    estado_civil     = COALESCE(postulante.estado_civil, EXCLUDED.estado_civil);

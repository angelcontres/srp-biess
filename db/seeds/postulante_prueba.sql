-- Postulante de prueba SOLO para desarrollo local (SRP-013). No ejecutar en producción.
-- Credenciales: cédula 0926687856, contraseña BiessPostulante2026*
-- Uso: docker exec -i srp-biess-postgres psql -U srp_user -d srp_biess < db/seeds/postulante_prueba.sql
-- Idempotente: si la cédula ya existe, no hace nada.
INSERT INTO postulante (cedula, nombres, apellidos, nacionalidad, correo, password_hash)
VALUES (
  '0926687856',
  'MARIA JOSE',
  'PEREZ LOPEZ',
  'ECUATORIANA',
  'postulante.test@biess.gob.ec',
  '$argon2id$v=19$m=65536,t=1,p=2$cSUqFj2dE9xCIZOic6W+Mg$uy8ChFmsSiczO7utxsT+RSrnsOI1XS4C4x0mEw0kThQ'
)
ON CONFLICT (cedula) DO NOTHING;

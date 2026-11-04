-- =====================================================================
-- SRP-BIESS · Semilla Completa de Pruebas y Desarrollo Local
-- =====================================================================
-- Este script es IDEMPOTENTE y configura todos los escenarios del sistema:
--
-- 1. USUARIOS CON CUENTA LISTA PARA INICIAR SESIÓN (Contraseña: BiessPostulante2026*)
--    - Cédula: 0926687856 | Correo: postulante.test@biess.gob.ec (Afiliada Guayas - Borrador en curso)
--    - Cédula: 1710034065 | Correo: jubilado.test@biess.gob.ec   (Jubilado Pichincha - Postulación ENVIADA)
--    - Cédula: 0912345675 | Correo: empleador.test@biess.gob.ec  (Empleador Guayas - Cuenta limpia)
--    - Cédula: 1304567892 | Correo: noelegible.test@biess.gob.ec (No Elegible Manabí - Sin calidad IESS)
--
-- 2. USUARIOS DISPONIBLES EN REGISTRO CIVIL PARA PROBAR REGISTRO DESDE CERO
--    (NO están en la base de datos para que puedas probar /registro completo):
--    - Cédula: 0110500006 (Azuay - Diego Fernando Cuenca Cabrera)
--      Fecha exp: 2022-04-10 | Código dactilar: A1234A5678 | Calidad: Afiliado
--    - Cédula: 1810500007 (Tungurahua - Elena Patricia Morales Vaca)
--      Fecha exp: 2019-10-15 | Código dactilar: T5678T1234 | Calidad: Jubilado
--
-- 3. PRUEBA DE RESILIENCIA (SERVICIO CAÍDO):
--    - Cédula: 0920000007 (Simula caída de servicio del Registro Civil / IESS - 503)
-- =====================================================================

-- ---------------------------------------------------------------------
-- 1. Sembrado de Postulantes
-- ---------------------------------------------------------------------
INSERT INTO postulante (id, cedula, nombres, apellidos, nacionalidad, genero, fecha_nacimiento, estado_civil, correo, password_hash, estado)
VALUES
  (
    '09266878-0000-4000-8000-000000000001',
    '0926687856',
    'MARIA JOSE',
    'PEREZ LOPEZ',
    'ECUATORIANA',
    'FEMENINO',
    '1985-03-21',
    'CASADO',
    'postulante.test@biess.gob.ec',
    '$argon2id$v=19$m=65536,t=1,p=2$cSUqFj2dE9xCIZOic6W+Mg$uy8ChFmsSiczO7utxsT+RSrnsOI1XS4C4x0mEw0kThQ',
    'ACTIVO'
  ),
  (
    '17100340-0000-4000-8000-000000000002',
    '1710034065',
    'LUIS ALBERTO',
    'TORRES VERA',
    'ECUATORIANA',
    'MASCULINO',
    '1958-07-09',
    'CASADO',
    'jubilado.test@biess.gob.ec',
    '$argon2id$v=19$m=65536,t=1,p=2$cSUqFj2dE9xCIZOic6W+Mg$uy8ChFmsSiczO7utxsT+RSrnsOI1XS4C4x0mEw0kThQ',
    'ACTIVO'
  ),
  (
    '09123456-0000-4000-8000-000000000003',
    '0912345675',
    'CARLOS ANDRES',
    'MORA SALAZAR',
    'ECUATORIANA',
    'MASCULINO',
    '1975-12-01',
    'SOLTERO',
    'empleador.test@biess.gob.ec',
    '$argon2id$v=19$m=65536,t=1,p=2$cSUqFj2dE9xCIZOic6W+Mg$uy8ChFmsSiczO7utxsT+RSrnsOI1XS4C4x0mEw0kThQ',
    'ACTIVO'
  ),
  (
    '13045678-0000-4000-8000-000000000004',
    '1304567892',
    'ANA LUCIA',
    'GOMEZ RUIZ',
    'COLOMBIANA',
    'FEMENINO',
    '1990-04-15',
    'SOLTERO',
    'noelegible.test@biess.gob.ec',
    '$argon2id$v=19$m=65536,t=1,p=2$cSUqFj2dE9xCIZOic6W+Mg$uy8ChFmsSiczO7utxsT+RSrnsOI1XS4C4x0mEw0kThQ',
    'ACTIVO'
  )
ON CONFLICT (cedula) DO UPDATE
SET nombres          = EXCLUDED.nombres,
    apellidos        = EXCLUDED.apellidos,
    genero           = EXCLUDED.genero,
    fecha_nacimiento = EXCLUDED.fecha_nacimiento,
    estado_civil     = EXCLUDED.estado_civil,
    nacionalidad     = EXCLUDED.nacionalidad,
    correo           = EXCLUDED.correo;

-- ---------------------------------------------------------------------
-- 2. Postulación en BORRADOR (0926687856 - María José Pérez López)
-- ---------------------------------------------------------------------
INSERT INTO postulacion (
  id, postulante_id, convocatoria_id, cargo_id, estado,
  elegibilidad_estado, requiere_revision, version, created_at, updated_at
)
VALUES (
  'a1111111-1111-4111-8111-111111111111',
  (SELECT id FROM postulante WHERE cedula = '0926687856'),
  1,
  1, -- Sector Afiliados
  'BORRADOR',
  'ELEGIBLE',
  false,
  1,
  now() - interval '2 days',
  now()
)
ON CONFLICT (postulante_id, convocatoria_id) DO UPDATE
SET cargo_id = EXCLUDED.cargo_id,
    elegibilidad_estado = EXCLUDED.elegibilidad_estado;

-- Secciones pre-llenadas para el borrador
INSERT INTO postulacion_seccion (postulacion_id, seccion, datos, completa)
VALUES
  (
    'a1111111-1111-4111-8111-111111111111',
    'personal',
    '{"estadoCivil":"CASADO","parejaNombres":"ANA PEREZ","parejaCedula":"1710034065"}'::jsonb,
    true
  ),
  (
    'a1111111-1111-4111-8111-111111111111',
    'domicilio',
    '{"paisId":"1","provinciaId":"9","cantonId":"901","provinciaManual":"","ciudadManual":"","telefonoCelular":"0991234567","correoAlterno":"postulante.alt@correo.ec","callePrincipal":"Av. 9 de Octubre","numero":"N12-34","calleSecundaria":"Malecón Simón Bolívar","sector":"Centro","referencia":"Frente al parque Centenario","telefonoDomicilio":"042123456"}'::jsonb,
    true
  ),
  (
    'a1111111-1111-4111-8111-111111111111',
    'familiares',
    '{"tieneFamiliares":"NO","familiares":[]}'::jsonb,
    true
  ),
  (
    'a1111111-1111-4111-8111-111111111111',
    'cargo',
    '{"tieneRespaldo":"NO"}'::jsonb,
    true
  ),
  (
    'a1111111-1111-4111-8111-111111111111',
    'formacion',
    '{"formaciones":[{"id":"f0000000-0000-4000-8000-000000000001","nivel":"TERCER_NIVEL","institucion":"ESPOL","tipoInstitucion":"NACIONAL","titulo":"Economista","area":"Economía y Finanzas","numeroSenescyt":"1006-2015-1234","fechaSenescyt":"2015-04-15","documento":null}]}'::jsonb,
    true
  ),
  (
    'a1111111-1111-4111-8111-111111111111',
    'experiencia',
    '{"experiencias":[{"id":"e0000000-0000-4000-8000-000000000001","tipo":"DIRECCION","cargo":"Gerente de Inversiones","institucion":"Banco de Guayaquil","fechaInicio":"2018-01-15","actualmente":true,"fechaFin":"","actividades":"Gestión del portafolio de activos previsionales","funciones":"Dirección ejecutiva y gestión de riesgos","documento":null}]}'::jsonb,
    true
  ),
  (
    'a1111111-1111-4111-8111-111111111111',
    'publicaciones',
    '{"tienePublicaciones":"NO","publicaciones":[]}'::jsonb,
    true
  ),
  (
    'a1111111-1111-4111-8111-111111111111',
    'capacitacion',
    '{"capacitaciones":[{"id":"c0000000-0000-4000-8000-000000000001","tipoEvento":"CURSO","nombre":"Gobierno Corporativo y Mercado Bursátil","institucion":"Bolsa de Valores de Guayaquil","fechaInicio":"2023-03-01","fechaFin":"2023-03-15","horas":"40","documento":null}]}'::jsonb,
    true
  ),
  (
    'a1111111-1111-4111-8111-111111111111',
    'inhabilidades',
    '{"respuestas":["NO","NO","NO","NO","NO","NO","NO","NO","NO","NO","NO","NO","NO"]}'::jsonb,
    true
  )
ON CONFLICT (postulacion_id, seccion) DO UPDATE
SET datos    = EXCLUDED.datos,
    completa = EXCLUDED.completa;

-- ---------------------------------------------------------------------
-- 3. Postulación ENVIADA (1710034065 - Luis Alberto Torres Vera)
-- ---------------------------------------------------------------------
INSERT INTO postulacion (
  id, codigo_unico, postulante_id, convocatoria_id, cargo_id, estado,
  fecha_creacion, fecha_guardado, fecha_envio,
  elegibilidad_estado, requiere_revision, version, created_at, updated_at
)
VALUES (
  'b2222222-2222-4222-8222-222222222222',
  'BIESS-2026-000001',
  (SELECT id FROM postulante WHERE cedula = '1710034065'),
  1,
  2, -- Sector Jubilados
  'ENVIADA',
  now() - interval '3 days',
  now() - interval '1 day',
  now() - interval '1 day',
  'ELEGIBLE',
  false,
  2,
  now() - interval '3 days',
  now() - interval '1 day'
)
ON CONFLICT (postulante_id, convocatoria_id) DO UPDATE
SET estado       = EXCLUDED.estado,
    codigo_unico = EXCLUDED.codigo_unico,
    fecha_envio  = EXCLUDED.fecha_envio;

-- Secciones para la postulación enviada
INSERT INTO postulacion_seccion (postulacion_id, seccion, datos, completa)
VALUES
  (
    'b2222222-2222-4222-8222-222222222222',
    'personal',
    '{"estadoCivil":"CASADO","parejaNombres":"MARTHA VERA","parejaCedula":"0912345675"}'::jsonb,
    true
  ),
  (
    'b2222222-2222-4222-8222-222222222222',
    'domicilio',
    '{"paisId":"1","provinciaId":"17","cantonId":"1701","provinciaManual":"","ciudadManual":"","telefonoCelular":"0987654321","correoAlterno":"","callePrincipal":"Av. Amazonas","numero":"N34-120","calleSecundaria":"Av. República","sector":"Iñaquito","referencia":"Cerca del Parque La Carolina","telefonoDomicilio":"022987654"}'::jsonb,
    true
  ),
  (
    'b2222222-2222-4222-8222-222222222222',
    'familiares',
    '{"tieneFamiliares":"NO","familiares":[]}'::jsonb,
    true
  ),
  (
    'b2222222-2222-4222-8222-222222222222',
    'cargo',
    '{"tieneRespaldo":"NO"}'::jsonb,
    true
  ),
  (
    'b2222222-2222-4222-8222-222222222222',
    'formacion',
    '{"formaciones":[{"id":"f0000000-0000-4000-8000-000000000002","nivel":"MAESTRIA","institucion":"Universidad Andina Simón Bolívar","tipoInstitucion":"NACIONAL","titulo":"Magíster en Derecho Financiero","area":"Derecho y Finanzas","numeroSenescyt":"1005-2010-5678","fechaSenescyt":"2010-09-20","documento":null}]}'::jsonb,
    true
  ),
  (
    'b2222222-2222-4222-8222-222222222222',
    'experiencia',
    '{"experiencias":[{"id":"e0000000-0000-4000-8000-000000000002","tipo":"DIRECCION","cargo":"Subdirector de Seguridad Social","institucion":"IESS","fechaInicio":"2005-06-01","actualmente":false,"fechaFin":"2020-12-31","actividades":"Dirección de pensiones y jubilaciones","funciones":"Gestión estratégica de la seguridad social","documento":null}]}'::jsonb,
    true
  ),
  (
    'b2222222-2222-4222-8222-222222222222',
    'publicaciones',
    '{"tienePublicaciones":"SI","publicaciones":[{"id":"p0000000-0000-4000-8000-000000000001","titulo":"Sostenibilidad de los Fondos Previsionales en el Ecuador","tipo":"PUBLICACION","medio":"Revista Jurídica y Financiera","fecha":"2019-11-10","relacion":"Estudio previsional IESS","documento":null}]}'::jsonb,
    true
  ),
  (
    'b2222222-2222-4222-8222-222222222222',
    'capacitacion',
    '{"capacitaciones":[{"id":"c0000000-0000-4000-8000-000000000002","tipoEvento":"SEMINARIO","nombre":"Régimen Previsional Latinoamericano","institucion":"OISS","fechaInicio":"2021-05-10","fechaFin":"2021-05-20","horas":"60","documento":null}]}'::jsonb,
    true
  ),
  (
    'b2222222-2222-4222-8222-222222222222',
    'inhabilidades',
    '{"respuestas":["NO","NO","NO","NO","NO","NO","NO","NO","NO","NO","NO","NO","NO"]}'::jsonb,
    true
  )
ON CONFLICT (postulacion_id, seccion) DO UPDATE
SET datos    = EXCLUDED.datos,
    completa = EXCLUDED.completa;

-- Auditoría de eventos para la postulación enviada
INSERT INTO auditoria_evento (actor_id, actor_tipo, accion, entidad, entidad_id, datos_despues, created_at)
VALUES
  (
    (SELECT id FROM postulante WHERE cedula = '1710034065'),
    'POSTULANTE',
    'POSTULACION_CREADA',
    'postulacion',
    'b2222222-2222-4222-8222-222222222222',
    '{"estado":"BORRADOR"}'::jsonb,
    now() - interval '3 days'
  ),
  (
    (SELECT id FROM postulante WHERE cedula = '1710034065'),
    'POSTULANTE',
    'CARGO_CONFIRMADO',
    'postulacion',
    'b2222222-2222-4222-8222-222222222222',
    '{"cargo_id":2,"sector":"JUBILADOS","elegibilidad":"ELEGIBLE"}'::jsonb,
    now() - interval '2 days'
  ),
  (
    (SELECT id FROM postulante WHERE cedula = '1710034065'),
    'POSTULANTE',
    'POSTULACION_ENVIADA',
    'postulacion',
    'b2222222-2222-4222-8222-222222222222',
    '{"codigo":"BIESS-2026-000001","estado":"ENVIADA"}'::jsonb,
    now() - interval '1 day'
  );

-- =====================================================================
-- SRP-BIESS · Base de Datos PostgreSQL 16
-- Concurso de Méritos y Oposición Directorio BIESS (EF-11 y EF-12)
-- Script Inicial DDL + Semillas (Idempotente)
-- =====================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ---------------------------------------------------------------------
-- 1. Tipos Enumerados
-- ---------------------------------------------------------------------
DO $$ BEGIN
  CREATE TYPE estado_postulante AS ENUM ('REGISTRADO', 'ACTIVO', 'BLOQUEADO', 'INACTIVO');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE estado_postulacion AS ENUM ('BORRADOR', 'LISTA_PARA_ENVIO', 'ENVIADA', 'EN_REVISION', 'NO_ELEGIBLE', 'CERRADA');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE genero AS ENUM ('MASCULINO', 'FEMENINO', 'OTRO', 'PREFIERO_NO_DECIR');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE estado_civil AS ENUM ('SOLTERO', 'CASADO', 'UNION_DE_HECHO', 'DIVORCIADO', 'VIUDO', 'OTRO');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE sector_cargo AS ENUM ('AFILIADOS', 'JUBILADOS', 'EMPLEADORES');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE nivel_instruccion AS ENUM ('TERCER_NIVEL', 'ESPECIALIDAD', 'MAESTRIA', 'PHD');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE tipo_institucion AS ENUM ('NACIONAL', 'EXTRANJERA');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE tipo_experiencia AS ENUM ('GENERAL', 'DIRECCION');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE tipo_publicacion AS ENUM ('PUBLICACION', 'TRABAJO_INVESTIGACION');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE tipo_evento AS ENUM ('CURSO', 'SEMINARIO', 'TALLER');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE tipo_documento AS ENUM ('RESPALDO_ORGANIZACIONAL', 'TITULO', 'EXPERIENCIA', 'PUBLICACION', 'CAPACITACION', 'FORMULARIO_POSTULACION');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE estado_documento AS ENUM ('PENDIENTE', 'VALIDADO', 'RECHAZADO');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE estado_notificacion AS ENUM ('PENDIENTE', 'ENVIADA', 'FALLIDA');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE efecto_inhabilidad AS ENUM ('BLOQUEANTE', 'ALERTA', 'INFORMATIVO');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE institucion_familiar AS ENUM ('IESS', 'BIESS');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE parentesco_familiar AS ENUM (
    'CONVIVIENTE', 'CONYUGE', 'CUNADO', 'HERMANO', 'HIJO_POLITICO',
    'HIJO', 'MADRE', 'NIETO', 'NUERA', 'PADRE', 'PRIMO_HERMANO',
    'SOBRINO', 'SUEGRO', 'TIO', 'YERNO'
  );
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE tipo_validacion AS ENUM ('FECHA_EXPEDICION', 'CODIGO_DACTILAR');
EXCEPTION WHEN duplicate_object THEN null; END $$;

-- ---------------------------------------------------------------------
-- 2. Catálogos Geográficos
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS pais (
  id           smallserial PRIMARY KEY,
  codigo_iso2  char(2) NOT NULL UNIQUE,
  codigo_iso3  char(3) NOT NULL UNIQUE,
  nombre       varchar(100) NOT NULL,
  activo       boolean NOT NULL DEFAULT true,
  created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS provincia (
  id           smallserial PRIMARY KEY,
  pais_id      smallint NOT NULL REFERENCES pais(id) ON DELETE RESTRICT,
  codigo       varchar(10) NOT NULL,
  nombre       varchar(100) NOT NULL,
  activo       boolean NOT NULL DEFAULT true,
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_provincia_pais_codigo UNIQUE (pais_id, codigo)
);

CREATE TABLE IF NOT EXISTS canton (
  id           smallserial PRIMARY KEY,
  provincia_id smallint NOT NULL REFERENCES provincia(id) ON DELETE RESTRICT,
  codigo       varchar(10) NOT NULL,
  nombre       varchar(100) NOT NULL,
  activo       boolean NOT NULL DEFAULT true,
  created_at   timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_canton_provincia_codigo UNIQUE (provincia_id, codigo)
);

-- ---------------------------------------------------------------------
-- 3. Cargos y Convocatoria
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cargo (
  id          smallserial PRIMARY KEY,
  codigo      varchar(10) NOT NULL UNIQUE,
  nombre      varchar(200) NOT NULL,
  sector      sector_cargo NOT NULL,
  activo      boolean NOT NULL DEFAULT true,
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS convocatoria (
  id                          smallserial PRIMARY KEY,
  codigo                      varchar(30) NOT NULL UNIQUE,
  nombre                      varchar(200) NOT NULL,
  fecha_inicio_postulacion    date NOT NULL,
  fecha_fin_postulacion       date NOT NULL,
  fase_postulacion_habilitada boolean NOT NULL DEFAULT false,
  version_formulario          varchar(20) NOT NULL DEFAULT '1.0',
  created_at                  timestamptz NOT NULL DEFAULT now(),
  updated_at                  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_convocatoria_fechas CHECK (fecha_fin_postulacion >= fecha_inicio_postulacion)
);

-- ---------------------------------------------------------------------
-- 4. Postulante y Autenticación (EF-11)
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS postulante (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  cedula              char(10) NOT NULL UNIQUE,
  nombres             varchar(120) NOT NULL,
  apellidos           varchar(120) NOT NULL,
  genero              genero,
  fecha_nacimiento    date,
  nacionalidad        varchar(60) NOT NULL,
  estado_civil        estado_civil,
  correo              varchar(254) NOT NULL UNIQUE,
  password_hash       varchar(255) NOT NULL,
  estado              estado_postulante NOT NULL DEFAULT 'ACTIVO',
  intentos_fallidos   smallint NOT NULL DEFAULT 0,
  bloqueado_hasta     timestamptz,
  ultimo_acceso       timestamptz,
  primer_ingreso      boolean NOT NULL DEFAULT true,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_postulante_cedula_len CHECK (length(cedula) = 10)
);

CREATE TABLE IF NOT EXISTS pareja_postulante (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulante_id     uuid NOT NULL UNIQUE REFERENCES postulante(id) ON DELETE CASCADE,
  nombres_apellidos varchar(200) NOT NULL,
  cedula            char(10) NOT NULL,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_pareja_cedula_len CHECK (length(cedula) = 10)
);

CREATE TABLE IF NOT EXISTS domicilio_postulante (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulante_id      uuid NOT NULL UNIQUE REFERENCES postulante(id) ON DELETE CASCADE,
  pais_id            smallint NOT NULL REFERENCES pais(id) ON DELETE RESTRICT,
  provincia_id       smallint REFERENCES provincia(id) ON DELETE RESTRICT,
  canton_id          smallint REFERENCES canton(id) ON DELETE RESTRICT,
  provincia_manual   varchar(100),
  ciudad_manual      varchar(100),
  telefono_celular   varchar(20) NOT NULL,
  correo_alterno     varchar(254),
  calle_principal    varchar(150) NOT NULL,
  numero             varchar(30) NOT NULL,
  calle_secundaria   varchar(150) NOT NULL,
  sector             varchar(100) NOT NULL,
  referencia         varchar(200) NOT NULL,
  telefono_domicilio varchar(20),
  created_at         timestamptz NOT NULL DEFAULT now(),
  updated_at         timestamptz NOT NULL DEFAULT now(),
  -- 52 = Ecuador en la semilla de países (sección 9).
  CONSTRAINT ck_domicilio_pais_ec CHECK (
    (pais_id = 52 AND provincia_id IS NOT NULL AND canton_id IS NOT NULL AND provincia_manual IS NULL AND ciudad_manual IS NULL)
    OR (pais_id <> 52 AND provincia_manual IS NOT NULL AND ciudad_manual IS NOT NULL)
  )
);

CREATE TABLE IF NOT EXISTS familiar_institucion (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulante_id     uuid NOT NULL REFERENCES postulante(id) ON DELETE CASCADE,
  nombres_apellidos varchar(200) NOT NULL,
  parentesco        parentesco_familiar NOT NULL,
  institucion       institucion_familiar NOT NULL,
  area              varchar(150) NOT NULL,
  cargo             varchar(150) NOT NULL,
  created_at        timestamptz NOT NULL DEFAULT now(),
  updated_at        timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------
-- 5. Postulación y Documentos (EF-12)
-- ---------------------------------------------------------------------
CREATE SEQUENCE IF NOT EXISTS seq_codigo_postulacion START WITH 1;

CREATE TABLE IF NOT EXISTS postulacion (
  id                            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  codigo_unico                  varchar(30) UNIQUE,
  postulante_id                 uuid NOT NULL REFERENCES postulante(id) ON DELETE RESTRICT,
  convocatoria_id               smallint NOT NULL REFERENCES convocatoria(id) ON DELETE RESTRICT,
  cargo_id                      smallint REFERENCES cargo(id) ON DELETE RESTRICT,
  estado                        estado_postulacion NOT NULL DEFAULT 'BORRADOR',
  fecha_creacion                timestamptz NOT NULL DEFAULT now(),
  fecha_guardado                timestamptz,
  fecha_envio                   timestamptz,
  elegibilidad_estado           varchar(20) NOT NULL DEFAULT 'PENDIENTE',
  elegibilidad_motivo           text,
  requiere_revision             boolean NOT NULL DEFAULT false,
  tiene_familiares_institucion  boolean,
  tiene_publicaciones           boolean,
  ultima_seccion                varchar(40),
  version                       integer NOT NULL DEFAULT 1,
  created_at                    timestamptz NOT NULL DEFAULT now(),
  updated_at                    timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_postulacion_vigente UNIQUE (postulante_id, convocatoria_id)
);

-- SRP-013: integración mínima del formulario de postulación (SDD 7.3.4).
-- Almacena cada sección del formulario como JSONB mientras se normaliza en las tablas
-- de SDD 4.6 (domicilio_postulante, familiar_institucion, formacion_academica, ...),
-- pendiente de SRP-005. Idempotente: se puede ejecutar sobre una base existente.
CREATE TABLE IF NOT EXISTS postulacion_seccion (
  postulacion_id  uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  seccion         varchar(40) NOT NULL,
  datos           jsonb NOT NULL,
  completa        boolean NOT NULL DEFAULT false,
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (postulacion_id, seccion),
  CONSTRAINT ck_postulacion_seccion_nombre CHECK (seccion IN (
    'personal', 'domicilio', 'familiares', 'cargo', 'formacion',
    'experiencia', 'publicaciones', 'capacitacion', 'inhabilidades'
  ))
);

CREATE TABLE IF NOT EXISTS documento (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulante_id      uuid NOT NULL REFERENCES postulante(id) ON DELETE CASCADE,
  postulacion_id     uuid REFERENCES postulacion(id) ON DELETE CASCADE,
  tipo               tipo_documento NOT NULL,
  nombre_original    varchar(255) NOT NULL,
  nombre_almacenado  varchar(255) NOT NULL UNIQUE,
  mime_type          varchar(100) NOT NULL,
  tamano_bytes       bigint NOT NULL,
  hash_sha256        char(64) NOT NULL,
  ruta               text NOT NULL,
  estado             estado_documento NOT NULL DEFAULT 'PENDIENTE',
  created_at         timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_documento_tamano CHECK (tamano_bytes > 0 AND tamano_bytes <= 4194304)
);

CREATE TABLE IF NOT EXISTS respaldo_organizacional (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id      uuid NOT NULL UNIQUE REFERENCES postulacion(id) ON DELETE CASCADE,
  tiene_respaldo      boolean NOT NULL,
  nombre_organizacion varchar(200),
  documento_id        uuid REFERENCES documento(id) ON DELETE SET NULL,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS formacion_academica (
  id                       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id           uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  nivel_instruccion        nivel_instruccion NOT NULL,
  institucion_educativa    varchar(200) NOT NULL,
  tipo_institucion         tipo_institucion NOT NULL,
  titulo_obtenido          varchar(200) NOT NULL,
  area_estudios            varchar(200) NOT NULL,
  numero_registro_senescyt varchar(50),
  fecha_registro_senescyt  date,
  documento_id             uuid REFERENCES documento(id) ON DELETE SET NULL,
  created_at               timestamptz NOT NULL DEFAULT now(),
  updated_at               timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS experiencia_profesional (
  id                       uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id           uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  tipo_experiencia         tipo_experiencia NOT NULL,
  cargo_puesto             varchar(200) NOT NULL,
  institucion_empresa      varchar(200) NOT NULL,
  fecha_inicio             date NOT NULL,
  fecha_fin                date,
  actualmente              boolean NOT NULL DEFAULT false,
  actividades_relacionadas text,
  descripcion_funciones    varchar(1000) NOT NULL,
  tiempo_anios             smallint NOT NULL DEFAULT 0,
  tiempo_meses             smallint NOT NULL DEFAULT 0,
  tiempo_dias              smallint NOT NULL DEFAULT 0,
  documento_id             uuid REFERENCES documento(id) ON DELETE SET NULL,
  created_at               timestamptz NOT NULL DEFAULT now(),
  updated_at               timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS publicacion (
  id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id     uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  titulo             varchar(250) NOT NULL,
  tipo               tipo_publicacion NOT NULL,
  medio_publicacion  varchar(200) NOT NULL,
  fecha_publicacion  date NOT NULL,
  relacion_formacion boolean NOT NULL,
  documento_id       uuid REFERENCES documento(id) ON DELETE SET NULL,
  created_at         timestamptz NOT NULL DEFAULT now(),
  updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS capacitacion (
  id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  tipo_evento    tipo_evento NOT NULL,
  nombre_evento  varchar(200) NOT NULL,
  institucion    varchar(200) NOT NULL,
  fecha_inicio   date NOT NULL,
  fecha_fin      date NOT NULL,
  numero_horas   smallint NOT NULL,
  documento_id   uuid REFERENCES documento(id) ON DELETE SET NULL,
  created_at     timestamptz NOT NULL DEFAULT now(),
  updated_at     timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------
-- 6. Inhabilidades y Consentimiento
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS inhabilidad_pregunta (
  id                   smallserial PRIMARY KEY,
  codigo               varchar(20) NOT NULL UNIQUE,
  texto                text NOT NULL,
  orden                smallint NOT NULL UNIQUE,
  efecto               efecto_inhabilidad NOT NULL,
  respuesta_bloqueante boolean,
  activo               boolean NOT NULL DEFAULT true,
  created_at           timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS inhabilidad_respuesta (
  id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id       uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  pregunta_id          smallint NOT NULL REFERENCES inhabilidad_pregunta(id) ON DELETE RESTRICT,
  respuesta            boolean NOT NULL,
  efecto_calculado     efecto_inhabilidad NOT NULL,
  institucion_remocion varchar(200),
  cargo_remocion       varchar(200),
  fecha_fin_remocion   date,
  motivo_remocion      text,
  created_at           timestamptz NOT NULL DEFAULT now(),
  updated_at           timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_inhabilidad_respuesta UNIQUE (postulacion_id, pregunta_id)
);

CREATE TABLE IF NOT EXISTS consentimiento (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id   uuid NOT NULL UNIQUE REFERENCES postulacion(id) ON DELETE CASCADE,
  version_terminos varchar(20) NOT NULL,
  aceptado         boolean NOT NULL CHECK (aceptado = true),
  textos_aceptados jsonb NOT NULL DEFAULT '{}'::jsonb,
  fecha_aceptacion timestamptz NOT NULL DEFAULT now(),
  ip               inet,
  user_agent       text
);

-- ---------------------------------------------------------------------
-- 7. Notificaciones (Outbox) y Auditoría
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notificacion (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id  uuid REFERENCES postulacion(id) ON DELETE SET NULL,
  postulante_id   uuid REFERENCES postulante(id) ON DELETE SET NULL,
  tipo            varchar(40) NOT NULL,
  destinatario    varchar(254) NOT NULL,
  asunto          varchar(200) NOT NULL,
  plantilla       varchar(80) NOT NULL,
  payload         jsonb NOT NULL DEFAULT '{}'::jsonb,
  estado          estado_notificacion NOT NULL DEFAULT 'PENDIENTE',
  intentos        smallint NOT NULL DEFAULT 0,
  fecha_creacion  timestamptz NOT NULL DEFAULT now(),
  proximo_intento timestamptz NOT NULL DEFAULT now(),
  fecha_envio     timestamptz,
  error           text
);

CREATE TABLE IF NOT EXISTS auditoria_evento (
  id            bigserial PRIMARY KEY,
  actor_id      uuid,
  actor_tipo    varchar(20) NOT NULL,
  accion        varchar(80) NOT NULL,
  entidad       varchar(80) NOT NULL,
  entidad_id    varchar(64),
  datos_antes   jsonb,
  datos_despues jsonb,
  ip            inet,
  user_agent    text,
  trace_id      uuid,
  created_at    timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------
-- 8. Datos Semilla (Seeds) Idempotentes
-- ---------------------------------------------------------------------

-- Países
INSERT INTO pais (id, codigo_iso2, codigo_iso3, nombre) VALUES
  (1,  'AF', 'AFG', 'Afganistán'),
  (2,  'AL', 'ALB', 'Albania'),
  (3,  'DE', 'DEU', 'Alemania'),
  (4,  'AD', 'AND', 'Andorra'),
  (5,  'AO', 'AGO', 'Angola'),
  (6,  'AG', 'ATG', 'Antigua y Barbuda'),
  (7,  'SA', 'SAU', 'Arabia Saudita'),
  (8,  'DZ', 'DZA', 'Argelia'),
  (9,  'AR', 'ARG', 'Argentina'),
  (10, 'AM', 'ARM', 'Armenia'),
  (11, 'AU', 'AUS', 'Australia'),
  (12, 'AT', 'AUT', 'Austria'),
  (13, 'AZ', 'AZE', 'Azerbaiyán'),
  (14, 'BS', 'BHS', 'Bahamas'),
  (15, 'BD', 'BGD', 'Bangladés'),
  (16, 'BB', 'BRB', 'Barbados'),
  (17, 'BH', 'BHR', 'Baréin'),
  (18, 'BE', 'BEL', 'Bélgica'),
  (19, 'BZ', 'BLZ', 'Belice'),
  (20, 'BJ', 'BEN', 'Benín'),
  (21, 'BY', 'BLR', 'Bielorrusia'),
  (22, 'MM', 'MMR', 'Birmania'),
  (23, 'BO', 'BOL', 'Bolivia'),
  (24, 'BA', 'BIH', 'Bosnia y Herzegovina'),
  (25, 'BW', 'BWA', 'Botsuana'),
  (26, 'BR', 'BRA', 'Brasil'),
  (27, 'BN', 'BRN', 'Brunéi'),
  (28, 'BG', 'BGR', 'Bulgaria'),
  (29, 'BF', 'BFA', 'Burkina Faso'),
  (30, 'BI', 'BDI', 'Burundi'),
  (31, 'BT', 'BTN', 'Bután'),
  (32, 'CV', 'CPV', 'Cabo Verde'),
  (33, 'KH', 'KHM', 'Camboya'),
  (34, 'CM', 'CMR', 'Camerún'),
  (35, 'CA', 'CAN', 'Canadá'),
  (36, 'QA', 'QAT', 'Catar'),
  (37, 'TD', 'TCD', 'Chad'),
  (38, 'CL', 'CHL', 'Chile'),
  (39, 'CN', 'CHN', 'China'),
  (40, 'CY', 'CYP', 'Chipre'),
  (41, 'VA', 'VAT', 'Ciudad del Vaticano'),
  (42, 'CO', 'COL', 'Colombia'),
  (43, 'KM', 'COM', 'Comoras'),
  (44, 'KP', 'PRK', 'Corea del Norte'),
  (45, 'KR', 'KOR', 'Corea del Sur'),
  (46, 'CI', 'CIV', 'Costa de Marfil'),
  (47, 'CR', 'CRI', 'Costa Rica'),
  (48, 'HR', 'HRV', 'Croacia'),
  (49, 'CU', 'CUB', 'Cuba'),
  (50, 'DK', 'DNK', 'Dinamarca'),
  (51, 'DM', 'DMA', 'Dominica'),
  (52, 'EC', 'ECU', 'Ecuador'),
  (53, 'EG', 'EGY', 'Egipto'),
  (54, 'SV', 'SLV', 'El Salvador'),
  (55, 'AE', 'ARE', 'Emiratos Árabes Unidos'),
  (56, 'ER', 'ERI', 'Eritrea'),
  (57, 'SK', 'SVK', 'Eslovaquia'),
  (58, 'SI', 'SVN', 'Eslovenia'),
  (59, 'ES', 'ESP', 'España'),
  (60, 'US', 'USA', 'Estados Unidos'),
  (61, 'EE', 'EST', 'Estonia'),
  (62, 'ET', 'ETH', 'Etiopía'),
  (63, 'PH', 'PHL', 'Filipinas'),
  (64, 'FI', 'FIN', 'Finlandia'),
  (65, 'FJ', 'FJI', 'Fiyi'),
  (66, 'FR', 'FRA', 'Francia'),
  (67, 'GA', 'GAB', 'Gabón'),
  (68, 'GM', 'GMB', 'Gambia'),
  (69, 'GE', 'GEO', 'Georgia'),
  (70, 'GH', 'GHA', 'Ghana'),
  (71, 'GD', 'GRD', 'Granada'),
  (72, 'GR', 'GRC', 'Grecia'),
  (73, 'GT', 'GTM', 'Guatemala'),
  (74, 'GY', 'GUY', 'Guyana'),
  (75, 'GN', 'GIN', 'Guinea'),
  (76, 'GW', 'GNB', 'Guinea-Bisáu'),
  (77, 'GQ', 'GNQ', 'Guinea Ecuatorial'),
  (78, 'HT', 'HTI', 'Haití'),
  (79, 'HN', 'HND', 'Honduras'),
  (80, 'HU', 'HUN', 'Hungría'),
  (81, 'IN', 'IND', 'India'),
  (82, 'ID', 'IDN', 'Indonesia'),
  (83, 'IQ', 'IRQ', 'Irak'),
  (84, 'IR', 'IRN', 'Irán'),
  (85, 'IE', 'IRL', 'Irlanda'),
  (86, 'IS', 'ISL', 'Islandia'),
  (87, 'MH', 'MHL', 'Islas Marshall'),
  (88, 'SB', 'SLB', 'Islas Salomón'),
  (89, 'IL', 'ISR', 'Israel'),
  (90, 'IT', 'ITA', 'Italia'),
  (91, 'JM', 'JAM', 'Jamaica'),
  (92, 'JP', 'JPN', 'Japón'),
  (93, 'JO', 'JOR', 'Jordania'),
  (94, 'KZ', 'KAZ', 'Kazajistán'),
  (95, 'KE', 'KEN', 'Kenia'),
  (96, 'KG', 'KGZ', 'Kirguistán'),
  (97, 'KI', 'KIR', 'Kiribati'),
  (98, 'KW', 'KWT', 'Kuwait'),
  (99, 'LA', 'LAO', 'Laos'),
  (100, 'LS', 'LSO', 'Lesoto'),
  (101, 'LV', 'LVA', 'Letonia'),
  (102, 'LB', 'LBN', 'Líbano'),
  (103, 'LR', 'LBR', 'Liberia'),
  (104, 'LY', 'LBY', 'Libia'),
  (105, 'LI', 'LIE', 'Liechtenstein'),
  (106, 'LT', 'LTU', 'Lituania'),
  (107, 'LU', 'LUX', 'Luxemburgo'),
  (108, 'MK', 'MKD', 'Macedonia del Norte'),
  (109, 'MG', 'MDG', 'Madagascar'),
  (110, 'MY', 'MYS', 'Malasia'),
  (111, 'MW', 'MWI', 'Malaui'),
  (112, 'MV', 'MDV', 'Maldivas'),
  (113, 'ML', 'MLI', 'Malí'),
  (114, 'MT', 'MLT', 'Malta'),
  (115, 'MA', 'MAR', 'Marruecos'),
  (116, 'MU', 'MUS', 'Mauricio'),
  (117, 'MR', 'MRT', 'Mauritania'),
  (118, 'MX', 'MEX', 'México'),
  (119, 'FM', 'FSM', 'Micronesia'),
  (120, 'MD', 'MDA', 'Moldavia'),
  (121, 'MC', 'MCO', 'Mónaco'),
  (122, 'MN', 'MNG', 'Mongolia'),
  (123, 'ME', 'MNE', 'Montenegro'),
  (124, 'MZ', 'MOZ', 'Mozambique'),
  (125, 'NA', 'NAM', 'Namibia'),
  (126, 'NR', 'NRU', 'Nauru'),
  (127, 'NP', 'NPL', 'Nepal'),
  (128, 'NI', 'NIC', 'Nicaragua'),
  (129, 'NE', 'NER', 'Níger'),
  (130, 'NG', 'NGA', 'Nigeria'),
  (131, 'NO', 'NOR', 'Noruega'),
  (132, 'NZ', 'NZL', 'Nueva Zelanda'),
  (133, 'OM', 'OMN', 'Omán'),
  (134, 'NL', 'NLD', 'Países Bajos'),
  (135, 'PK', 'PAK', 'Pakistán'),
  (136, 'PW', 'PLW', 'Palaos'),
  (137, 'PA', 'PAN', 'Panamá'),
  (138, 'PG', 'PNG', 'Papúa Nueva Guinea'),
  (139, 'PY', 'PRY', 'Paraguay'),
  (140, 'PE', 'PER', 'Perú'),
  (141, 'PL', 'POL', 'Polonia'),
  (142, 'PT', 'PRT', 'Portugal'),
  (143, 'GB', 'GBR', 'Reino Unido'),
  (144, 'CF', 'CAF', 'República Centroafricana'),
  (145, 'CZ', 'CZE', 'República Checa'),
  (146, 'CG', 'COG', 'República del Congo'),
  (147, 'CD', 'COD', 'República Democrática del Congo'),
  (148, 'DO', 'DOM', 'República Dominicana'),
  (149, 'RW', 'RWA', 'Ruanda'),
  (150, 'RO', 'ROU', 'Rumanía'),
  (151, 'RU', 'RUS', 'Rusia'),
  (152, 'WS', 'WSM', 'Samoa'),
  (153, 'KN', 'KNA', 'San Cristóbal y Nieves'),
  (154, 'SM', 'SMR', 'San Marino'),
  (155, 'VC', 'VCT', 'San Vicente y las Granadinas'),
  (156, 'LC', 'LCA', 'Santa Lucía'),
  (157, 'ST', 'STP', 'Santo Tomé y Príncipe'),
  (158, 'SN', 'SEN', 'Senegal'),
  (159, 'RS', 'SRB', 'Serbia'),
  (160, 'SC', 'SYC', 'Seychelles'),
  (161, 'SL', 'SLE', 'Sierra Leona'),
  (162, 'SG', 'SGP', 'Singapur'),
  (163, 'SY', 'SYR', 'Siria'),
  (164, 'SO', 'SOM', 'Somalia'),
  (165, 'LK', 'LKA', 'Sri Lanka'),
  (166, 'SZ', 'SWZ', 'Suazilandia'),
  (167, 'ZA', 'ZAF', 'Sudáfrica'),
  (168, 'SD', 'SDN', 'Sudán'),
  (169, 'SS', 'SSD', 'Sudán del Sur'),
  (170, 'SE', 'SWE', 'Suecia'),
  (171, 'CH', 'CHE', 'Suiza'),
  (172, 'SR', 'SUR', 'Surinam'),
  (173, 'TH', 'THA', 'Tailandia'),
  (174, 'TZ', 'TZA', 'Tanzania'),
  (175, 'TJ', 'TJK', 'Tayikistán'),
  (176, 'TL', 'TLS', 'Timor Oriental'),
  (177, 'TG', 'TGO', 'Togo'),
  (178, 'TO', 'TON', 'Tonga'),
  (179, 'TT', 'TTO', 'Trinidad y Tobago'),
  (180, 'TN', 'TUN', 'Túnez'),
  (181, 'TM', 'TKM', 'Turkmenistán'),
  (182, 'TR', 'TUR', 'Turquía'),
  (183, 'TV', 'TUV', 'Tuvalu'),
  (184, 'UA', 'UKR', 'Ucrania'),
  (185, 'UG', 'UGA', 'Uganda'),
  (186, 'UY', 'URY', 'Uruguay'),
  (187, 'UZ', 'UZB', 'Uzbekistán'),
  (188, 'VU', 'VUT', 'Vanuatu'),
  (189, 'VE', 'VEN', 'Venezuela'),
  (190, 'VN', 'VNM', 'Vietnam'),
  (191, 'YE', 'YEM', 'Yemen'),
  (192, 'DJ', 'DJI', 'Yibuti'),
  (193, 'ZM', 'ZMB', 'Zambia'),
  (194, 'ZW', 'ZWE', 'Zimbabue')
ON CONFLICT (codigo_iso2) DO NOTHING;

-- 24 Provincias de Ecuador (pais_id 52 = Ecuador; códigos INEC)
INSERT INTO provincia (id, pais_id, codigo, nombre) VALUES
  (1, 52, '01', 'Azuay'),
  (2, 52, '02', 'Bolívar'),
  (3, 52, '03', 'Cañar'),
  (4, 52, '04', 'Carchi'),
  (5, 52, '05', 'Cotopaxi'),
  (6, 52, '06', 'Chimborazo'),
  (7, 52, '07', 'El Oro'),
  (8, 52, '08', 'Esmeraldas'),
  (9, 52, '09', 'Guayas'),
  (10, 52, '10', 'Imbabura'),
  (11, 52, '11', 'Loja'),
  (12, 52, '12', 'Los Ríos'),
  (13, 52, '13', 'Manabí'),
  (14, 52, '14', 'Morona Santiago'),
  (15, 52, '15', 'Napo'),
  (16, 52, '16', 'Pastaza'),
  (17, 52, '17', 'Pichincha'),
  (18, 52, '18', 'Tungurahua'),
  (19, 52, '19', 'Zamora Chinchipe'),
  (20, 52, '20', 'Galápagos'),
  (21, 52, '21', 'Sucumbíos'),
  (22, 52, '22', 'Orellana'),
  (23, 52, '23', 'Santo Domingo de los Tsáchilas'),
  (24, 52, '24', 'Santa Elena')
ON CONFLICT (pais_id, codigo) DO NOTHING;

-- Cantones por provincia (provincia_id y prefijo de código según el INEC).
-- El sufijo del código es secuencial dentro de la provincia; reemplazar por el código
-- oficial del INEC al cargar el catálogo productivo (SDD 16.3).
INSERT INTO canton (id, provincia_id, codigo, nombre) VALUES
-- Azuay (01)
  (101, 1, '0101', 'Cuenca'),
  (102, 1, '0102', 'Gualaceo'),
  (103, 1, '0103', 'Chordeleg'),
  (104, 1, '0104', 'Sígsig'),
  (105, 1, '0105', 'Girón'),
  (106, 1, '0106', 'Santa Isabel'),
  (107, 1, '0107', 'Pucará'),
  (108, 1, '0108', 'Camilo Ponce Enríquez'),
-- Bolívar (02)
  (109, 2, '0201', 'Guaranda'),
  (110, 2, '0202', 'Chillanes'),
  (111, 2, '0203', 'Chimbo'),
  (112, 2, '0204', 'Echeandía'),
  (113, 2, '0205', 'San Miguel'),
  (114, 2, '0206', 'Caluma'),
  (115, 2, '0207', 'Las Naves'),
-- Cañar (03)
  (116, 3, '0301', 'Azogues'),
  (117, 3, '0302', 'Biblián'),
  (118, 3, '0303', 'Cañar'),
  (119, 3, '0304', 'Déleg'),
  (120, 3, '0305', 'El Tambo'),
  (121, 3, '0306', 'Suscal'),
  (122, 3, '0307', 'La Troncal'),
-- Carchi (04)
  (123, 4, '0401', 'Tulcán'),
  (124, 4, '0402', 'Bolívar'),
  (125, 4, '0403', 'Espejo'),
  (126, 4, '0404', 'Mira'),
  (127, 4, '0405', 'Montúfar'),
  (128, 4, '0406', 'San Pedro de Huaca'),
-- Chimborazo (06)
  (129, 6, '0601', 'Riobamba'),
  (130, 6, '0602', 'Alausí'),
  (131, 6, '0603', 'Colta'),
  (132, 6, '0604', 'Chunchi'),
  (133, 6, '0605', 'Guamote'),
  (134, 6, '0606', 'Guano'),
  (135, 6, '0607', 'Pallatanga'),
  (136, 6, '0608', 'Penipe'),
  (137, 6, '0609', 'Cumandá'),
-- Cotopaxi (05)
  (138, 5, '0501', 'Latacunga'),
  (139, 5, '0502', 'La Maná'),
  (140, 5, '0503', 'Pangua'),
  (141, 5, '0504', 'Pujilí'),
  (142, 5, '0505', 'Salcedo'),
  (143, 5, '0506', 'Saquisilí'),
  (144, 5, '0507', 'Sigchos'),
-- El Oro (07)
  (145, 7, '0701', 'Machala'),
  (146, 7, '0702', 'Arenillas'),
  (147, 7, '0703', 'Atahualpa'),
  (148, 7, '0704', 'Balsas'),
  (149, 7, '0705', 'Chilla'),
  (150, 7, '0706', 'El Guabo'),
  (151, 7, '0707', 'Huaquillas'),
  (152, 7, '0708', 'Las Lajas'),
  (153, 7, '0709', 'Marcabelí'),
  (154, 7, '0710', 'Pasaje'),
  (155, 7, '0711', 'Piñas'),
  (156, 7, '0712', 'Portovelo'),
  (157, 7, '0713', 'Santa Rosa'),
  (158, 7, '0714', 'Zaruma'),
-- Esmeraldas (08)
  (159, 8, '0801', 'Esmeraldas'),
  (160, 8, '0802', 'Atacames'),
  (161, 8, '0803', 'Eloy Alfaro'),
  (162, 8, '0804', 'Muisne'),
  (163, 8, '0805', 'Quinindé'),
  (164, 8, '0806', 'Rioverde'),
  (165, 8, '0807', 'San Lorenzo'),
-- Galápagos (20)
  (166, 20, '2001', 'San Cristóbal'),
  (167, 20, '2002', 'Isabela'),
  (168, 20, '2003', 'Santa Cruz'),
-- Guayas (09)
  (169, 9, '0901', 'Guayaquil'),
  (170, 9, '0902', 'Durán'),
  (171, 9, '0903', 'Milagro'),
  (172, 9, '0904', 'Samborondón'),
  (173, 9, '0905', 'Daule'),
  (174, 9, '0906', 'Playas'),
  (175, 9, '0907', 'Balzar'),
  (176, 9, '0908', 'Colimes'),
  (177, 9, '0909', 'El Empalme'),
  (178, 9, '0910', 'El Triunfo'),
  (179, 9, '0911', 'Isidro Ayora'),
  (180, 9, '0912', 'Lomas de Sargentillo'),
  (181, 9, '0913', 'Naranjal'),
  (182, 9, '0914', 'Naranjito'),
  (183, 9, '0915', 'Nobol'),
  (184, 9, '0916', 'Pedro Carbo'),
  (185, 9, '0917', 'Salitre'),
  (186, 9, '0918', 'Simón Bolívar'),
  (187, 9, '0919', 'Yaguachi'),
-- Imbabura (10)
  (188, 10, '1001', 'Ibarra'),
  (189, 10, '1002', 'Antonio Ante'),
  (190, 10, '1003', 'Cotacachi'),
  (191, 10, '1004', 'Otavalo'),
  (192, 10, '1005', 'Pimampiro'),
  (193, 10, '1006', 'San Miguel de Urcuquí'),
-- Loja (11)
  (194, 11, '1101', 'Loja'),
  (195, 11, '1102', 'Calvas'),
  (196, 11, '1103', 'Catamayo'),
  (197, 11, '1104', 'Celica'),
  (198, 11, '1105', 'Chaguarpamba'),
  (199, 11, '1106', 'Espíndola'),
  (200, 11, '1107', 'Gonzanamá'),
  (201, 11, '1108', 'Macará'),
  (202, 11, '1109', 'Paltas'),
  (203, 11, '1110', 'Puyango'),
  (204, 11, '1111', 'Saraguro'),
  (205, 11, '1112', 'Sozoranga'),
  (206, 11, '1113', 'Zapotillo'),
-- Los Ríos (12)
  (207, 12, '1201', 'Babahoyo'),
  (208, 12, '1202', 'Baba'),
  (209, 12, '1203', 'Buena Fe'),
  (210, 12, '1204', 'Mocache'),
  (211, 12, '1205', 'Montalvo'),
  (212, 12, '1206', 'Palenque'),
  (213, 12, '1207', 'Puebloviejo'),
  (214, 12, '1208', 'Quevedo'),
  (215, 12, '1209', 'Urdaneta'),
  (216, 12, '1210', 'Ventanas'),
  (217, 12, '1211', 'Vinces'),
  (218, 12, '1212', 'Valencia'),
-- Manabí (13)
  (219, 13, '1301', 'Portoviejo'),
  (220, 13, '1302', 'Manta'),
  (221, 13, '1303', 'Chone'),
  (222, 13, '1304', 'Bolívar'),
  (223, 13, '1305', 'El Carmen'),
  (224, 13, '1306', 'Flavio Alfaro'),
  (225, 13, '1307', 'Jipijapa'),
  (226, 13, '1308', 'Junín'),
  (227, 13, '1309', 'Montecristi'),
  (228, 13, '1310', 'Paján'),
  (229, 13, '1311', 'Pichincha'),
  (230, 13, '1312', 'Rocafuerte'),
  (231, 13, '1313', 'Santa Ana'),
  (232, 13, '1314', 'Sucre'),
  (233, 13, '1315', 'Tosagua'),
  (234, 13, '1316', '24 de Mayo'),
  (235, 13, '1317', 'Pedernales'),
  (236, 13, '1318', 'Olmedo'),
  (237, 13, '1319', 'Puerto López'),
  (238, 13, '1320', 'Jama'),
  (239, 13, '1321', 'Jaramijó'),
  (240, 13, '1322', 'San Vicente'),
-- Morona Santiago (14)
  (241, 14, '1401', 'Morona'),
  (242, 14, '1402', 'Gualaquiza'),
  (243, 14, '1403', 'Limón Indanza'),
  (244, 14, '1404', 'Palora'),
  (245, 14, '1405', 'San Juan Bosco'),
  (246, 14, '1406', 'Santiago'),
  (247, 14, '1407', 'Sucúa'),
  (248, 14, '1408', 'Taisha'),
  (249, 14, '1409', 'Tiwintza'),
-- Napo (15)
  (250, 15, '1501', 'Tena'),
  (251, 15, '1502', 'Archidona'),
  (252, 15, '1503', 'Carlos Julio Arosemena Tola'),
  (253, 15, '1504', 'El Chaco'),
  (254, 15, '1505', 'Quijos'),
-- Orellana (22)
  (255, 22, '2201', 'Francisco de Orellana'),
  (256, 22, '2202', 'Aguarico'),
  (257, 22, '2203', 'La Joya de los Sachas'),
  (258, 22, '2204', 'Loreto'),
-- Pastaza (16)
  (259, 16, '1601', 'Pastaza'),
  (260, 16, '1602', 'Arajuno'),
  (261, 16, '1603', 'Mera'),
  (262, 16, '1604', 'Santa Clara'),
-- Pichincha (17)
  (263, 17, '1701', 'Quito'),
  (264, 17, '1702', 'Cayambe'),
  (265, 17, '1703', 'Mejía'),
  (266, 17, '1704', 'Pedro Moncayo'),
  (267, 17, '1705', 'Pedro Vicente Maldonado'),
  (268, 17, '1706', 'Puerto Quito'),
  (269, 17, '1707', 'Rumiñahui'),
  (270, 17, '1708', 'San Miguel de los Bancos'),
-- Santa Elena (24)
  (271, 24, '2401', 'Santa Elena'),
  (272, 24, '2402', 'La Libertad'),
-- Santo Domingo de los Tsáchilas (23)
  (273, 23, '2301', 'La Concordia'),
  (274, 23, '2302', 'Santo Domingo'),
-- Sucumbíos (21)
  (275, 21, '2101', 'Cascales'),
  (276, 21, '2102', 'Cuyabeno'),
  (277, 21, '2103', 'Gonzalo Pizarro'),
  (278, 21, '2104', 'Lago Agrio'),
  (279, 21, '2105', 'Putumayo'),
  (280, 21, '2106', 'Shushufindi'),
  (281, 21, '2107', 'Sucumbíos'),
-- Tungurahua (18)
  (282, 18, '1801', 'Ambato'),
  (283, 18, '1802', 'Baños de Agua Santa'),
  (284, 18, '1803', 'Cevallos'),
  (285, 18, '1804', 'Mocha'),
  (286, 18, '1805', 'Patate'),
  (287, 18, '1806', 'Pelileo'),
  (288, 18, '1807', 'Píllaro'),
  (289, 18, '1808', 'Quero'),
  (290, 18, '1809', 'Tisaleo'),
-- Zamora Chinchipe (19)
  (291, 19, '1901', 'Centinela del Cóndor'),
  (292, 19, '1902', 'Chinchipe'),
  (293, 19, '1903', 'El Pangui'),
  (294, 19, '1904', 'Nangaritza'),
  (295, 19, '1905', 'Palanda'),
  (296, 19, '1906', 'Paquisha'),
  (297, 19, '1907', 'Yacuambi'),
  (298, 19, '1908', 'Yantzaza'),
  (299, 19, '1909', 'Zamora')
ON CONFLICT (provincia_id, codigo) DO NOTHING;

-- Cargos del Directorio del BIESS
INSERT INTO cargo (id, codigo, nombre, sector) VALUES
  (1, 'AFI', 'Miembro Principal y Alterno en Representación del Sector de Afiliados', 'AFILIADOS'),
  (2, 'JUB', 'Miembro Principal y Alterno en Representación del Sector de los Jubilados', 'JUBILADOS'),
  (3, 'EMP', 'Miembro Principal y Alterno en Representación del Sector Empleador', 'EMPLEADORES')
ON CONFLICT (codigo) DO NOTHING;

-- Convocatoria de demostración
INSERT INTO convocatoria (
  id, codigo, nombre, fecha_inicio_postulacion, fecha_fin_postulacion,
  fase_postulacion_habilitada, version_formulario
) VALUES (
  1, 'BIESS-2026', 'Concurso público de méritos y oposición — Directorio del BIESS 2026',
  DATE '2026-10-01', DATE '2026-11-30', true, '2026.1'
)
ON CONFLICT (codigo) DO NOTHING;

-- 13 Preguntas de Inhabilidades (RF v1.4, páginas 7 y 8)
INSERT INTO inhabilidad_pregunta (id, codigo, texto, orden, efecto, respuesta_bloqueante, activo) VALUES
  (1,  '1.9.1',  '¿Es ecuatoriano/a y está en pleno goce de los derechos de participación política?', 1,  'BLOQUEANTE', false, true),
  (2,  '1.9.2',  '¿Se encuentra inhabilitado para ejercer el comercio?', 2,  'ALERTA', NULL, true),
  (3,  '1.9.3',  '¿Está en mora, directa o indirectamente, en el pago de sus obligaciones en cualquiera de las instituciones del Estado o bajo control de la Superintendencia de Bancos?', 3,  'ALERTA', NULL, true),
  (4,  '1.9.4',  '¿Mantiene algún vínculo, por propiedad o administración, con instituciones del sistema financiero y seguros privados, bolsas de valores o administradoras de fondos?', 4,  'ALERTA', NULL, true),
  (5,  '1.9.5',  '¿Es funcionario o empleado del IESS - BIESS actualmente?', 5,  'ALERTA', NULL, true),
  (6,  '1.9.6',  '¿Mantiene interés propio o representa a terceros en compañías aseguradoras del sistema de seguridad social?', 6,  'ALERTA', NULL, true),
  (7,  '1.9.7',  '¿En los últimos 5 años, ha sido removido por algún Organismo de Control?', 7,  'ALERTA', NULL, true),
  (8,  '1.9.8',  '¿Mantiene obligaciones pendientes con el SRI?', 8,  'ALERTA', NULL, true),
  (9,  '1.9.9',  '¿Ha incurrido en los últimos 5 años en castigo de sus obligaciones por parte de cualquier institución financiera?', 9,  'ALERTA', NULL, true),
  (10, '1.9.10', '¿Actualmente se encuentra litigando en contra del IESS - BIESS?', 10, 'ALERTA', NULL, true),
  (11, '1.9.11', '¿Se encuentra procesado, enjuiciado o condenado por algún delito relacionado con corrupción o crimen organizado?', 11, 'ALERTA', NULL, true),
  (12, '1.9.12', '¿Mantiene en firme algún tipo de responsabilidades por la Contraloría General del Estado?', 12, 'ALERTA', NULL, true),
  (13, '1.9.13', '¿Se encuentra registrado en la base de datos de la Unidad de Análisis Financiero y Económico (UAFE)?', 13, 'ALERTA', NULL, true)
ON CONFLICT (codigo) DO NOTHING;

-- Sincronización de secuencias
SELECT setval(pg_get_serial_sequence('pais', 'id'), COALESCE((SELECT max(id) FROM pais), 1), true);
SELECT setval(pg_get_serial_sequence('provincia', 'id'), COALESCE((SELECT max(id) FROM provincia), 1), true);
SELECT setval(pg_get_serial_sequence('canton', 'id'), COALESCE((SELECT max(id) FROM canton), 1), true);
SELECT setval(pg_get_serial_sequence('cargo', 'id'), COALESCE((SELECT max(id) FROM cargo), 1), true);
SELECT setval(pg_get_serial_sequence('convocatoria', 'id'), COALESCE((SELECT max(id) FROM convocatoria), 1), true);
SELECT setval(pg_get_serial_sequence('inhabilidad_pregunta', 'id'), COALESCE((SELECT max(id) FROM inhabilidad_pregunta), 1), true);

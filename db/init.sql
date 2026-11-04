-- =====================================================================
-- SRP-BIESS · Base de Datos PostgreSQL 16
-- Concurso de Méritos y Oposición Directorio BIESS (EF-11 y EF-12)
-- Script Inicial DDL + Semillas (Idempotente)
--
-- Reiniciar la bd para que ejecute los nuevos cambios sin flujo de pago con este archivo:
-- ! docker compose stop postgres && docker compose rm -f postgres && docker volume rm srp-biess_postgres_data && docker compose up -d postgres
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
  email_secundario    text,
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
  updated_at         timestamptz NOT NULL DEFAULT now()
);

-- Domicilio en Ecuador: provincia y cantón del catálogo; fuera de Ecuador: provincia y
-- ciudad escritas (RF-EF12-07). Un CHECK no puede consultar otra tabla, por eso se usa un
-- trigger que reconoce a Ecuador por su código ISO 'EC' y no por su id (SRP-019).
CREATE OR REPLACE FUNCTION fn_domicilio_valida_ubicacion() RETURNS trigger AS $$
DECLARE
  es_ecuador boolean;
BEGIN
  SELECT codigo_iso2 = 'EC' INTO es_ecuador FROM pais WHERE id = NEW.pais_id;
  IF es_ecuador THEN
    IF NEW.provincia_id IS NULL OR NEW.canton_id IS NULL
       OR NEW.provincia_manual IS NOT NULL OR NEW.ciudad_manual IS NOT NULL THEN
      RAISE EXCEPTION 'Un domicilio en Ecuador requiere provincia y cantón del catálogo, sin provincia ni ciudad escritas'
        USING ERRCODE = 'check_violation';
    END IF;
    IF NOT EXISTS (
      SELECT 1 FROM canton c JOIN provincia p ON p.id = c.provincia_id
      WHERE c.id = NEW.canton_id AND p.id = NEW.provincia_id AND p.pais_id = NEW.pais_id
    ) THEN
      RAISE EXCEPTION 'El cantón no pertenece a la provincia o la provincia no pertenece al país'
        USING ERRCODE = 'check_violation';
    END IF;
  ELSIF NEW.provincia_manual IS NULL OR NEW.ciudad_manual IS NULL THEN
    RAISE EXCEPTION 'Un domicilio fuera de Ecuador requiere provincia o estado y ciudad'
      USING ERRCODE = 'check_violation';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE domicilio_postulante DROP CONSTRAINT IF EXISTS ck_domicilio_pais_ec;
DROP TRIGGER IF EXISTS trg_domicilio_valida_ubicacion ON domicilio_postulante;
CREATE TRIGGER trg_domicilio_valida_ubicacion
  BEFORE INSERT OR UPDATE ON domicilio_postulante
  FOR EACH ROW EXECUTE FUNCTION fn_domicilio_valida_ubicacion();

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

CREATE TABLE IF NOT EXISTS postulacion_envio_idempotencia (
  idempotency_key varchar(255) PRIMARY KEY,
  postulacion_id   uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  response         jsonb NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_postulacion_envio_idempotencia_postulacion
  ON postulacion_envio_idempotencia (postulacion_id);

-- ---------------------------------------------------------------------
-- 7. Notificaciones (Outbox) y Auditoría
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS notificacion (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id  uuid REFERENCES postulacion(id) ON DELETE SET NULL,
  postulante_id   uuid REFERENCES postulante(id) ON DELETE SET NULL,
  tipo            varchar(40) NOT NULL,
  destinatario    varchar(254) NOT NULL,
  cc              text[] DEFAULT '{}'::text[],
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

-- 24 Provincias de Ecuador (pais_id 52 = Ecuador en la semilla de países; códigos INEC)
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

-- Cantones del Ecuador con su código oficial del INEC (SRP-019).
-- Fuente: INEC, Clasificador Geográfico Estadístico, códigos de la División Político
--   Administrativa 2026, hoja CANTONES:
--   https://aplicaciones2.ecuadorencifras.gob.ec/SIN/descargas/cdpa2026.xlsx
--   Fecha de corte: descargado el 03/11/2026 (archivo publicado el 27/02/2026).
-- 222 cantones en las 24 provincias. Se excluye la "Zona en estudio" (código 90), que no
-- pertenece a ninguna provincia. Los nombres pasan de mayúsculas a mayúscula inicial.
-- El id es el código INEC sin el cero inicial (0101 -> 101). Idempotente: si el canton
-- existe, se actualizan su provincia, código y nombre por id; los cantones del Ecuador
-- que ya no figuran en la fuente se desactivan (no se borran, por los domicilios).
-- Bases creadas con la semilla anterior, cuyos ids no coincidían con el código: se libera el
-- código de esas filas para que la carga por id no choque con uq_canton_provincia_codigo.
-- Sus ids pasan a representar otro cantón, así que en esas bases conviene recrear la base.
UPDATE canton SET codigo = 'x' || id, activo = false
WHERE codigo <> lpad(id::text, 4, '0')
  AND provincia_id IN (
    SELECT p.id FROM provincia p JOIN pais ON pais.id = p.pais_id WHERE pais.codigo_iso2 = 'EC'
  );

WITH oficial (id, provincia_id, codigo, nombre) AS (VALUES
  -- Azuay (01)
  (101, 1, '0101', 'Cuenca'),
  (102, 1, '0102', 'Girón'),
  (103, 1, '0103', 'Gualaceo'),
  (104, 1, '0104', 'Nabón'),
  (105, 1, '0105', 'Paute'),
  (106, 1, '0106', 'Pucará'),
  (107, 1, '0107', 'San Fernando'),
  (108, 1, '0108', 'Santa Isabel'),
  (109, 1, '0109', 'Sígsig'),
  (110, 1, '0110', 'Oña'),
  (111, 1, '0111', 'Chordeleg'),
  (112, 1, '0112', 'El Pan'),
  (113, 1, '0113', 'Sevilla de Oro'),
  (114, 1, '0114', 'Guachapala'),
  (115, 1, '0115', 'Camilo Ponce Enríquez'),
  -- Bolívar (02)
  (201, 2, '0201', 'Guaranda'),
  (202, 2, '0202', 'Chillanes'),
  (203, 2, '0203', 'Chimbo'),
  (204, 2, '0204', 'Echeandía'),
  (205, 2, '0205', 'San Miguel'),
  (206, 2, '0206', 'Caluma'),
  (207, 2, '0207', 'Las Naves'),
  -- Cañar (03)
  (301, 3, '0301', 'Azogues'),
  (302, 3, '0302', 'Biblián'),
  (303, 3, '0303', 'Cañar'),
  (304, 3, '0304', 'La Troncal'),
  (305, 3, '0305', 'El Tambo'),
  (306, 3, '0306', 'Déleg'),
  (307, 3, '0307', 'Suscal'),
  -- Carchi (04)
  (401, 4, '0401', 'Tulcán'),
  (402, 4, '0402', 'Bolívar'),
  (403, 4, '0403', 'Espejo'),
  (404, 4, '0404', 'Mira'),
  (405, 4, '0405', 'Montúfar'),
  (406, 4, '0406', 'San Pedro de Huaca'),
  -- Cotopaxi (05)
  (501, 5, '0501', 'Latacunga'),
  (502, 5, '0502', 'La Maná'),
  (503, 5, '0503', 'Pangua'),
  (504, 5, '0504', 'Pujilí'),
  (505, 5, '0505', 'Salcedo'),
  (506, 5, '0506', 'Saquisilí'),
  (507, 5, '0507', 'Sigchos'),
  -- Chimborazo (06)
  (601, 6, '0601', 'Riobamba'),
  (602, 6, '0602', 'Alausí'),
  (603, 6, '0603', 'Colta'),
  (604, 6, '0604', 'Chambo'),
  (605, 6, '0605', 'Chunchi'),
  (606, 6, '0606', 'Guamote'),
  (607, 6, '0607', 'Guano'),
  (608, 6, '0608', 'Pallatanga'),
  (609, 6, '0609', 'Penipe'),
  (610, 6, '0610', 'Cumandá'),
  -- El Oro (07)
  (701, 7, '0701', 'Machala'),
  (702, 7, '0702', 'Arenillas'),
  (703, 7, '0703', 'Atahualpa'),
  (704, 7, '0704', 'Balsas'),
  (705, 7, '0705', 'Chilla'),
  (706, 7, '0706', 'El Guabo'),
  (707, 7, '0707', 'Huaquillas'),
  (708, 7, '0708', 'Marcabelí'),
  (709, 7, '0709', 'Pasaje'),
  (710, 7, '0710', 'Piñas'),
  (711, 7, '0711', 'Portovelo'),
  (712, 7, '0712', 'Santa Rosa'),
  (713, 7, '0713', 'Zaruma'),
  (714, 7, '0714', 'Las Lajas'),
  -- Esmeraldas (08)
  (801, 8, '0801', 'Esmeraldas'),
  (802, 8, '0802', 'Eloy Alfaro'),
  (803, 8, '0803', 'Muisne'),
  (804, 8, '0804', 'Quinindé'),
  (805, 8, '0805', 'San Lorenzo'),
  (806, 8, '0806', 'Atacames'),
  (807, 8, '0807', 'Rioverde'),
  -- Guayas (09)
  (901, 9, '0901', 'Guayaquil'),
  (902, 9, '0902', 'Alfredo Baquerizo Moreno (Juján)'),
  (903, 9, '0903', 'Balao'),
  (904, 9, '0904', 'Balzar'),
  (905, 9, '0905', 'Colimes'),
  (906, 9, '0906', 'Daule'),
  (907, 9, '0907', 'Durán'),
  (908, 9, '0908', 'El Empalme'),
  (909, 9, '0909', 'El Triunfo'),
  (910, 9, '0910', 'Milagro'),
  (911, 9, '0911', 'Naranjal'),
  (912, 9, '0912', 'Naranjito'),
  (913, 9, '0913', 'Palestina'),
  (914, 9, '0914', 'Pedro Carbo'),
  (916, 9, '0916', 'Samborondón'),
  (918, 9, '0918', 'Santa Lucía'),
  (919, 9, '0919', 'Salitre'),
  (920, 9, '0920', 'San Jacinto de Yaguachi'),
  (921, 9, '0921', 'Playas'),
  (922, 9, '0922', 'Simón Bolívar'),
  (923, 9, '0923', 'Coronel Marcelino Maridueña'),
  (924, 9, '0924', 'Lomas de Sargentillo'),
  (925, 9, '0925', 'Nobol'),
  (927, 9, '0927', 'General Antonio Elizalde'),
  (928, 9, '0928', 'Isidro Ayora'),
  -- Imbabura (10)
  (1001, 10, '1001', 'Ibarra'),
  (1002, 10, '1002', 'Antonio Ante'),
  (1003, 10, '1003', 'Cotacachi'),
  (1004, 10, '1004', 'Otavalo'),
  (1005, 10, '1005', 'Pimampiro'),
  (1006, 10, '1006', 'San Miguel de Urcuquí'),
  -- Loja (11)
  (1101, 11, '1101', 'Loja'),
  (1102, 11, '1102', 'Calvas'),
  (1103, 11, '1103', 'Catamayo'),
  (1104, 11, '1104', 'Celica'),
  (1105, 11, '1105', 'Chaguarpamba'),
  (1106, 11, '1106', 'Espíndola'),
  (1107, 11, '1107', 'Gonzanamá'),
  (1108, 11, '1108', 'Macará'),
  (1109, 11, '1109', 'Paltas'),
  (1110, 11, '1110', 'Puyango'),
  (1111, 11, '1111', 'Saraguro'),
  (1112, 11, '1112', 'Sozoranga'),
  (1113, 11, '1113', 'Zapotillo'),
  (1114, 11, '1114', 'Pindal'),
  (1115, 11, '1115', 'Quilanga'),
  (1116, 11, '1116', 'Olmedo'),
  -- Los Ríos (12)
  (1201, 12, '1201', 'Babahoyo'),
  (1202, 12, '1202', 'Baba'),
  (1203, 12, '1203', 'Montalvo'),
  (1204, 12, '1204', 'Puebloviejo'),
  (1205, 12, '1205', 'Quevedo'),
  (1206, 12, '1206', 'Urdaneta'),
  (1207, 12, '1207', 'Ventanas'),
  (1208, 12, '1208', 'Vinces'),
  (1209, 12, '1209', 'Palenque'),
  (1210, 12, '1210', 'Buena Fe'),
  (1211, 12, '1211', 'Valencia'),
  (1212, 12, '1212', 'Mocache'),
  (1213, 12, '1213', 'Quinsaloma'),
  -- Manabí (13)
  (1301, 13, '1301', 'Portoviejo'),
  (1302, 13, '1302', 'Bolívar'),
  (1303, 13, '1303', 'Chone'),
  (1304, 13, '1304', 'El Carmen'),
  (1305, 13, '1305', 'Flavio Alfaro'),
  (1306, 13, '1306', 'Jipijapa'),
  (1307, 13, '1307', 'Junín'),
  (1308, 13, '1308', 'Manta'),
  (1309, 13, '1309', 'Montecristi'),
  (1310, 13, '1310', 'Paján'),
  (1311, 13, '1311', 'Pichincha'),
  (1312, 13, '1312', 'Rocafuerte'),
  (1313, 13, '1313', 'Santa Ana'),
  (1314, 13, '1314', 'Sucre'),
  (1315, 13, '1315', 'Tosagua'),
  (1316, 13, '1316', '24 de Mayo'),
  (1317, 13, '1317', 'Pedernales'),
  (1318, 13, '1318', 'Olmedo'),
  (1319, 13, '1319', 'Puerto López'),
  (1320, 13, '1320', 'Jama'),
  (1321, 13, '1321', 'Jaramijó'),
  (1322, 13, '1322', 'San Vicente'),
  -- Morona Santiago (14)
  (1401, 14, '1401', 'Morona'),
  (1402, 14, '1402', 'Gualaquiza'),
  (1403, 14, '1403', 'Limón Indanza'),
  (1404, 14, '1404', 'Palora'),
  (1405, 14, '1405', 'Santiago'),
  (1406, 14, '1406', 'Sucúa'),
  (1407, 14, '1407', 'Huamboya'),
  (1408, 14, '1408', 'San Juan Bosco'),
  (1409, 14, '1409', 'Taisha'),
  (1410, 14, '1410', 'Logroño'),
  (1411, 14, '1411', 'Pablo Sexto'),
  (1412, 14, '1412', 'Tiwintza'),
  (1413, 14, '1413', 'Sevilla Don Bosco'),
  -- Napo (15)
  (1501, 15, '1501', 'Tena'),
  (1503, 15, '1503', 'Archidona'),
  (1504, 15, '1504', 'El Chaco'),
  (1507, 15, '1507', 'Quijos'),
  (1509, 15, '1509', 'Carlos Julio Arosemena Tola'),
  -- Pastaza (16)
  (1601, 16, '1601', 'Pastaza'),
  (1602, 16, '1602', 'Mera'),
  (1603, 16, '1603', 'Santa Clara'),
  (1604, 16, '1604', 'Arajuno'),
  -- Pichincha (17)
  (1701, 17, '1701', 'Distrito Metropolitano de Quito'),
  (1702, 17, '1702', 'Cayambe'),
  (1703, 17, '1703', 'Mejía'),
  (1704, 17, '1704', 'Pedro Moncayo'),
  (1705, 17, '1705', 'Rumiñahui'),
  (1707, 17, '1707', 'San Miguel de los Bancos'),
  (1708, 17, '1708', 'Pedro Vicente Maldonado'),
  (1709, 17, '1709', 'Puerto Quito'),
  -- Tungurahua (18)
  (1801, 18, '1801', 'Ambato'),
  (1802, 18, '1802', 'Baños de Agua Santa'),
  (1803, 18, '1803', 'Cevallos'),
  (1804, 18, '1804', 'Mocha'),
  (1805, 18, '1805', 'Patate'),
  (1806, 18, '1806', 'Quero'),
  (1807, 18, '1807', 'San Pedro de Pelileo'),
  (1808, 18, '1808', 'Santiago de Píllaro'),
  (1809, 18, '1809', 'Tisaleo'),
  -- Zamora Chinchipe (19)
  (1901, 19, '1901', 'Zamora'),
  (1902, 19, '1902', 'Chinchipe'),
  (1903, 19, '1903', 'Nangaritza'),
  (1904, 19, '1904', 'Yacuambi'),
  (1905, 19, '1905', 'Yantzaza'),
  (1906, 19, '1906', 'El Pangui'),
  (1907, 19, '1907', 'Centinela del Cóndor'),
  (1908, 19, '1908', 'Palanda'),
  (1909, 19, '1909', 'Paquisha'),
  -- Galápagos (20)
  (2001, 20, '2001', 'San Cristóbal'),
  (2002, 20, '2002', 'Isabela'),
  (2003, 20, '2003', 'Santa Cruz'),
  -- Sucumbíos (21)
  (2101, 21, '2101', 'Lago Agrio'),
  (2102, 21, '2102', 'Gonzalo Pizarro'),
  (2103, 21, '2103', 'Putumayo'),
  (2104, 21, '2104', 'Shushufindi'),
  (2105, 21, '2105', 'Sucumbíos'),
  (2106, 21, '2106', 'Cascales'),
  (2107, 21, '2107', 'Cuyabeno'),
  -- Orellana (22)
  (2201, 22, '2201', 'Francisco de Orellana'),
  (2202, 22, '2202', 'Aguarico'),
  (2203, 22, '2203', 'La Joya de los Sachas'),
  (2204, 22, '2204', 'Loreto'),
  -- Santo Domingo de los Tsáchilas (23)
  (2301, 23, '2301', 'Santo Domingo'),
  (2302, 23, '2302', 'La Concordia'),
  -- Santa Elena (24)
  (2401, 24, '2401', 'Santa Elena'),
  (2402, 24, '2402', 'La Libertad'),
  (2403, 24, '2403', 'Salinas')
), cargados AS (
  INSERT INTO canton (id, provincia_id, codigo, nombre, activo)
  SELECT id::smallint, provincia_id::smallint, codigo, nombre, true FROM oficial
  ON CONFLICT (id) DO UPDATE
    SET provincia_id = EXCLUDED.provincia_id, codigo = EXCLUDED.codigo,
        nombre = EXCLUDED.nombre, activo = true
  RETURNING id
)
UPDATE canton SET activo = false
WHERE id NOT IN (SELECT id FROM oficial)
  AND provincia_id IN (
    SELECT p.id FROM provincia p JOIN pais ON pais.id = p.pais_id WHERE pais.codigo_iso2 = 'EC'
  );

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


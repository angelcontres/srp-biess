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
  CREATE TYPE tipo_documento AS ENUM ('RESPALDO_ORGANIZACIONAL', 'TITULO', 'EXPERIENCIA', 'PUBLICACION', 'CAPACITACION', 'FORMULARIO_POSTULACION', 'COMPROBANTE_PAGO');
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

DO $$ BEGIN
  CREATE TYPE estado_pago AS ENUM ('PENDIENTE', 'APROBADO', 'RECHAZADO', 'REVERTIDO');
EXCEPTION WHEN duplicate_object THEN null; END $$;

DO $$ BEGIN
  CREATE TYPE estado_factura AS ENUM ('EMITIDA', 'ANULADA', 'FALLIDA');
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
  CONSTRAINT ck_domicilio_pais_ec CHECK (
    (pais_id = 1 AND provincia_id IS NOT NULL AND canton_id IS NOT NULL AND provincia_manual IS NULL AND ciudad_manual IS NULL)
    OR (pais_id <> 1 AND provincia_manual IS NOT NULL AND ciudad_manual IS NOT NULL)
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
  estado_arancel                varchar(20) NOT NULL DEFAULT 'PENDIENTE',
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
-- 7. Pagos y Facturación Electrónica (ADR-014)
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS transaccion_pago (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  postulacion_id      uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  postulante_id       uuid NOT NULL REFERENCES postulante(id) ON DELETE CASCADE,
  codigo_transaccion  varchar(64) NOT NULL UNIQUE,
  monto               numeric(10,2) NOT NULL CHECK (monto > 0),
  moneda              char(3) NOT NULL DEFAULT 'USD',
  franquicia          varchar(30) NOT NULL,
  ultimos_cuatro      char(4) NOT NULL,
  estado              estado_pago NOT NULL DEFAULT 'PENDIENTE',
  gateway_provider    varchar(30) NOT NULL DEFAULT 'MOCK',
  autorizacion_codigo varchar(64),
  fecha_transaccion   timestamptz NOT NULL DEFAULT now(),
  error_mensaje       text
);

CREATE TABLE IF NOT EXISTS factura (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  transaccion_pago_id uuid NOT NULL UNIQUE REFERENCES transaccion_pago(id) ON DELETE CASCADE,
  postulacion_id      uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  numero_factura      varchar(30) NOT NULL UNIQUE,
  clave_acceso_sri    char(49) NOT NULL UNIQUE,
  subtotal            numeric(10,2) NOT NULL,
  iva                 numeric(10,2) NOT NULL,
  total               numeric(10,2) NOT NULL,
  estado              estado_factura NOT NULL DEFAULT 'EMITIDA',
  enviada_correo      boolean NOT NULL DEFAULT false,
  documento_pdf_id    uuid REFERENCES documento(id) ON DELETE SET NULL,
  xml_autorizado      text,
  created_at          timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------
-- 8. Notificaciones (Outbox) y Auditoría
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
-- 9. Datos Semilla (Seeds) Idempotentes
-- ---------------------------------------------------------------------

-- Países
INSERT INTO pais (id, codigo_iso2, codigo_iso3, nombre) VALUES
  (1, 'EC', 'ECU', 'Ecuador'),
  (2, 'CO', 'COL', 'Colombia'),
  (3, 'PE', 'PER', 'Perú'),
  (4, 'US', 'USA', 'Estados Unidos'),
  (5, 'ES', 'ESP', 'España')
ON CONFLICT (codigo_iso2) DO NOTHING;

-- 24 Provincias de Ecuador
INSERT INTO provincia (id, pais_id, codigo, nombre) VALUES
  (1,  1, '01', 'Azuay'),
  (2,  1, '02', 'Bolívar'),
  (3,  1, '03', 'Cañar'),
  (4,  1, '04', 'Carchi'),
  (5,  1, '05', 'Cotopaxi'),
  (6,  1, '06', 'Chimborazo'),
  (7,  1, '07', 'El Oro'),
  (8,  1, '08', 'Esmeraldas'),
  (9,  1, '09', 'Guayas'),
  (10, 1, '10', 'Imbabura'),
  (11, 1, '11', 'Loja'),
  (12, 1, '12', 'Los Ríos'),
  (13, 1, '13', 'Manabí'),
  (14, 1, '14', 'Morona Santiago'),
  (15, 1, '15', 'Napo'),
  (16, 1, '16', 'Pastaza'),
  (17, 1, '17', 'Pichincha'),
  (18, 1, '18', 'Tungurahua'),
  (19, 1, '19', 'Zamora Chinchipe'),
  (20, 1, '20', 'Galápagos'),
  (21, 1, '21', 'Sucumbíos'),
  (22, 1, '22', 'Orellana'),
  (23, 1, '23', 'Santo Domingo de los Tsáchilas'),
  (24, 1, '24', 'Santa Elena')
ON CONFLICT (pais_id, codigo) DO NOTHING;

-- Cantones de demostración para cascada geográfica
INSERT INTO canton (id, provincia_id, codigo, nombre) VALUES
-- Azuay (01)
  (101,  1, '0101', 'Cuenca'),
  (102,  1, '0102', 'Gualaceo'),
  (103,  1, '0103', 'Chordeleg'),
  (104,  1, '0104', 'Sígsig'),
  (105,  1, '0105', 'Girón'),
  (106,  1, '0106', 'Santa Isabel'),
  (107,  1, '0107', 'Pucará'),
  (108,  1, '0108', 'Camilo Ponce Enríquez'),
-- Bolívar (02)
  (109,  2, '0201', 'Guaranda'),
  (110,  2, '0202', 'Chillanes'),
  (111,  2, '0203', 'Chimbo'),
  (112,  2, '0204', 'Echeandía'),
  (113,  2, '0205', 'San Miguel'),
  (114,  2, '0206', 'Caluma'),
  (115,  2, '0207', 'Las Naves'),
-- Cañar (03)
  (116,  3, '0301', 'Azogues'),
  (117,  3, '0302', 'Biblián'),
  (118,  3, '0303', 'Cañar'),
  (119,  3, '0304', 'Déleg'),
  (120,  3, '0305', 'El Tambo'),
  (121,  3, '0306', 'Suscal'),
  (122,  3, '0307', 'La Troncal'),
-- Carchi (04)
  (123,  4, '0401', 'Tulcán'),
  (124,  4, '0402', 'Bolívar'),
  (125,  4, '0403', 'Espejo'),
  (126,  4, '0404', 'Mira'),
  (127,  4, '0405', 'Montúfar'),
  (128,  4, '0406', 'San Pedro de Huaca'),
-- Chimborazo (05)
  (129,  5, '0501', 'Riobamba'),
  (130,  5, '0502', 'Alausí'),
  (131,  5, '0503', 'Colta'),
  (132,  5, '0504', 'Chunchi'),
  (133,  5, '0505', 'Guamote'),
  (134,  5, '0506', 'Guano'),
  (135,  5, '0507', 'Pallatanga'),
  (136,  5, '0508', 'Penipe'),
  (137,  5, '0509', 'Cumandá'),
-- Cotopaxi (06)
  (138,  6, '0601', 'Latacunga'),
  (139,  6, '0602', 'La Maná'),
  (140,  6, '0603', 'Pangua'),
  (141,  6, '0604', 'Pujilí'),
  (142,  6, '0605', 'Salcedo'),
  (143,  6, '0606', 'Saquisilí'),
  (144,  6, '0607', 'Sigchos'),
-- El Oro (07)
  (145,  7, '0701', 'Machala'),
  (146,  7, '0702', 'Arenillas'),
  (147,  7, '0703', 'Atahualpa'),
  (148,  7, '0704', 'Balsas'),
  (149,  7, '0705', 'Chilla'),
  (150,  7, '0706', 'El Guabo'),
  (151,  7, '0707', 'Huaquillas'),
  (152,  7, '0708', 'Las Lajas'),
  (153,  7, '0709', 'Marcabelí'),
  (154,  7, '0710', 'Pasaje'),
  (155,  7, '0711', 'Piñas'),
  (156,  7, '0712', 'Portovelo'),
  (157,  7, '0713', 'Santa Rosa'),
  (158,  7, '0714', 'Zaruma'),
-- Esmeraldas (08)
  (159,  8, '0801', 'Esmeraldas'),
  (160,  8, '0802', 'Atacames'),
  (161,  8, '0803', 'Eloy Alfaro'),
  (162,  8, '0804', 'Muisne'),
  (163,  8, '0805', 'Quinindé'),
  (164,  8, '0806', 'Rioverde'),
  (165,  8, '0807', 'San Lorenzo'),
-- Galápagos (09)
  (166,  9, '0901', 'San Cristóbal'),
  (167,  9, '0902', 'Isabela'),
  (168,  9, '0903', 'Santa Cruz'),
-- Guayas (10)
  (169, 10, '1001', 'Guayaquil'),
  (170, 10, '1002', 'Durán'),
  (171, 10, '1003', 'Milagro'),
  (172, 10, '1004', 'Samborondón'),
  (173, 10, '1005', 'Daule'),
  (174, 10, '1006', 'Playas'),
  (175, 10, '1007', 'Balzar'),
  (176, 10, '1008', 'Colimes'),
  (177, 10, '1009', 'El Empalme'),
  (178, 10, '1010', 'El Triunfo'),
  (179, 10, '1011', 'Isidro Ayora'),
  (180, 10, '1012', 'Lomas de Sargentillo'),
  (181, 10, '1013', 'Naranjal'),
  (182, 10, '1014', 'Naranjito'),
  (183, 10, '1015', 'Nobol'),
  (184, 10, '1016', 'Pedro Carbo'),
  (185, 10, '1017', 'Salitre'),
  (186, 10, '1018', 'Simón Bolívar'),
  (187, 10, '1019', 'Yaguachi'),
-- Imbabura (11)
  (188, 11, '1101', 'Ibarra'),
  (189, 11, '1102', 'Antonio Ante'),
  (190, 11, '1103', 'Cotacachi'),
  (191, 11, '1104', 'Otavalo'),
  (192, 11, '1105', 'Pimampiro'),
  (193, 11, '1106', 'San Miguel de Urcuquí'),
-- Loja (12)
  (194, 12, '1201', 'Loja'),
  (195, 12, '1202', 'Calvas'),
  (196, 12, '1203', 'Catamayo'),
  (197, 12, '1204', 'Celica'),
  (198, 12, '1205', 'Chaguarpamba'),
  (199, 12, '1206', 'Espíndola'),
  (200, 12, '1207', 'Gonzanamá'),
  (201, 12, '1208', 'Macará'),
  (202, 12, '1209', 'Paltas'),
  (203, 12, '1210', 'Puyango'),
  (204, 12, '1211', 'Saraguro'),
  (205, 12, '1212', 'Sozoranga'),
  (206, 12, '1213', 'Zapotillo'),
-- Los Ríos (13)
  (207, 13, '1301', 'Babahoyo'),
  (208, 13, '1302', 'Baba'),
  (209, 13, '1303', 'Buena Fe'),
  (210, 13, '1304', 'Mocache'),
  (211, 13, '1305', 'Montalvo'),
  (212, 13, '1306', 'Palenque'),
  (213, 13, '1307', 'Puebloviejo'),
  (214, 13, '1308', 'Quevedo'),
  (215, 13, '1309', 'Urdaneta'),
  (216, 13, '1310', 'Ventanas'),
  (217, 13, '1311', 'Vinces'),
  (218, 13, '1312', 'Valencia'),
-- Manabí (14)
  (219, 14, '1401', 'Portoviejo'),
  (220, 14, '1402', 'Manta'),
  (221, 14, '1403', 'Chone'),
  (222, 14, '1404', 'Bolívar'),
  (223, 14, '1405', 'El Carmen'),
  (224, 14, '1406', 'Flavio Alfaro'),
  (225, 14, '1407', 'Jipijapa'),
  (226, 14, '1408', 'Junín'),
  (227, 14, '1409', 'Montecristi'),
  (228, 14, '1410', 'Paján'),
  (229, 14, '1411', 'Pichincha'),
  (230, 14, '1412', 'Rocafuerte'),
  (231, 14, '1413', 'Santa Ana'),
  (232, 14, '1414', 'Sucre'),
  (233, 14, '1415', 'Tosagua'),
  (234, 14, '1416', '24 de Mayo'),
  (235, 14, '1417', 'Pedernales'),
  (236, 14, '1418', 'Olmedo'),
  (237, 14, '1419', 'Puerto López'),
  (238, 14, '1420', 'Jama'),
  (239, 14, '1421', 'Jaramijó'),
  (240, 14, '1422', 'San Vicente'),
-- Morona Santiago (15)
  (241, 15, '1501', 'Morona'),
  (242, 15, '1502', 'Gualaquiza'),
  (243, 15, '1503', 'Limón Indanza'),
  (244, 15, '1504', 'Palora'),
  (245, 15, '1505', 'San Juan Bosco'),
  (246, 15, '1506', 'Santiago'),
  (247, 15, '1507', 'Sucúa'),
  (248, 15, '1508', 'Taisha'),
  (249, 15, '1509', 'Tiwintza'),
-- Napo (16)
  (250, 16, '1601', 'Tena'),
  (251, 16, '1602', 'Archidona'),
  (252, 16, '1603', 'Carlos Julio Arosemena Tola'),
  (253, 16, '1604', 'El Chaco'),
  (254, 16, '1605', 'Quijos'),
-- Orellana (17)
  (255, 17, '1701', 'Francisco de Orellana'),
  (256, 17, '1702', 'Aguarico'),
  (257, 17, '1703', 'La Joya de los Sachas'),
  (258, 17, '1704', 'Loreto'),
-- Pastaza (18)
  (259, 18, '1801', 'Pastaza'),
  (260, 18, '1802', 'Arajuno'),
  (261, 18, '1803', 'Mera'),
  (262, 18, '1804', 'Santa Clara'),
-- Pichincha (19)
  (263, 19, '1901', 'Quito'),
  (264, 19, '1902', 'Cayambe'),
  (265, 19, '1903', 'Mejía'),
  (266, 19, '1904', 'Pedro Moncayo'),
  (267, 19, '1905', 'Pedro Vicente Maldonado'),
  (268, 19, '1906', 'Puerto Quito'),
  (269, 19, '1907', 'Rumiñahui'),
  (270, 19, '1908', 'San Miguel de los Bancos'),
-- Santa Elena (20)
  (271, 20, '2001', 'Santa Elena'),
  (272, 20, '2002', 'La Libertad'),
-- Santo Domingo de los Tsáchilas (21)
  (273, 21, '2101', 'La Concordia'),
  (274, 21, '2102', 'Santo Domingo'),
-- Sucumbíos (22)
  (275, 22, '2201', 'Cascales'),
  (276, 22, '2202', 'Cuyabeno'),
  (277, 22, '2203', 'Gonzalo Pizarro'),
  (278, 22, '2204', 'Lago Agrio'),
  (279, 22, '2205', 'Putumayo'),
  (280, 22, '2206', 'Shushufindi'),
  (281, 22, '2207', 'Sucumbíos'),
-- Tungurahua (23)
  (282, 23, '2301', 'Ambato'),
  (283, 23, '2302', 'Baños de Agua Santa'),
  (284, 23, '2303', 'Cevallos'),
  (285, 23, '2304', 'Mocha'),
  (286, 23, '2305', 'Patate'),
  (287, 23, '2306', 'Pelileo'),
  (288, 23, '2307', 'Píllaro'),
  (289, 23, '2308', 'Quero'),
  (290, 23, '2309', 'Tisaleo'),
-- Zamora Chinchipe (24)
  (291, 24, '2401', 'Centinela del Cóndor'),
  (292, 24, '2402', 'Chinchipe'),
  (293, 24, '2403', 'El Pangui'),
  (294, 24, '2404', 'Nangaritza'),
  (295, 24, '2405', 'Palanda'),
  (296, 24, '2406', 'Paquisha'),
  (297, 24, '2407', 'Yacuambi'),
  (298, 24, '2408', 'Yantzaza'),
  (299, 24, '2409', 'Zamora');   

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

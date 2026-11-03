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

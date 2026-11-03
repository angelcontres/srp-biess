CREATE TABLE IF NOT EXISTS postulacion_envio_idempotencia (
  idempotency_key varchar(255) PRIMARY KEY,
  postulacion_id   uuid NOT NULL REFERENCES postulacion(id) ON DELETE CASCADE,
  response         jsonb NOT NULL,
  created_at       timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_postulacion_envio_idempotencia_postulacion
  ON postulacion_envio_idempotencia (postulacion_id);

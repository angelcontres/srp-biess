-- Elimina el módulo de pagos y facturación (ADR-014 retirado): la postulación al
-- Directorio del BIESS es gratuita. Ver SRP-010 y SRP-016 en docs/TICKETS.md.
-- Idempotente: se puede ejecutar sobre una base existente o ya migrada.
DROP TABLE IF EXISTS factura;
DROP TABLE IF EXISTS transaccion_pago;
DROP TYPE IF EXISTS estado_factura;
DROP TYPE IF EXISTS estado_pago;
ALTER TABLE IF EXISTS postulacion DROP COLUMN IF EXISTS estado_arancel;

-- PostgreSQL no permite quitar un valor de un ENUM: se recrea tipo_documento sin
-- 'COMPROBANTE_PAGO', solo si todavía lo tiene.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_enum e JOIN pg_type t ON t.oid = e.enumtypid
    WHERE t.typname = 'tipo_documento' AND e.enumlabel = 'COMPROBANTE_PAGO'
  ) THEN
    DELETE FROM documento WHERE tipo = 'COMPROBANTE_PAGO';
    ALTER TYPE tipo_documento RENAME TO tipo_documento_old;
    CREATE TYPE tipo_documento AS ENUM ('RESPALDO_ORGANIZACIONAL', 'TITULO', 'EXPERIENCIA', 'PUBLICACION', 'CAPACITACION', 'FORMULARIO_POSTULACION');
    ALTER TABLE documento ALTER COLUMN tipo TYPE tipo_documento USING tipo::text::tipo_documento;
    DROP TYPE tipo_documento_old;
  END IF;
END $$;

-- SRP-019: la validación del domicilio reconoce a Ecuador por su código ISO 'EC' y no
-- por el id 52 de la semilla. Reemplaza el CHECK ck_domicilio_pais_ec por un trigger.
-- Idempotente. El catálogo de cantones oficial del INEC vive en db/init.sql: en una
-- base existente, vuelva a ejecutar ese script (también es idempotente).
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

-- Agregar campo CC para correos secundarios en notificaciones
ALTER TABLE notificacion ADD COLUMN IF NOT EXISTS cc text[] DEFAULT '{}'::text[];

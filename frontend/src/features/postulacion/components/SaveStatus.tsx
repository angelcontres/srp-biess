import { CircleAlert, CircleCheck, LoaderCircle, TriangleAlert } from 'lucide-react';
import { formatRelativeSave, formatTime } from '@/utils/format';
import type { EstadoGuardado } from '../hooks/useSeccionForm';

export interface SaveStatusProps {
  estado: EstadoGuardado;
  isDirty: boolean;
  ultimoGuardado: string | null;
}

/** Persistent save indicator (SDD 8.6.2, DESIGN save-status.md). */
export default function SaveStatus({ estado, isDirty, ultimoGuardado }: SaveStatusProps) {
  let icon = <CircleCheck size={16} strokeWidth={2} aria-hidden="true" className="shrink-0 text-success" />;
  let text = ultimoGuardado ? `Último guardado: ${formatRelativeSave(ultimoGuardado)}.` : 'Aún no ha guardado esta sección.';

  if (estado === 'saving') {
    icon = <LoaderCircle size={16} strokeWidth={2} aria-hidden="true" className="shrink-0 motion-safe:animate-spin" />;
    text = 'Guardando…';
  } else if (estado === 'error') {
    icon = <CircleAlert size={16} strokeWidth={2} aria-hidden="true" className="shrink-0 text-danger" />;
    text = 'No se guardaron los cambios. Sus datos siguen en pantalla. Reintente.';
  } else if (isDirty) {
    icon = <TriangleAlert size={16} strokeWidth={2} aria-hidden="true" className="shrink-0 text-warning" />;
    text = 'Hay cambios sin guardar.';
  } else if (estado === 'saved' && ultimoGuardado) {
    text = `Progreso guardado a las ${formatTime(ultimoGuardado)}.`;
  }

  return (
    <p role="status" className="flex items-center gap-1.5 text-sm text-muted">
      {icon}
      <span>{text}</span>
    </p>
  );
}

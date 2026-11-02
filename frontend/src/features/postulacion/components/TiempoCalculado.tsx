import { todayIso } from '@/utils/format';
import type { ExperienciaItem } from '../schemas';
import { calcularTiempo, formatTiempo } from '../utils/experiencia';
import type { Periodo } from '../utils/experiencia';
import { dmyToIso } from '../utils/fechas';

/** Period of an item, or null while its dates are incomplete or inconsistent. */
export function periodoDe(item: Pick<ExperienciaItem, 'fechaInicio' | 'fechaFin' | 'actualmente'>): Periodo | null {
  const inicioIso = dmyToIso(item.fechaInicio);
  const finIso = item.actualmente ? todayIso() : dmyToIso(item.fechaFin);
  if (!inicioIso || !finIso || finIso < inicioIso) return null;
  return { inicioIso, finIso };
}

export function tiempoTexto(item: Pick<ExperienciaItem, 'fechaInicio' | 'fechaFin' | 'actualmente'>): string | null {
  const periodo = periodoDe(item);
  const tiempo = periodo ? calcularTiempo(periodo.inicioIso, periodo.finIso) : null;
  return tiempo ? formatTiempo(tiempo) : null;
}

/** Calculated field (SDD 8.6.1): value + "Calculado automáticamente", polite live region. */
export default function TiempoCalculado({ etiqueta, valor, ayuda }: { etiqueta: string; valor: string | null; ayuda?: string }) {
  return (
    <div>
      <p className="text-base font-medium text-neutral-900">{etiqueta}</p>
      <p aria-live="polite" className="mt-1 text-base tabular-nums text-neutral-900">
        {valor ?? 'Se calculará cuando ingrese fechas válidas.'}
      </p>
      <p className="text-sm text-muted">Calculado automáticamente.{ayuda ? ` ${ayuda}` : ''}</p>
    </div>
  );
}

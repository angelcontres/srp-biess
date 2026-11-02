import { daysInMonth } from './fechas';

export interface TiempoExperiencia {
  anios: number;
  meses: number;
  dias: number;
}

interface CalendarDate {
  y: number;
  m: number;
  d: number;
}

function parseIso(iso: string): CalendarDate {
  const [y, m, d] = iso.split('-').map(Number);
  return { y: y ?? 0, m: m ?? 1, d: d ?? 1 };
}

function toUtcMs({ y, m, d }: CalendarDate): number {
  return Date.UTC(y, m - 1, d);
}

/** Adds whole months keeping the start day, clamped to the end of the target month. */
function addMonthsClamped(start: CalendarDate, months: number): CalendarDate {
  const index = start.m - 1 + months;
  const y = start.y + Math.floor(index / 12);
  const m = (index % 12) + 1;
  return { y, m, d: Math.min(start.d, daysInMonth(y, m)) };
}

const MS_PER_DAY = 86_400_000;

/**
 * Calendar difference in years, months and days, anchored on the start day
 * (SRS 7.5). 15/01/2020 -> 10/03/2023 = 3 años, 1 mes, 23 días;
 * 29/02/2020 -> 28/02/2021 = 1 año; 31/01/2023 -> 28/02/2023 = 1 mes.
 * Returns null when the end is before the start.
 */
export function calcularTiempo(inicioIso: string, finIso: string): TiempoExperiencia | null {
  const start = parseIso(inicioIso);
  const end = parseIso(finIso);
  const endMs = toUtcMs(end);
  if (endMs < toUtcMs(start)) return null;

  let months = (end.y - start.y) * 12 + (end.m - start.m);
  while (months > 0 && toUtcMs(addMonthsClamped(start, months)) > endMs) months -= 1;

  const anchor = addMonthsClamped(start, months);
  const dias = Math.round((endMs - toUtcMs(anchor)) / MS_PER_DAY);
  return { anios: Math.floor(months / 12), meses: months % 12, dias };
}

function plural(value: number, singular: string, pluralForm: string): string {
  return `${value} ${value === 1 ? singular : pluralForm}`;
}

/** "3 años, 1 mes y 23 días" */
export function formatTiempo({ anios, meses, dias }: TiempoExperiencia): string {
  return `${plural(anios, 'año', 'años')}, ${plural(meses, 'mes', 'meses')} y ${plural(dias, 'día', 'días')}`;
}

export interface Periodo {
  inicioIso: string;
  finIso: string;
}

function diasEntre({ inicioIso, finIso }: Periodo): number {
  return Math.round((toUtcMs(parseIso(finIso)) - toUtcMs(parseIso(inicioIso))) / MS_PER_DAY);
}

/**
 * Reference total (SRS 7.5): sum of days converted with 365 days/year and
 * 30 days/month, for display only. The server recalculates the official value.
 */
export function totalReferencial(periodos: Periodo[]): TiempoExperiencia {
  let restante = periodos.reduce((sum, periodo) => sum + Math.max(0, diasEntre(periodo)), 0);
  const anios = Math.floor(restante / 365);
  restante -= anios * 365;
  const meses = Math.floor(restante / 30);
  return { anios, meses, dias: restante - meses * 30 };
}

/** True when two periods share at least one day. */
export function haySuperposicion(periodos: Periodo[]): boolean {
  const ordenados = [...periodos].sort((a, b) => a.inicioIso.localeCompare(b.inicioIso));
  let finMaximo = '';
  for (const periodo of ordenados) {
    if (finMaximo && periodo.inicioIso <= finMaximo) return true;
    if (periodo.finIso > finMaximo) finMaximo = periodo.finIso;
  }
  return false;
}

import { formatIsoDate } from '@/utils/format';

/**
 * Dates are typed as "dd/mm/aaaa" text (SDD 8.6.5: no native picker dependency)
 * and exchanged with the API as ISO calendar dates (yyyy-mm-dd).
 */
const DMY = /^(\d{2})\/(\d{2})\/(\d{4})$/;

/** "15/01/2020" -> "2020-01-15"; null when the text is not a real calendar date. */
export function dmyToIso(value: string): string | null {
  const match = DMY.exec(value.trim());
  if (!match) return null;
  const [, dd, mm, yyyy] = match;
  const day = Number(dd);
  const month = Number(mm);
  const year = Number(yyyy);
  if (year < 1900 || month < 1 || month > 12 || day < 1) return null;
  if (day > daysInMonth(year, month)) return null;
  return `${yyyy}-${mm}-${dd}`;
}

/** "2020-01-15" -> "15/01/2020" */
export function isoToDmy(iso: string | null | undefined): string {
  if (!iso) return '';
  const [yyyy, mm, dd] = iso.slice(0, 10).split('-');
  return yyyy && mm && dd ? `${dd}/${mm}/${yyyy}` : '';
}

export function daysInMonth(year: number, month: number): number {
  return new Date(Date.UTC(year, month, 0)).getUTCDate();
}

/** Keeps digits and inserts the slashes while typing; pasting a full date also works. */
export function maskDmy(raw: string): string {
  const digits = raw.replace(/\D/g, '').slice(0, 8);
  if (digits.length <= 2) return digits;
  if (digits.length <= 4) return `${digits.slice(0, 2)}/${digits.slice(2)}`;
  return `${digits.slice(0, 2)}/${digits.slice(2, 4)}/${digits.slice(4)}`;
}

/** "15/01/2020" -> "15 de enero de 2020" (Intl es-EC); the raw text if incomplete. */
export function fechaLegible(dmy: string): string {
  const iso = dmyToIso(dmy);
  return iso ? formatIsoDate(iso) : dmy;
}

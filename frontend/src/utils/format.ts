/** Locale and time zone mandated by DESIGN content.md. */
export const LOCALE = 'es-EC';
export const TIME_ZONE = 'America/Guayaquil';

const dateTimeFormatter = new Intl.DateTimeFormat(LOCALE, {
  timeZone: TIME_ZONE,
  day: 'numeric',
  month: 'long',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
});

const timeFormatter = new Intl.DateTimeFormat(LOCALE, {
  timeZone: TIME_ZONE,
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
});

const dayKeyFormatter = new Intl.DateTimeFormat('en-CA', {
  timeZone: TIME_ZONE,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
});

const longDateFormatter = new Intl.DateTimeFormat(LOCALE, {
  timeZone: 'UTC',
  day: 'numeric',
  month: 'long',
  year: 'numeric',
});

/** "1 de noviembre de 2026, 21:08" */
export function formatDateTime(value: Date | string): string {
  return dateTimeFormatter.format(new Date(value));
}

/** "14:32" */
export function formatTime(value: Date | string): string {
  return timeFormatter.format(new Date(value));
}

/** "hoy, 14:32" or "1 de noviembre de 2026, 14:32" (save indicator). */
export function formatRelativeSave(value: Date | string, now: Date = new Date()): string {
  const date = new Date(value);
  return dayKeyFormatter.format(date) === dayKeyFormatter.format(now)
    ? `hoy, ${formatTime(date)}`
    : formatDateTime(date);
}

/** Formats an ISO calendar date (yyyy-mm-dd) without shifting it by time zone. */
export function formatIsoDate(isoDate: string): string {
  return longDateFormatter.format(new Date(`${isoDate}T00:00:00Z`));
}

/** Today's calendar date in Ecuador as yyyy-mm-dd. */
export function todayIso(now: Date = new Date()): string {
  return dayKeyFormatter.format(now);
}

const sizeFormatter = new Intl.NumberFormat(LOCALE, { maximumFractionDigits: 1 });

/** "3,4 MB" / "850 KB" */
export function formatFileSize(bytes: number): string {
  if (bytes >= 1024 * 1024) return `${sizeFormatter.format(bytes / (1024 * 1024))} MB`;
  return `${sizeFormatter.format(Math.max(1, Math.round(bytes / 1024)))} KB`;
}

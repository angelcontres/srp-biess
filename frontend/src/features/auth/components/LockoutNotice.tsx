import { Clock, TriangleAlert } from 'lucide-react';

export interface LockoutNoticeProps {
  /** Formatted remaining time string (e.g. "09:59"). */
  formattedTime: string;
  /** Whether the block is active. */
  isLocked: boolean;
  /** Optional custom class name. */
  className?: string;
}

/**
 * Reactive lockout component for HTTP 423 responses (EF-11 / SRP-012).
 * Displays a 10-minute live countdown timer with accessible warning indicators.
 */
export default function LockoutNotice({
  formattedTime,
  isLocked,
  className = '',
}: LockoutNoticeProps) {
  if (!isLocked) {
    return null;
  }

  return (
    <div
      role="alert"
      aria-live="assertive"
      className={`rounded-md border border-danger border-l-4 bg-danger-50 p-4 text-neutral-900 ${className}`}
    >
      <div className="flex items-start gap-3">
        <TriangleAlert
          size={20}
          strokeWidth={2}
          aria-hidden="true"
          className="mt-0.5 shrink-0 text-danger"
        />
        <div className="min-w-0 flex-1">
          <h2 className="text-base font-semibold leading-tight text-neutral-900">
            Cuenta bloqueada temporalmente
          </h2>
          <p className="mt-1 text-sm text-neutral-700">
            Ha superado el límite de 3 intentos fallidos permitidos por la Resolución C.D. 701.
            Por motivos de seguridad institucional, los intentos han sido bloqueados durante 10 minutos.
          </p>

          <div className="mt-3 flex items-center gap-2 rounded bg-surface px-3 py-2 border border-line-strong text-sm">
            <Clock size={16} strokeWidth={2} aria-hidden="true" className="shrink-0 text-muted" />
            <span className="text-muted">Tiempo restante para volver a intentar:</span>
            <span className="font-semibold tabular-nums text-danger">{formattedTime}</span>
          </div>
        </div>
      </div>
    </div>
  );
}

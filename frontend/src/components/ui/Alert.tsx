import type { ReactNode } from 'react';
import { CircleAlert, CircleCheck, Info, TriangleAlert, X } from 'lucide-react';

export type AlertVariant = 'info' | 'success' | 'warning' | 'danger';

export interface AlertProps {
  /**
   * `info` institucional (reglas, resultados administrativos, fase cerrada),
   * `danger` error técnico, `warning` advertencia, `success` confirmación.
   */
  variant?: AlertVariant;
  title?: string;
  children?: ReactNode;
  /** Renders a close button when provided. */
  onDismiss?: () => void;
  className?: string;
}

const variantStyles: Record<
  AlertVariant,
  { container: string; icon: string; Icon: typeof Info }
> = {
  info: {
    container: 'border-primary border-l-4 bg-info-50',
    icon: 'text-primary',
    Icon: Info,
  },
  success: {
    container: 'border-success border-l-4 bg-success-50',
    icon: 'text-success',
    Icon: CircleCheck,
  },
  warning: {
    container: 'border-warning border-l-4 bg-warning-50',
    icon: 'text-warning',
    Icon: TriangleAlert,
  },
  danger: {
    container: 'border-danger border-l-4 bg-danger-50',
    icon: 'text-danger',
    Icon: CircleAlert,
  },
};

export default function Alert({
  variant = 'info',
  title,
  children,
  onDismiss,
  className = '',
}: AlertProps) {
  const { container, icon, Icon } = variantStyles[variant];

  return (
    <div
      role="alert"
      className={`flex gap-3 rounded-md border p-4 text-base text-neutral-900 ${container} ${className}`}
    >
      <Icon size={20} strokeWidth={2} aria-hidden="true" className={`shrink-0 ${icon}`} />

      <div className="flex-1 space-y-1">
        {title ? <p className="font-semibold">{title}</p> : null}
        {children ? <div>{children}</div> : null}
      </div>

      {onDismiss ? (
        <button
          type="button"
          onClick={onDismiss}
          aria-label="Cerrar mensaje"
          className="inline-flex h-11 w-11 shrink-0 items-center justify-center self-start rounded-md text-neutral-700 hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
        >
          <X size={20} strokeWidth={2} aria-hidden="true" />
        </button>
      ) : null}
    </div>
  );
}

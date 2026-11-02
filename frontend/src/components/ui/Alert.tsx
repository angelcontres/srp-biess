import type { ReactNode } from 'react';
import { AlertTriangle, CheckCircle2, Info, X, XCircle } from 'lucide-react';

export type AlertVariant = 'success' | 'warning' | 'danger' | 'info';

export interface AlertProps {
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
  success: {
    container: 'border-success/30 bg-success/10 text-neutral-900',
    icon: 'text-success',
    Icon: CheckCircle2,
  },
  warning: {
    container: 'border-warning/30 bg-warning/10 text-neutral-900',
    icon: 'text-warning',
    Icon: AlertTriangle,
  },
  danger: {
    container: 'border-danger/30 bg-danger/10 text-neutral-900',
    icon: 'text-danger',
    Icon: XCircle,
  },
  info: {
    container: 'border-primary/30 bg-primary/10 text-neutral-900',
    icon: 'text-primary',
    Icon: Info,
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
      className={`flex items-start gap-3 rounded-lg border p-4 text-sm ${container} ${className}`}
    >
      <Icon aria-hidden="true" className={`mt-0.5 h-5 w-5 shrink-0 ${icon}`} />

      <div className="flex-1 space-y-1">
        {title ? <p className="font-semibold">{title}</p> : null}
        {children ? <div className="text-neutral-900/80">{children}</div> : null}
      </div>

      {onDismiss ? (
        <button
          type="button"
          onClick={onDismiss}
          aria-label="Cerrar mensaje"
          className="rounded p-1 text-neutral-900/60 transition-colors hover:bg-neutral-900/10 hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
        >
          <X aria-hidden="true" className="h-4 w-4" />
        </button>
      ) : null}
    </div>
  );
}

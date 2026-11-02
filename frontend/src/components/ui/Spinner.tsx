interface SpinnerProps {
  /** Visual size of the loader. */
  size?: 'sm' | 'md' | 'lg';
  /** Stroke colour token; `current` inherits the parent text colour. */
  color?: 'primary' | 'secondary' | 'success' | 'warning' | 'danger' | 'current';
  /** Accessible label announced to screen readers. */
  label?: string;
  className?: string;
}

const sizeClasses: Record<NonNullable<SpinnerProps['size']>, string> = {
  sm: 'h-4 w-4',
  md: 'h-6 w-6',
  lg: 'h-10 w-10',
};

const colorClasses: Record<NonNullable<SpinnerProps['color']>, string> = {
  primary: 'text-primary',
  secondary: 'text-secondary',
  success: 'text-success',
  warning: 'text-warning',
  danger: 'text-danger',
  current: 'text-current',
};

export default function Spinner({
  size = 'md',
  color = 'primary',
  label = 'Cargando',
  className = '',
}: SpinnerProps) {
  return (
    <span role="status" className="inline-flex items-center justify-center align-middle">
      <svg
        className={`animate-spin ${sizeClasses[size]} ${colorClasses[color]} ${className}`}
        viewBox="0 0 24 24"
        fill="none"
        aria-hidden="true"
        focusable="false"
      >
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="3"
        />
        <path
          className="opacity-90"
          fill="currentColor"
          d="M12 2a10 10 0 0 1 10 10h-3a7 7 0 0 0-7-7V2Z"
        />
      </svg>
      <span className="sr-only">{label}</span>
    </span>
  );
}

import { LoaderCircle } from 'lucide-react';

export interface SpinnerProps {
  /** Icon size in px: 16 for captions, 20 is the default (DESIGN §5). */
  size?: 16 | 20 | 32;
  /** Accessible label announced to screen readers. */
  label?: string;
  className?: string;
}

/**
 * Operation loader. DESIGN §7 allows it only inside buttons, always next to text:
 * never as a page-level loading state (use skeletons for layout).
 */
export default function Spinner({ size = 20, label = 'Cargando', className = '' }: SpinnerProps) {
  return (
    <span role="status" className="inline-flex items-center justify-center align-middle">
      <LoaderCircle
        size={size}
        strokeWidth={2}
        aria-hidden="true"
        className={`shrink-0 motion-safe:animate-spin ${className}`}
      />
      <span className="sr-only">{label}</span>
    </span>
  );
}

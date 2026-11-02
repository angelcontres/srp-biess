import { CircleAlert, CircleCheck, Lock, ShieldCheck, TriangleAlert } from 'lucide-react';
import type { ReactNode } from 'react';

export type BadgeVariant = 'complete' | 'incomplete' | 'error' | 'verified' | 'neutral' | 'locked';

export interface BadgeProps {
  variant: BadgeVariant;
  children: ReactNode;
  className?: string;
}

const styles: Record<BadgeVariant, { classes: string; Icon: typeof CircleCheck | null }> = {
  complete: { classes: 'bg-success-50 text-success', Icon: CircleCheck },
  incomplete: { classes: 'bg-warning-50 text-warning', Icon: TriangleAlert },
  error: { classes: 'bg-danger-50 text-danger', Icon: CircleAlert },
  verified: { classes: 'bg-info-50 text-primary', Icon: ShieldCheck },
  neutral: { classes: 'bg-neutral-100 text-neutral-700', Icon: null },
  locked: { classes: 'bg-neutral-100 text-neutral-700', Icon: Lock },
};

/** Status badge: always icon + text, never color alone (DESIGN badge.md). */
export default function Badge({ variant, children, className = '' }: BadgeProps) {
  const { classes, Icon } = styles[variant];
  return (
    <span
      className={`inline-flex items-center gap-1 rounded px-2 py-0.5 text-sm font-medium ${classes} ${className}`}
    >
      {Icon ? <Icon size={16} strokeWidth={2} aria-hidden="true" className="shrink-0" /> : null}
      {children}
    </span>
  );
}

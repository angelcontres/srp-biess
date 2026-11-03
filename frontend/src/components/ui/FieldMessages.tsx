import { CircleAlert } from 'lucide-react';

export interface FieldMessagesProps {
  error?: string;
  errorId: string;
}

/** Inline validation message shared by every field (DESIGN field.md). */
export default function FieldError({ error, errorId }: FieldMessagesProps) {
  if (!error) return null;
  return (
    <p id={errorId} role="alert" className="mt-2 flex gap-2 text-base text-danger">
      <CircleAlert size={20} strokeWidth={2} aria-hidden="true" className="shrink-0" />
      <span>{error}</span>
    </p>
  );
}

export const controlFocusClasses =
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark';

export function controlBorderClasses(hasError: boolean): string {
  return hasError ? 'border-2 border-danger' : 'border border-line-input';
}

export function joinIds(...ids: Array<string | false | null | undefined>): string | undefined {
  return ids.filter(Boolean).join(' ') || undefined;
}

/**
 * Required marker: visible asterisk plus the word for screen readers. SDD 8.6.1 forbids an
 * asterisk without text, so every form that uses it shows <RequiredNote /> as well.
 */
export function RequiredMark() {
  return (
    <>
      <span aria-hidden="true" className="ml-0.5 font-semibold text-neutral-900">
        *
      </span>
      <span className="sr-only"> (obligatorio)</span>
    </>
  );
}

/** Explains the asterisk once per form. */
export function RequiredNote({ className = '' }: { className?: string }) {
  return (
    <p className={`text-sm text-muted ${className}`}>
      Los campos marcados con <span aria-hidden="true">*</span>
      <span className="sr-only">asterisco</span> son obligatorios.
    </p>
  );
}

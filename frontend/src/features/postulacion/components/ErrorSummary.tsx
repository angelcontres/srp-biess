import { useEffect, useRef } from 'react';
import type { ErrorResumen } from '../schemas/resolver';

export interface ErrorSummaryProps {
  errores: ErrorResumen[];
  /** Changes on every failed attempt so the summary takes focus again. */
  intento: number;
}

function focusField(id: string) {
  const target = document.getElementById(id);
  if (!target) return;
  const focusable =
    target.matches('input, select, textarea, button') ? target : target.querySelector<HTMLElement>('input, select, textarea, button');
  target.scrollIntoView({ block: 'center' });
  (focusable ?? target).focus({ preventScroll: true });
}

/** Error summary that receives focus after a failed save (DESIGN error-summary.md). */
export default function ErrorSummary({ errores, intento }: ErrorSummaryProps) {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (errores.length > 0) ref.current?.focus();
  }, [errores, intento]);

  if (errores.length === 0) return null;

  return (
    <div
      ref={ref}
      role="alert"
      tabIndex={-1}
      aria-labelledby="resumen-errores-titulo"
      className="rounded-md border border-danger border-l-4 bg-danger-50 p-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
    >
      <p id="resumen-errores-titulo" className="font-semibold text-neutral-900">
        Revise los siguientes datos:
      </p>
      <ul className="mt-2 list-disc space-y-1 pl-5">
        {errores.map((error) => (
          <li key={`${error.id}-${error.message}`}>
            <a
              href={`#${error.id}`}
              onClick={(event) => {
                event.preventDefault();
                focusField(error.id);
              }}
              className="text-danger hover:text-danger"
            >
              {error.message}
            </a>
          </li>
        ))}
      </ul>
    </div>
  );
}

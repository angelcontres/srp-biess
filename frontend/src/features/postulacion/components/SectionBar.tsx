import { useEffect, useRef } from 'react';
import Button from '@/components/ui/Button';
import { SECCIONES, TOTAL_SECCIONES } from '../constants';
import type { SeccionId } from '../constants';
import { usePostulacion } from '../PostulacionContext';

export interface SectionBarProps {
  seccion: SeccionId;
  /** Hidden in read-only sections (Resumen, blocked states). */
  onGuardarProgreso?: () => void;
  guardando?: boolean;
}

/**
 * Sticky section bar (DESIGN section-bar.md). "Sección N de 10" is the position;
 * the progress line shows sections completed according to the server.
 */
export default function SectionBar({ seccion, onGuardarProgreso, guardando = false }: SectionBarProps) {
  const { postulacion } = usePostulacion();
  const headingRef = useRef<HTMLHeadingElement>(null);
  const indice = SECCIONES.findIndex((item) => item.id === seccion);
  const titulo = SECCIONES[indice]?.titulo ?? '';
  const completas = Object.values(postulacion.secciones).filter((estado) => estado === 'COMPLETA').length;

  useEffect(() => {
    document.title = `Sección ${indice + 1} de ${TOTAL_SECCIONES}: ${titulo} | Postulación BIESS`;
    headingRef.current?.focus();
    window.scrollTo({ top: 0 });
  }, [indice, titulo]);

  return (
    <div className="sticky top-0 z-30 border-b border-line-strong bg-surface">
      <div className="mx-auto flex max-w-5xl items-start justify-between gap-3 px-4 py-3 sm:px-6">
        <div className="min-w-0">
          <p className="text-sm tabular-nums text-muted">
            Sección {indice + 1} de {TOTAL_SECCIONES}
          </p>
          <h1
            ref={headingRef}
            tabIndex={-1}
            className="text-xl font-bold leading-tight text-neutral-900 outline-none sm:text-2xl"
          >
            {titulo}
          </h1>
        </div>
        {onGuardarProgreso ? (
          <Button
            variant="secondary"
            className="shrink-0 whitespace-nowrap"
            onClick={onGuardarProgreso}
            isLoading={guardando}
            loadingText="Guardando…"
          >
            Guardar progreso
          </Button>
        ) : null}
      </div>
      <div
        role="progressbar"
        aria-label="Avance de la postulación"
        aria-valuemin={0}
        aria-valuemax={TOTAL_SECCIONES}
        aria-valuenow={completas}
        aria-valuetext={`${completas} de ${TOTAL_SECCIONES} secciones completas`}
        className="h-1 w-full bg-line"
      >
        <div
          className="h-full bg-primary motion-safe:transition-[width] motion-safe:duration-200"
          style={{ width: `${(completas / TOTAL_SECCIONES) * 100}%` }}
        />
      </div>
    </div>
  );
}

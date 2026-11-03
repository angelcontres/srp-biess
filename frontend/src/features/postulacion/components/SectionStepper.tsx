import { useState } from 'react';
import { ChevronDown, CircleAlert, CircleCheck, Lock } from 'lucide-react';
import { SECCIONES, SECCIONES_POST_CARGO } from '../constants';
import type { SeccionId } from '../constants';
import { usePostulacion } from '../PostulacionContext';

type EstadoPaso = 'completa' | 'actual' | 'pendiente' | 'errores' | 'bloqueada';

function useEstados(): Array<{ id: SeccionId; numero: number; nombre: string; estado: EstadoPaso }> {
  const { postulacion, seccionActual } = usePostulacion();
  const elegible = postulacion.elegibilidad?.elegible === true;
  return SECCIONES.map((seccion, index) => {
    const servidor = postulacion.secciones[seccion.id];
    let estado: EstadoPaso = 'pendiente';
    if (SECCIONES_POST_CARGO.includes(seccion.id) && !elegible) estado = 'bloqueada';
    else if (seccion.id === seccionActual) estado = 'actual';
    else if (servidor === 'COMPLETA') estado = 'completa';
    else if (servidor === 'CON_ERRORES') estado = 'errores';
    return { id: seccion.id, numero: index + 1, nombre: seccion.corto, estado };
  });
}

function Marca({ numero, estado }: { numero: number; estado: EstadoPaso }) {
  if (estado === 'completa') return <CircleCheck size={24} strokeWidth={2} aria-hidden="true" className="shrink-0 text-success" />;
  if (estado === 'errores') return <CircleAlert size={24} strokeWidth={2} aria-hidden="true" className="shrink-0 text-danger" />;
  if (estado === 'bloqueada') return <Lock size={24} strokeWidth={2} aria-hidden="true" className="shrink-0 text-muted" />;
  return (
    <span
      aria-hidden="true"
      className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-sm tabular-nums ${
        estado === 'actual' ? 'bg-primary font-semibold text-white' : 'border border-line-input text-neutral-700'
      }`}
    >
      {numero}
    </span>
  );
}

const SR_TEXT: Partial<Record<EstadoPaso, string>> = {
  completa: ', completada',
  pendiente: ', pendiente',
};

function Lista({ orientacion }: { orientacion: 'vertical' | 'horizontal' }) {
  const { irA } = usePostulacion();
  const pasos = useEstados();
  return (
    <ol className={orientacion === 'vertical' ? 'space-y-1' : 'flex gap-2'}>
      {pasos.map((paso) => {
        const contenido = (
          <>
            <Marca numero={paso.numero} estado={paso.estado} />
            <span className={paso.estado === 'actual' ? 'font-semibold' : undefined}>
              {paso.nombre}
              {paso.estado === 'errores' ? <span className="block text-sm text-danger">Con errores</span> : null}
              {paso.estado === 'bloqueada' ? <span className="block text-sm text-muted">No disponible</span> : null}
              {SR_TEXT[paso.estado] ? <span className="sr-only">{SR_TEXT[paso.estado]}</span> : null}
            </span>
          </>
        );
        const base = 'flex min-h-11 items-center gap-3 rounded-md px-2 py-1 text-left text-base text-neutral-900';
        return (
          <li key={paso.id} className={orientacion === 'horizontal' ? 'shrink-0' : undefined}>
            {paso.estado === 'bloqueada' ? (
              <div className={`${base} text-muted`}>{contenido}</div>
            ) : (
              <button
                type="button"
                onClick={() => irA(paso.id)}
                aria-current={paso.estado === 'actual' ? 'step' : undefined}
                className={`${base} w-full hover:bg-neutral-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark ${
                  paso.estado === 'actual' ? 'bg-info-50' : ''
                }`}
              >
                {contenido}
              </button>
            )}
          </li>
        );
      })}
    </ol>
  );
}

/** Desktop sidebar (≥ 1024 px) and tablet strip (640–1023 px), DESIGN stepper.md. */
export function SectionStepperNav() {
  const { postulacion } = usePostulacion();
  const completas = Object.values(postulacion.secciones).filter((estado) => estado === 'COMPLETA').length;
  return (
    <nav aria-label="Secciones de la postulación" className="hidden sm:block">
      <p className="mb-3 text-sm tabular-nums text-muted">{completas} de 10 secciones completas</p>
      <div className="lg:hidden">
        <div className="overflow-x-auto pb-2">
          <Lista orientacion="horizontal" />
        </div>
      </div>
      <div className="hidden lg:block">
        <Lista orientacion="vertical" />
      </div>
    </nav>
  );
}

/** Mobile (< 640 px): collapsible list opened with "Ver todas las secciones". */
export function SectionStepperMobile() {
  const [abierto, setAbierto] = useState(false);
  return (
    <nav aria-label="Secciones de la postulación" className="sm:hidden">
      <button
        type="button"
        aria-expanded={abierto}
        aria-controls="lista-secciones-movil"
        onClick={() => setAbierto((value) => !value)}
        className="flex min-h-11 items-center gap-2 text-base font-semibold text-primary underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
      >
        Ver todas las secciones
        <ChevronDown size={20} strokeWidth={2} aria-hidden="true" className={abierto ? 'rotate-180' : undefined} />
      </button>
      {abierto ? (
        <div id="lista-secciones-movil" className="mt-2 rounded-lg border border-line-strong bg-surface p-2">
          <Lista orientacion="vertical" />
        </div>
      ) : null}
    </nav>
  );
}

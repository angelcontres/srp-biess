import type { FormEvent, ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Link } from 'react-router-dom';
import { RotateCw } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import Skeleton from '@/components/ui/Skeleton';
import { RUTAS } from '@/features/auth/session/routes';
import { formatDateTime } from '@/utils/format';
import { SECCIONES } from '../constants';
import type { SeccionId } from '../constants';
import type { EstadoGuardado, ErrorServidor } from '../hooks/useSeccionForm';
import { ESCRITORIO, useMediaQuery } from '../hooks/useMediaQuery';
import { useOnlineStatus } from '../hooks/useOnlineStatus';
import { usePostulacion } from '../PostulacionContext';
import type { ErrorResumen } from '../schemas/resolver';
import ErrorSummary from './ErrorSummary';
import SaveStatus from './SaveStatus';
import SectionBar from './SectionBar';
import { RequiredNote } from '@/components/ui/FieldMessages';

export interface SeccionLayoutProps {
  seccion: SeccionId;
  instrucciones?: ReactNode;
  children: ReactNode;
  cargando: boolean;
  mostrarSkeleton: boolean;
  errorCarga: boolean;
  onReintentarCarga: () => void;
  borradorRecuperado: string | null;
  onDescartarBorrador: () => void;
  estado: EstadoGuardado;
  errorServidor: ErrorServidor | null;
  errores: ErrorResumen[];
  intentoResumen: number;
  ultimoGuardado: string | null;
  isDirty: boolean;
  guardarProgreso: () => Promise<boolean>;
  onContinuar: () => void;
  /** Primary action label; "Guardar y continuar" by default. */
  etiquetaPrimaria?: string;
  /** Hides the save actions (blocked or read-only states). */
  soloLectura?: boolean;
}

/**
 * Section structure of the new-form-section skill: sticky bar, backend alert,
 * recovered-draft alert, error summary, instructions, content and footer.
 */
export default function SeccionLayout({
  seccion,
  instrucciones,
  children,
  cargando,
  mostrarSkeleton,
  errorCarga,
  onReintentarCarga,
  borradorRecuperado,
  onDescartarBorrador,
  estado,
  errorServidor,
  errores,
  intentoResumen,
  ultimoGuardado,
  isDirty,
  guardarProgreso,
  onContinuar,
  etiquetaPrimaria = 'Guardar y continuar',
  soloLectura: soloLecturaSeccion = false,
}: SeccionLayoutProps) {
  const { irA, postulacion, slotGuardado } = usePostulacion();
  // >= 1024 px the save status lives under the section list, always in view; below, at the footer.
  // Rendered once so screen readers do not hear two status regions.
  const guardadoLateral = useMediaQuery(ESCRITORIO) && slotGuardado !== null;
  // The server only accepts changes in BORRADOR (409 POS_CERRADA otherwise): NO_ELEGIBLE,
  // ENVIADA, etc. are shown read-only instead of failing on save (SDD 10.11).
  const cerrada = postulacion.estado !== 'BORRADOR';
  const soloLectura = soloLecturaSeccion || cerrada;
  const online = useOnlineStatus();
  const indice = SECCIONES.findIndex((item) => item.id === seccion);
  const anterior = SECCIONES[indice - 1];
  const guardando = estado === 'saving';

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!guardando) onContinuar();
  };

  return (
    <>
      {guardadoLateral && slotGuardado && !soloLectura && !cargando && !errorCarga
        ? createPortal(<SaveStatus estado={estado} isDirty={isDirty} ultimoGuardado={ultimoGuardado} />, slotGuardado)
        : null}
      <SectionBar
        seccion={seccion}
        onGuardarProgreso={soloLectura || cargando || errorCarga ? undefined : () => void guardarProgreso()}
        guardando={guardando}
      />

      <div className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6">
        {!online ? (
          <Alert variant="warning" title="Sin conexión.">
            Sus cambios no se pueden guardar hasta que se restablezca.
          </Alert>
        ) : null}

        {cerrada ? (
          <Alert variant="info" title="Su postulación ya no admite cambios.">
            {postulacion.estado === 'NO_ELEGIBLE' ? (
              <p>
                No cumple los requisitos de elegibilidad para el cargo seleccionado. Puede consultar el resultado en{' '}
                <Link
                  to={RUTAS.miPostulacion}
                  className="text-primary underline underline-offset-4 hover:text-primary-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
                >
                  Mi postulación
                </Link>.
              </p>
            ) : (
              <p>La información se muestra solo para consulta.</p>
            )}
          </Alert>
        ) : null}

        {errorServidor ? (
          <Alert variant="danger" title={errorServidor.titulo}>
            <p>{errorServidor.detalle}</p>
            {errorServidor.referencia ? (
              <p className="mt-1 text-sm tabular-nums text-muted">Código de referencia: {errorServidor.referencia}</p>
            ) : null}
          </Alert>
        ) : null}

        {borradorRecuperado ? (
          <Alert variant="info" title="Se recuperaron datos no guardados.">
            <p>
              Se recuperaron datos no guardados del {formatDateTime(borradorRecuperado)}. Revíselos y presione
              “Guardar progreso”.
            </p>
            <Button variant="tertiary" className="mt-2 min-h-11" onClick={onDescartarBorrador}>
              Descartar borrador
            </Button>
          </Alert>
        ) : null}

        <ErrorSummary errores={errores} intento={intentoResumen} />

        {errorCarga ? (
          <Alert variant="danger" title="No se pudo cargar la sección.">
            <p>No se pudo conectar con el sistema. Reintente en unos minutos.</p>
            <Button
              variant="secondary"
              className="mt-3"
              onClick={onReintentarCarga}
              leftIcon={<RotateCw size={20} strokeWidth={2} aria-hidden="true" />}
            >
              Reintentar
            </Button>
          </Alert>
        ) : cargando ? (
          mostrarSkeleton ? <Skeleton /> : <div aria-busy="true" className="min-h-48" />
        ) : (
          <form noValidate onSubmit={handleSubmit} className="space-y-8">
            {instrucciones ? <div className="max-w-prose text-base text-neutral-700">{instrucciones}</div> : null}
            {soloLectura ? null : <RequiredNote />}

            {/* A disabled fieldset disables every control of the section in one place. */}
            <fieldset disabled={cerrada} className="min-w-0 space-y-8">
              {children}
            </fieldset>

            {soloLectura ? null : (
              <div className="space-y-4 border-t border-line pt-6">
                {guardadoLateral ? null : <SaveStatus estado={estado} isDirty={isDirty} ultimoGuardado={ultimoGuardado} />}
                <div className="flex flex-col gap-3 sm:flex-row-reverse sm:items-center sm:justify-between">
                  <Button type="submit" isLoading={guardando} loadingText="Guardando…" className="w-full sm:w-auto">
                    {etiquetaPrimaria}
                  </Button>
                  {anterior ? (
                    <Button variant="tertiary" className="min-h-11" onClick={() => irA(anterior.id)}>
                      Volver a la sección anterior
                    </Button>
                  ) : null}
                </div>
              </div>
            )}
          </form>
        )}
      </div>
    </>
  );
}

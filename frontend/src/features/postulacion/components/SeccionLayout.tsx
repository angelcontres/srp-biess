import type { FormEvent, ReactNode } from 'react';
import { RotateCw } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import Skeleton from '@/components/ui/Skeleton';
import { formatDateTime } from '@/utils/format';
import { SECCIONES } from '../constants';
import type { SeccionId } from '../constants';
import type { EstadoGuardado, ErrorServidor } from '../hooks/useSeccionForm';
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
  soloLectura = false,
}: SeccionLayoutProps) {
  const { irA } = usePostulacion();
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

            {children}

            {soloLectura ? null : (
              <div className="space-y-4 border-t border-line pt-6">
                <SaveStatus estado={estado} isDirty={isDirty} ultimoGuardado={ultimoGuardado} />
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

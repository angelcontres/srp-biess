import { useEffect, useRef, useState } from 'react';
import { Link } from 'react-router-dom';
import { CircleAlert, Copy, Download, Info } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import { RUTAS } from '@/features/auth/session/routes';
import { formatDateTime } from '@/utils/format';
import { CARGOS } from '../constants';
import { guardarArchivo } from '../utils/descarga';
import { usePostulacion } from '../PostulacionContext';
import type { Elegibilidad, EnvioResultado } from '../types';

function usePageHeading(titulo: string) {
  const ref = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    document.title = `${titulo} | Postulación BIESS`;
    ref.current?.focus();
    window.scrollTo({ top: 0 });
  }, [titulo]);
  return ref;
}

/**
 * Eligibility result: its own screen, institutional style, never `danger`
 * (SDD 8.5.6, DESIGN prohibited.md #17).
 */
export function ElegibilidadView({ elegibilidad }: { elegibilidad: Elegibilidad }) {
  const { irA } = usePostulacion();
  const titulo = elegibilidad.elegible
    ? 'Cumple los requisitos de elegibilidad'
    : 'No cumple los requisitos de elegibilidad para este cargo';
  const headingRef = usePageHeading(titulo);
  const cargo = CARGOS.find((item) => item.id === elegibilidad.cargoId);

  return (
    <div className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6">
      <h1 ref={headingRef} tabIndex={-1} className="text-xl font-bold leading-tight text-neutral-900 outline-none sm:text-2xl">
        {titulo}
      </h1>
      <div className="flex gap-3 rounded-md border border-primary border-l-4 bg-info-50 p-4">
        <Info size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-primary" />
        <div className="max-w-prose space-y-2 text-base text-neutral-900">
          <p className="font-semibold">Resultado de la validación de elegibilidad</p>
          {cargo ? <p>Cargo: {cargo.nombre}.</p> : null}
          <p className="font-semibold">{elegibilidad.mensaje}</p>
          <p className="text-sm tabular-nums text-muted">Fecha y hora del resultado: {formatDateTime(elegibilidad.fecha)} (hora de Ecuador).</p>
        </div>
      </div>
      {elegibilidad.elegible ? (
        <div className="flex flex-col gap-3 sm:flex-row-reverse sm:justify-between">
          <Button className="w-full sm:w-auto" onClick={() => irA('cargo')}>
            Continuar con el respaldo organizacional
          </Button>
        </div>
      ) : (
        <p className="max-w-prose text-base text-neutral-700">
          Este resultado queda registrado en su postulación. Puede volver a consultarlo en{' '}
          <Link
            to={RUTAS.miPostulacion}
            className="text-primary underline underline-offset-4 hover:text-primary-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
          >
            Mi postulación
          </Link>.
        </p>
      )}
    </div>
  );
}

/** Confirmation after sending (SDD 8.5.9). */
export function EnviadaView({ resultado, cargoNombre }: { resultado: EnvioResultado; cargoNombre?: string }) {
  const headingRef = usePageHeading('Postulación enviada');
  const [copiado, setCopiado] = useState(false);
  const { gateway, postulacion } = usePostulacion();
  const [descargando, setDescargando] = useState(false);
  const [errorDescarga, setErrorDescarga] = useState(false);

  // The PDF requires the JWT (SDD 7.5.8), so it is fetched through the API client, not an <a href>.
  const descargar = async () => {
    setErrorDescarga(false);
    setDescargando(true);
    try {
      guardarArchivo(await gateway.descargarFormulario(postulacion.id), `Postulacion-BIESS-${resultado.codigoUnico}.pdf`);
    } catch {
      setErrorDescarga(true);
    } finally {
      setDescargando(false);
    }
  };

  return (
    <div className="mx-auto max-w-5xl space-y-8 px-4 py-8 sm:px-6">
      <h1 ref={headingRef} tabIndex={-1} className="text-xl font-bold leading-tight text-neutral-900 outline-none sm:text-2xl">
        Postulación enviada
      </h1>
      <div className="space-y-4 rounded-lg border border-line-strong bg-surface p-4 sm:p-6">
        <div>
          <p className="text-sm text-muted">Código de su postulación</p>
          <p className="text-2xl font-bold tabular-nums text-neutral-900">{resultado.codigoUnico}</p>
          <div className="mt-2 flex flex-wrap items-center gap-3">
            <Button
              variant="secondary"
              leftIcon={<Copy size={20} strokeWidth={2} aria-hidden="true" />}
              onClick={() => {
                void navigator.clipboard?.writeText(resultado.codigoUnico).then(() => setCopiado(true));
              }}
            >
              Copiar código
            </Button>
            <p role="status" className="text-sm text-muted">
              {copiado ? 'Código copiado.' : ''}
            </p>
          </div>
        </div>
        <dl className="grid gap-3 sm:grid-cols-2">
          <div>
            <dt className="text-sm text-muted">Fecha y hora de envío</dt>
            <dd className="tabular-nums">{formatDateTime(resultado.fechaEnvio)} (hora de Ecuador)</dd>
          </div>
          <div>
            <dt className="text-sm text-muted">Estado</dt>
            <dd>Enviada</dd>
          </div>
          {cargoNombre ? (
            <div className="sm:col-span-2">
              <dt className="text-sm text-muted">Cargo al que postula</dt>
              <dd>{cargoNombre}</dd>
            </div>
          ) : null}
        </dl>
        <div className="space-y-2">
          <Button
            onClick={() => void descargar()}
            isLoading={descargando}
            loadingText="Descargando"
            leftIcon={<Download size={20} strokeWidth={2} aria-hidden="true" />}
          >
            Descargar PDF
          </Button>
          {errorDescarga ? (
            <p role="alert" className="flex gap-2 text-base text-neutral-900">
              <CircleAlert size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-danger" />
              No se pudo descargar el formulario. Reintente.
            </p>
          ) : null}
        </div>
      </div>
      <Alert variant="info" title="Descargue, imprima y firme su formulario.">
        El plazo y lugar de entrega lo establecerá el Instituto Ecuatoriano de Seguridad Social (IESS)
      </Alert>
    </div>
  );
}

import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { CircleAlert, Download, Info, RotateCw } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import Skeleton from '@/components/ui/Skeleton';
import { RUTAS } from '@/features/auth/session/routes';
import { formatDateTime } from '@/utils/format';
import { CARGOS } from '../constants';
import { postulacionGateway } from '../services';
import type { PostulacionGateway } from '../services';
import type { EstadoPostulacion, EventoHistorial, PostulacionEstado } from '../types';
import { guardarArchivo } from '../utils/descarga';

/**
 * State labels of the SDD state model (SDD 10.11). Only BORRADOR has a literal text in
 * SDD 8.5.10; the explanatory texts of the other states are still pending.
 */
const ESTADOS: Record<EstadoPostulacion, { etiqueta: string; explicacion: string }> = {
  BORRADOR: {
    etiqueta: 'Borrador — no enviada',
    explicacion: 'Su postulación todavía no se ha enviado. Puede continuar completándola.',
  },
  LISTA_PARA_ENVIO: {
    etiqueta: 'Lista para envío',
    explicacion: '[PENDIENTE: texto explicativo del estado y próximo paso, según el modelo de estados]',
  },
  ENVIADA: {
    etiqueta: 'Enviada',
    explicacion: '[PENDIENTE: texto explicativo del estado y próximo paso, según el modelo de estados]',
  },
  EN_REVISION: {
    etiqueta: 'En revisión',
    explicacion: '[PENDIENTE: texto explicativo del estado y próximo paso, según el modelo de estados]',
  },
  NO_ELEGIBLE: {
    etiqueta: 'No elegible',
    explicacion:
      'No cumple los requisitos de elegibilidad para el cargo seleccionado. La postulación no puede continuar.',
  },
  CERRADA: {
    etiqueta: 'Cerrada',
    explicacion: '[PENDIENTE: texto explicativo del estado y próximo paso, según el modelo de estados]',
  },
};

const ENVIADAS: EstadoPostulacion[] = ['ENVIADA', 'EN_REVISION', 'CERRADA'];

export interface MiPostulacionPageProps {
  gateway?: PostulacionGateway;
}

/** "Mi postulación" (SDD 8.5.10): state, position, eligibility result and, once sent, code and PDF. */
export default function MiPostulacionPage({ gateway = postulacionGateway }: MiPostulacionPageProps) {
  const navigate = useNavigate();
  const [postulacion, setPostulacion] = useState<PostulacionEstado | null>(null);
  const [errorCarga, setErrorCarga] = useState(false);
  const [mostrarSkeleton, setMostrarSkeleton] = useState(false);

  const [descargando, setDescargando] = useState(false);
  const [errorDescarga, setErrorDescarga] = useState(false);
  // null: loading; 'error': could not be loaded.
  const [historial, setHistorial] = useState<EventoHistorial[] | 'error' | null>(null);

  // This page is its own route (outside PostulacionContext): it uses the `gateway` prop.
  // The PDF requires the JWT (SDD 7.5.8), so it is fetched through the API client, not an <a href>.
  const descargar = async () => {
    if (!postulacion) return;
    setErrorDescarga(false);
    setDescargando(true);
    try {
      const nombre = postulacion.codigoUnico
        ? `Postulacion-BIESS-${postulacion.codigoUnico}.pdf`
        : 'Postulacion-BIESS.pdf';
      guardarArchivo(await gateway.descargarFormulario(postulacion.id), nombre);
    } catch {
      setErrorDescarga(true);
    } finally {
      setDescargando(false);
    }
  };

  useEffect(() => {
    document.title = 'Mi postulación | Postulación BIESS';
  }, []);

  const cargar = useCallback(async () => {
    setErrorCarga(false);
    const timer = setTimeout(() => setMostrarSkeleton(true), 1000);
    try {
      const { postulacion: actual } = await gateway.cargar();
      setPostulacion(actual);
      setHistorial(null);
      gateway
        .historial(actual.id)
        .then(setHistorial)
        .catch(() => setHistorial('error'));
    } catch {
      setErrorCarga(true);
    } finally {
      clearTimeout(timer);
      setMostrarSkeleton(false);
    }
  }, [gateway]);

  useEffect(() => {
    void cargar();
  }, [cargar]);

  const estado = postulacion ? ESTADOS[postulacion.estado] : null;
  const elegibilidad = postulacion?.elegibilidad ?? null;
  const cargo = CARGOS.find((item) => item.id === elegibilidad?.cargoId);
  const enviada = postulacion ? ENVIADAS.includes(postulacion.estado) : false;

  return (
    <div className="mx-auto max-w-5xl space-y-12 px-4 py-8 sm:px-6">
      <h1 className="text-xl font-bold leading-tight text-neutral-900 sm:text-2xl">Mi postulación</h1>

      {errorCarga ? (
        <Alert variant="danger" title="No se pudo cargar su postulación.">
          <p>No se pudo conectar con el sistema. Reintente en unos minutos.</p>
          <Button
            variant="secondary"
            className="mt-3"
            onClick={() => void cargar()}
            leftIcon={<RotateCw size={20} strokeWidth={2} aria-hidden="true" />}
          >
            Reintentar
          </Button>
        </Alert>
      ) : !postulacion || !estado ? (
        <div aria-busy="true" className="min-h-48">
          {mostrarSkeleton ? <Skeleton label="Cargando su postulación" /> : null}
        </div>
      ) : (
        <>
          <section
            aria-labelledby="estado-actual"
            className="space-y-4 rounded-lg border border-line-strong bg-surface p-4 sm:p-6"
          >
            <h2 id="estado-actual" className="text-lg font-semibold text-neutral-900">
              Estado actual
            </h2>
            <dl className="grid gap-4 sm:grid-cols-2">
              <div>
                <dt className="text-sm text-muted">Estado</dt>
                <dd className="text-base font-semibold text-neutral-900">{estado.etiqueta}</dd>
                <dd className="mt-1 max-w-prose text-base text-neutral-700">{estado.explicacion}</dd>
              </div>
              <div>
                <dt className="text-sm text-muted">Cargo al que postula</dt>
                <dd className="text-base text-neutral-900">
                  {cargo ? cargo.nombre : 'Todavía no ha seleccionado un cargo.'}
                </dd>
              </div>
              <div>
                <dt className="text-sm text-muted">Código de su postulación</dt>
                <dd className="text-base text-neutral-900">
                  {enviada && postulacion.codigoUnico ? (
                    <span className="font-semibold tabular-nums">{postulacion.codigoUnico}</span>
                  ) : (
                    'Se asigna al enviar la postulación.'
                  )}
                </dd>
              </div>
              <div>
                <dt className="text-sm text-muted">{enviada ? 'Fecha y hora de envío' : 'Último guardado'}</dt>
                <dd className="text-base tabular-nums text-neutral-900">
                  {enviada
                    ? postulacion.fechaEnvio
                      ? `${formatDateTime(postulacion.fechaEnvio)} (hora de Ecuador)`
                      : 'No disponible.'
                    : postulacion.ultimoGuardado
                      ? `${formatDateTime(postulacion.ultimoGuardado)} (hora de Ecuador)`
                      : 'Todavía no ha guardado información.'}
                </dd>
              </div>
            </dl>
            {postulacion.estado === 'BORRADOR' ? (
              <Button className="w-full sm:w-auto" onClick={() => navigate(RUTAS.formulario)}>
                Continuar con su postulación
              </Button>
            ) : null}
          </section>

          {elegibilidad ? (
            <section aria-labelledby="resultado-elegibilidad" className="space-y-4">
              <h2 id="resultado-elegibilidad" className="text-lg font-semibold text-neutral-900">
                Resultado de la validación de elegibilidad
              </h2>
              {/* Institutional administrative result, never `danger` (SDD 8.5.6, DESIGN prohibited #17). */}
              <div className="flex gap-3 rounded-md border border-primary border-l-4 bg-info-50 p-4">
                <Info size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-primary" />
                <div className="min-w-0 max-w-prose space-y-2 text-base text-neutral-900">
                  <p className="font-semibold">
                    {elegibilidad.elegible
                      ? 'Cumple los requisitos de elegibilidad'
                      : 'No cumple los requisitos de elegibilidad para este cargo'}
                  </p>
                  <p className="font-semibold">{elegibilidad.mensaje}</p>
                  <p className="text-sm tabular-nums text-muted">
                    Fecha y hora del resultado: {formatDateTime(elegibilidad.fecha)} (hora de Ecuador).
                  </p>
                </div>
              </div>
            </section>
          ) : null}

          <section aria-labelledby="historial" className="space-y-4">
            <h2 id="historial" className="text-lg font-semibold text-neutral-900">
              Historial de estados
            </h2>
            {historial === null ? (
              <Skeleton fields={2} label="Cargando el historial" />
            ) : historial === 'error' ? (
              <p role="alert" className="flex gap-2 text-base text-neutral-900">
                <CircleAlert size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-danger" />
                No se pudo cargar el historial. Recargue la página para intentarlo de nuevo.
              </p>
            ) : (
              // SDD 8.5.10: newest first, with date and time.
              <ol className="divide-y divide-line rounded-lg border border-line-strong bg-surface">
                {historial.map((evento, index) => (
                  <li
                    key={`${evento.estado}-${evento.fecha}-${index}`}
                    className="flex flex-wrap justify-between gap-2 p-4"
                  >
                    <span className="text-base font-semibold text-neutral-900">
                      {ESTADOS[evento.estado]?.etiqueta ?? evento.estado}
                    </span>
                    <time dateTime={evento.fecha} className="text-base tabular-nums text-neutral-700">
                      {formatDateTime(evento.fecha)} (hora de Ecuador)
                    </time>
                  </li>
                ))}
              </ol>
            )}
          </section>

          <section aria-labelledby="documentos" className="space-y-4">
            <h2 id="documentos" className="text-lg font-semibold text-neutral-900">
              Formulario y comprobante
            </h2>
            {enviada ? (
              <div className="space-y-2">
                <p className="max-w-prose text-base text-neutral-700">Descargue, imprima y firme su formulario.</p>
                {/* The only primary action once sent (SDD 8.5.10: "Descargar"). */}
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
            ) : (
              <p className="max-w-prose text-base text-neutral-700">
                El formulario en PDF se genera al enviar la postulación.
              </p>
            )}
          </section>

          {postulacion.estado !== 'BORRADOR' ? (
            <p className="text-base">
              <Link
                to={RUTAS.formulario}
                className="inline-flex min-h-11 items-center text-primary underline underline-offset-4 hover:text-primary-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
              >
                Consultar la información de su postulación
              </Link>
            </p>
          ) : null}
        </>
      )}
    </div>
  );
}

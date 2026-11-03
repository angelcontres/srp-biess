import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { RotateCw } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import Modal from '@/components/ui/Modal';
import Skeleton from '@/components/ui/Skeleton';
import { CARGOS, SECCIONES_POST_CARGO } from '../constants';
import type { SeccionId } from '../constants';
import { ElegibilidadView, EnviadaView } from '../components/ResultViews';
import { SectionStepperMobile, SectionStepperNav } from '../components/SectionStepper';
import { hashCedula } from '../hooks/useDraftCache';
import { PostulacionContext } from '../PostulacionContext';
import type { PostulacionContextValue, SeccionGuard } from '../PostulacionContext';
import { postulacionGateway } from '../services';
import type { PostulacionGateway } from '../services';
import { SECCION_COMPONENTES } from '../sections';
import type { DatosVerificados, Elegibilidad, EnvioResultado, PostulacionEstado } from '../types';

type Vista =
  | { tipo: 'seccion' }
  | { tipo: 'elegibilidad'; elegibilidad: Elegibilidad }
  | { tipo: 'enviada'; resultado: EnvioResultado };

interface Carga {
  postulacion: PostulacionEstado;
  verificados: DatosVerificados;
  draftNamespace: string;
}

export interface PostulacionPageProps {
  gateway?: PostulacionGateway;
}

/**
 * Multi-section application form (SRP-013, EF-12). Rendered inside <main> of the
 * app shell; the route `/postulacion` and its guard belong to SRP-015.
 */
export default function PostulacionPage({ gateway = postulacionGateway }: PostulacionPageProps) {
  const [carga, setCarga] = useState<Carga | null>(null);
  const [errorCarga, setErrorCarga] = useState(false);
  const [mostrarSkeleton, setMostrarSkeleton] = useState(false);
  const [seccionActual, setSeccionActual] = useState<SeccionId>('personal');
  const [vista, setVista] = useState<Vista>({ tipo: 'seccion' });
  const [pendiente, setPendiente] = useState<SeccionId | null>(null);
  const [guardandoSalida, setGuardandoSalida] = useState(false);
  const guardRef = useRef<SeccionGuard | null>(null);
  const seguirEditandoRef = useRef<HTMLButtonElement>(null);

  const cargar = useCallback(async () => {
    setErrorCarga(false);
    const timer = setTimeout(() => setMostrarSkeleton(true), 1000);
    try {
      const { postulacion, verificados } = await gateway.cargar();
      const draftNamespace = await hashCedula(verificados.cedula);
      const elegible = postulacion.elegibilidad?.elegible === true;
      const reanudar =
        SECCIONES_POST_CARGO.includes(postulacion.seccionReanudacion) && !elegible ? 'cargo' : postulacion.seccionReanudacion;
      setSeccionActual(reanudar);
      setCarga({ postulacion, verificados, draftNamespace });
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

  // Closing or reloading the tab with unsaved changes (SDD 8.3).
  useEffect(() => {
    const handler = (event: BeforeUnloadEvent) => {
      if (guardRef.current?.isDirty()) event.preventDefault();
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, []);

  const navegar = useCallback((seccion: SeccionId) => {
    setPendiente(null);
    setSeccionActual(seccion);
    setVista({ tipo: 'seccion' });
  }, []);

  const contexto = useMemo<PostulacionContextValue | null>(() => {
    if (!carga) return null;
    return {
      gateway,
      postulacion: carga.postulacion,
      actualizarPostulacion: (postulacion) => setCarga((prev) => (prev ? { ...prev, postulacion } : prev)),
      verificados: carga.verificados,
      draftNamespace: carga.draftNamespace,
      seccionActual,
      irA: (seccion) => {
        if (vista.tipo === 'seccion' && seccion === seccionActual) return;
        if (guardRef.current?.isDirty()) setPendiente(seccion);
        else navegar(seccion);
      },
      registrarGuard: (guard) => {
        guardRef.current = guard;
      },
      mostrarElegibilidad: (elegibilidad) => setVista({ tipo: 'elegibilidad', elegibilidad }),
    };
  }, [carga, gateway, navegar, seccionActual, vista.tipo]);

  if (errorCarga) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6">
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
      </div>
    );
  }

  if (!contexto || !carga) {
    return (
      <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6" aria-busy="true">
        {mostrarSkeleton ? <Skeleton label="Cargando su postulación" /> : null}
      </div>
    );
  }

  const Seccion = SECCION_COMPONENTES[seccionActual];
  const cargo = CARGOS.find((item) => item.id === carga.postulacion.elegibilidad?.cargoId);

  return (
    <PostulacionContext.Provider value={contexto}>
      <div className="mx-auto max-w-7xl lg:flex lg:gap-8 lg:px-6">
        <aside className="border-b border-line-strong bg-surface px-4 py-4 sm:px-6 lg:w-64 lg:shrink-0 lg:border-b-0 lg:bg-transparent lg:px-0 lg:py-8">
          <SectionStepperNav />
          <SectionStepperMobile />
        </aside>
        <div className="min-w-0 flex-1">
          {vista.tipo === 'elegibilidad' ? (
            <ElegibilidadView elegibilidad={vista.elegibilidad} />
          ) : vista.tipo === 'enviada' ? (
            <EnviadaView resultado={vista.resultado} cargoNombre={cargo?.nombre} />
          ) : (
            <Seccion key={seccionActual} onEnviada={(resultado) => setVista({ tipo: 'enviada', resultado })} />
          )}
        </div>
      </div>

      <Modal
        open={pendiente !== null}
        onClose={() => setPendiente(null)}
        title="Tiene cambios sin guardar en esta sección."
        initialFocusRef={seguirEditandoRef}
        footer={
          <>
            <Button ref={seguirEditandoRef} variant="tertiary" className="min-h-11" onClick={() => setPendiente(null)}>
              Seguir editando
            </Button>
            <Button
              variant="secondary"
              onClick={() => {
                guardRef.current?.descartar();
                if (pendiente) navegar(pendiente);
              }}
            >
              Salir sin guardar
            </Button>
            <Button
              isLoading={guardandoSalida}
              loadingText="Guardando…"
              onClick={async () => {
                const destino = pendiente;
                setGuardandoSalida(true);
                const ok = (await guardRef.current?.guardar()) ?? true;
                setGuardandoSalida(false);
                if (ok && destino) navegar(destino);
                else setPendiente(null);
              }}
            >
              Guardar y salir
            </Button>
          </>
        }
      >
        <p>Si sale sin guardar, perderá los cambios que no guardó en esta sección.</p>
      </Modal>
    </PostulacionContext.Provider>
  );
}

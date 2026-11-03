import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useForm } from 'react-hook-form';
import type { DefaultValues, FieldErrors } from 'react-hook-form';
import type { ZodTypeAny } from 'zod';
import { normalizeError } from '@/services/api/client';
import { SECCIONES } from '../constants';
import { usePostulacion } from '../PostulacionContext';
import { aplanarErrores, seccionResolver } from '../schemas/resolver';
import type { ErrorResumen, ModoValidacion } from '../schemas/resolver';
import { CONFLICTO_VERSION } from '../services';
import type { SeccionEditable, ValoresSeccion } from '../types';
import { useDraftCache } from './useDraftCache';

export type EstadoGuardado = 'idle' | 'saving' | 'saved' | 'error';

export interface ErrorServidor {
  titulo: string;
  detalle: string;
  referencia?: string;
}

const SKELETON_DELAY_MS = 1000;

function describirError(error: unknown): ErrorServidor {
  const code = (error as { code?: string })?.code;
  const normalized = normalizeError(error);
  if (code === CONFLICTO_VERSION || normalized.code === CONFLICTO_VERSION) {
    return {
      titulo: 'No se pudo guardar.',
      detalle: 'La postulación fue modificada en otra sesión. Recargue la página. Sus datos siguen en pantalla.',
    };
  }
  if (normalized.code === 'NETWORK_ERROR' || !normalized.status) {
    return {
      titulo: 'No se pudo guardar.',
      detalle: 'No se pudo conectar con el sistema. Sus datos siguen en pantalla. Reintente en unos minutos.',
    };
  }
  return {
    titulo: 'No se pudo guardar.',
    detalle: 'Sus datos siguen en pantalla. Vuelva a presionar “Guardar progreso”.',
    referencia: normalized.traceId,
  };
}

/**
 * Wires one section: load from the server, local draft, lenient/strict validation,
 * saving without ever clearing the form on failure, error summary and the
 * unsaved-changes guard (new-form-section skill, DESIGN states.md).
 */
export function useSeccionForm<K extends SeccionEditable>(
  seccion: K,
  schema: ZodTypeAny,
  valoresIniciales: ValoresSeccion[K],
) {
  const ctx = usePostulacion();
  const { gateway, postulacion, actualizarPostulacion, draftNamespace, irA, registrarGuard } = ctx;
  const modoRef = useRef<ModoValidacion>('progreso');
  const resolver = useMemo(() => seccionResolver<ValoresSeccion[K]>(schema, modoRef), [schema]);
  const form = useForm<ValoresSeccion[K]>({
    resolver,
    mode: 'onBlur',
    reValidateMode: 'onBlur',
    defaultValues: valoresIniciales as DefaultValues<ValoresSeccion[K]>,
  });
  const draft = useDraftCache<ValoresSeccion[K]>(draftNamespace, seccion, schema);

  const [cargando, setCargando] = useState(true);
  const [mostrarSkeleton, setMostrarSkeleton] = useState(false);
  const [errorCarga, setErrorCarga] = useState(false);
  const [borradorRecuperado, setBorradorRecuperado] = useState<string | null>(null);
  const [estado, setEstado] = useState<EstadoGuardado>('idle');
  const [errorServidor, setErrorServidor] = useState<ErrorServidor | null>(null);
  const [errores, setErrores] = useState<ErrorResumen[]>([]);
  const [intentoResumen, setIntentoResumen] = useState(0);
  const [ultimoGuardado, setUltimoGuardado] = useState<string | null>(postulacion.ultimoGuardado);
  const { isDirty } = form.formState;

  const cargar = useCallback(async () => {
    setCargando(true);
    setErrorCarga(false);
    const timer = setTimeout(() => setMostrarSkeleton(true), SKELETON_DELAY_MS);
    try {
      const servidor = (await gateway.obtenerSeccion(postulacion.id, seccion)) ?? valoresIniciales;
      form.reset(servidor);
      if (draft.restored) {
        // Server values stay as defaults, so the draft shows as unsaved changes.
        form.reset(draft.restored.values, { keepDefaultValues: true });
        setBorradorRecuperado(draft.restored.savedAt);
      }
    } catch {
      setErrorCarga(true);
    } finally {
      clearTimeout(timer);
      setMostrarSkeleton(false);
      setCargando(false);
    }
    // Loads once per section mount: form, draft and initial values are stable for it.
  }, [seccion, postulacion.id]);

  useEffect(() => {
    void cargar();
  }, [cargar]);

  useEffect(() => {
    if (cargando) return undefined;
    const subscription = form.watch((values) => draft.schedule(values as ValoresSeccion[K]));
    return () => subscription.unsubscribe();
  }, [cargando, form, draft]);

  const versionRef = useRef(postulacion.version);
  versionRef.current = postulacion.version;

  const guardar = useCallback(
    (modo: ModoValidacion) =>
      new Promise<boolean>((resolve) => {
        modoRef.current = modo;
        const onValid = async (values: ValoresSeccion[K]) => {
          modoRef.current = 'progreso';
          setErrores([]);
          setEstado('saving');
          setErrorServidor(null);
          try {
            const completa = schema.safeParse(values).success;
            const actualizada = await gateway.guardarSeccion(postulacion.id, seccion, values, versionRef.current, completa);
            actualizarPostulacion(actualizada);
            form.reset(values);
            draft.clear();
            setBorradorRecuperado(null);
            setUltimoGuardado(actualizada.ultimoGuardado ?? new Date().toISOString());
            setEstado('saved');
            resolve(true);
          } catch (error) {
            // A server error never clears the form (CLAUDE.md, DESIGN states.md).
            setErrorServidor(describirError(error));
            setEstado('error');
            resolve(false);
          }
        };
        const onInvalid = (fieldErrors: FieldErrors<ValoresSeccion[K]>) => {
          modoRef.current = 'progreso';
          setErrores(aplanarErrores(fieldErrors as FieldErrors));
          setIntentoResumen((value) => value + 1);
          resolve(false);
        };
        void form.handleSubmit(onValid, onInvalid)();
      }),
    [actualizarPostulacion, draft, form, gateway, postulacion.id, schema, seccion],
  );

  const guardarProgreso = useCallback(() => guardar('progreso'), [guardar]);

  const indice = SECCIONES.findIndex((item) => item.id === seccion);
  const guardarYContinuar = useCallback(
    async (validacionExtra?: () => ErrorResumen[]) => {
      const extra = validacionExtra?.() ?? [];
      if (extra.length > 0) {
        setErrores(extra);
        setIntentoResumen((value) => value + 1);
        return;
      }
      const ok = await guardar('completo');
      const siguiente = SECCIONES[indice + 1];
      if (ok && siguiente) irA(siguiente.id);
    },
    [guardar, indice, irA],
  );

  // Unsaved-changes guard used by the stepper and the "previous section" link.
  const dirtyRef = useRef(isDirty);
  dirtyRef.current = isDirty;
  useEffect(() => {
    registrarGuard({ isDirty: () => dirtyRef.current, guardar: guardarProgreso, descartar: draft.clear });
    return () => registrarGuard(null);
  }, [draft.clear, guardarProgreso, registrarGuard]);

  const descartarBorrador = useCallback(() => {
    draft.clear();
    form.reset();
    setBorradorRecuperado(null);
  }, [draft, form]);

  /** Props ready for <SeccionLayout {...layout} />. */
  const layout = {
    cargando,
    mostrarSkeleton,
    errorCarga,
    onReintentarCarga: () => void cargar(),
    borradorRecuperado,
    onDescartarBorrador: descartarBorrador,
    estado,
    errorServidor,
    errores,
    intentoResumen,
    ultimoGuardado,
    isDirty,
    guardarProgreso,
  };

  return {
    layout,
    form,
    cargando,
    mostrarSkeleton,
    errorCarga,
    reintentarCarga: cargar,
    borradorRecuperado,
    descartarBorrador,
    estado,
    errorServidor,
    errores,
    intentoResumen,
    ultimoGuardado,
    isDirty,
    guardarProgreso,
    guardarYContinuar,
  };
}

export type SeccionFormApi = ReturnType<typeof useSeccionForm>;

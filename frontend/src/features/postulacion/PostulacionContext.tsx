import { createContext, useContext } from 'react';
import type { SeccionId } from './constants';
import type { PostulacionGateway } from './services';
import type { DatosVerificados, Elegibilidad, PostulacionEstado } from './types';

/** Lets the page ask the current section about unsaved changes before leaving. */
export interface SeccionGuard {
  isDirty: () => boolean;
  /** Saves progress; resolves true when the server accepted it. */
  guardar: () => Promise<boolean>;
  /** Drops the local draft when the user leaves without saving. */
  descartar: () => void;
}

export interface PostulacionContextValue {
  gateway: PostulacionGateway;
  postulacion: PostulacionEstado;
  actualizarPostulacion: (postulacion: PostulacionEstado) => void;
  verificados: DatosVerificados;
  /** Hash of the cédula used in draft keys. */
  draftNamespace: string;
  seccionActual: SeccionId;
  /** Navigates to a section, asking for confirmation when there are unsaved changes. */
  irA: (seccion: SeccionId) => void;
  registrarGuard: (guard: SeccionGuard | null) => void;
  mostrarElegibilidad: (elegibilidad: Elegibilidad) => void;
  /** Slot under the section list (>= 1024 px) where the current section shows its save status. */
  slotGuardado: HTMLElement | null;
}

export const PostulacionContext = createContext<PostulacionContextValue | null>(null);

export function usePostulacion(): PostulacionContextValue {
  const value = useContext(PostulacionContext);
  if (!value) throw new Error('usePostulacion debe usarse dentro de PostulacionPage.');
  return value;
}

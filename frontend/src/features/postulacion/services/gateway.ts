import type { UploadedFileInfo } from '@/components/ui/FileUpload';
import type {
  CatalogoGeografico,
  DatosVerificados,
  Elegibilidad,
  EnvioResultado,
  EventoHistorial,
  PostulacionEstado,
  SeccionEditable,
  TipoDocumento,
  ValoresSeccion,
} from '../types';

/**
 * Port used by the form (SDD 6.1: pages never call the API directly).
 * `httpGateway` follows SDD 7.5; `mockGateway` is a development adapter used
 * while the backend endpoints of SRP-005 are not available.
 */
export interface PostulacionGateway {
  /** GET /postulaciones/actual, or POST /postulaciones when there is none (RF-EF12-01). */
  cargar(): Promise<{ postulacion: PostulacionEstado; verificados: DatosVerificados }>;
  obtenerSeccion<K extends SeccionEditable>(id: string, seccion: K): Promise<ValoresSeccion[K] | null>;
  guardarSeccion<K extends SeccionEditable>(
    id: string,
    seccion: K,
    valores: ValoresSeccion[K],
    version: number,
    /** Result of the strict schema; the server stores it as the section status. */
    completa: boolean,
  ): Promise<PostulacionEstado>;
  /** PUT /postulaciones/{id}/cargo: definitive selection + eligibility check. */
  confirmarCargo(
    id: string,
    cargoId: number,
    version: number,
  ): Promise<{ postulacion: PostulacionEstado; elegibilidad: Elegibilidad }>;
  subirDocumento(
    id: string,
    tipo: TipoDocumento,
    archivo: File,
    onProgress: (percent: number) => void,
    signal: AbortSignal,
  ): Promise<UploadedFileInfo>;
  /** DELETE /postulaciones/{id}/documentos/{docId}: removes metadata and the stored PDF. */
  eliminarDocumento(id: string, docId: string): Promise<void>;
  /**
   * GET /postulaciones/{id}/documentos/{docId}/download?exp&sig: asks for a fresh signed
   * URL first (they expire in 5 minutes) and returns the PDF as a Blob.
   */
  descargarDocumento(id: string, docId: string): Promise<Blob>;
  enviar(id: string, version: number): Promise<EnvioResultado>;
  /** GET /postulaciones/{id}/pdf (SDD 7.5.8): the form generated on submission, with the JWT. */
  descargarFormulario(id: string): Promise<Blob>;
  /** GET /postulaciones/{id}/historial (SDD 8.5.10): state changes, newest first. */
  historial(id: string): Promise<EventoHistorial[]>;
  catalogo: CatalogoGeografico;
}

/** Version conflict (409 POS_CONFLICTO_VERSION), RF-EF12-02. */
export const CONFLICTO_VERSION = 'POS_CONFLICTO_VERSION';

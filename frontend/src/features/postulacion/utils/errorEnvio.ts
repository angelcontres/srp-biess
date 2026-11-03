import axios from 'axios';
import { normalizeError } from '@/services/api/client';

/** What the summary shows when /guardar or /enviar fails (SDD 8.5.8: always say it was not sent). */
export interface ErrorEnvio {
  titulo: string;
  detalle: string;
  /** Items still missing (422 POS_INCOMPLETA from /guardar). */
  faltantes?: string[];
  /** Offer the link to "Mi postulación" (already sent or closed). */
  verMiPostulacion?: boolean;
  /** `info` when nothing went wrong (already sent); `danger` otherwise. */
  variante: 'danger' | 'info';
}

const NO_ENVIADA = 'La postulación no fue enviada.';

export function describirErrorEnvio(error: unknown): ErrorEnvio {
  const problema = normalizeError(error);
  if (!problema.status) {
    return {
      variante: 'danger',
      titulo: NO_ENVIADA,
      detalle: 'No se pudo conectar con el sistema. Su información no se perdió. Reintente en unos minutos.',
    };
  }
  switch (problema.code) {
    case 'POS_INCOMPLETA': {
      const datos = axios.isAxiosError<{ faltantes?: unknown }>(error) ? error.response?.data : undefined;
      const faltantes = Array.isArray(datos?.faltantes)
        ? datos.faltantes.filter((f): f is string => typeof f === 'string')
        : [];
      return {
        variante: 'danger',
        titulo: NO_ENVIADA,
        detalle: 'Su postulación todavía no está completa. Revise lo siguiente:',
        faltantes,
      };
    }
    case 'POS_YA_ENVIADA':
      return {
        variante: 'info',
        titulo: 'Su postulación ya fue enviada.',
        detalle: 'No es necesario enviarla otra vez. Puede consultar su código y descargar su formulario.',
        verMiPostulacion: true,
      };
    case 'POS_CERRADA':
    case 'POS_NO_LISTA_PARA_ENVIO':
      return {
        variante: 'danger',
        titulo: NO_ENVIADA,
        detalle: 'Su postulación ya no admite el envío en su estado actual.',
        verMiPostulacion: true,
      };
    case 'POS_PDF_NO_GENERADO':
      return {
        variante: 'danger',
        titulo: NO_ENVIADA,
        detalle: 'No fue posible generar su formulario. Intente nuevamente.',
      };
    case 'POS_ACCESO_DENEGADO':
      return { variante: 'danger', titulo: NO_ENVIADA, detalle: 'No tiene permiso para enviar esta postulación.' };
    case 'POS_CONFLICTO_VERSION':
      return {
        variante: 'danger',
        titulo: NO_ENVIADA,
        detalle: 'La postulación fue modificada en otra sesión. Recargue la página y vuelva a enviarla.',
      };
    default:
      return {
        variante: 'danger',
        titulo: NO_ENVIADA,
        detalle:
          'Ocurrió un problema en el sistema. Su información no se perdió. Vuelva a presionar “Enviar postulación”. Si persiste, contacte a soporte.',
      };
  }
}

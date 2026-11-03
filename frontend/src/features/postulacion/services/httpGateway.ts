import axios from 'axios';
import { api } from '@/services/api/client';
import { dmyToIso, isoToDmy } from '../utils/fechas';
import type { PostulacionGateway } from './gateway';
import type { DatosVerificados, OpcionPais, PostulacionEstado, SeccionEditable } from '../types';

// Endpoints available (SRP-013 minimal integration): GET /postulaciones/actual,
// POST /postulaciones, GET /postulaciones/{id} and the section saves of SDD 7.3.4.
// TODO(backend: SRP-005): PUT /cargo (eligibility), /documentos and /enviar.
// TODO(backend: SRP-007): catalog endpoints of SDD 7.3.2.

/** API envelope `{ success, data, message }` or a bare body. */
function unwrap<T>(body: unknown): T {
  if (body && typeof body === 'object' && 'data' in body) return (body as { data: T }).data;
  return body as T;
}

const DMY = /^\d{2}\/\d{2}\/\d{4}$/;
const ISO_DATE = /^\d{4}-\d{2}-\d{2}$/;

/** Form dates (dd/mm/aaaa) travel as ISO calendar dates. */
function mapDates(value: unknown, convert: (text: string) => string | null, pattern: RegExp): unknown {
  if (typeof value === 'string') return pattern.test(value) ? convert(value) ?? value : value;
  if (Array.isArray(value)) return value.map((item) => mapDates(item, convert, pattern));
  if (value && typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value).map(([key, item]) => [key, mapDates(item, convert, pattern)]),
    );
  }
  return value;
}

/** Route and verb per section (SDD 7.3.4). Repeatable sections are sent as a whole. */
const RUTAS: Record<SeccionEditable, { method: 'patch' | 'put'; path: string }> = {
  personal: { method: 'patch', path: 'informacion-personal' },
  domicilio: { method: 'put', path: 'domicilio' },
  // Lists are saved as a whole; SDD per-item CRUD remains for SRP-005.
  familiares: { method: 'put', path: 'familiares' },
  cargo: { method: 'put', path: 'respaldo-organizacional' },
  formacion: { method: 'put', path: 'formacion-academica' },
  experiencia: { method: 'put', path: 'experiencia' },
  publicaciones: { method: 'put', path: 'publicaciones' },
  capacitacion: { method: 'put', path: 'capacitaciones' },
  inhabilidades: { method: 'put', path: 'inhabilidades' },
};

interface Detalle {
  postulacion: PostulacionEstado;
  verificados: DatosVerificados;
  secciones?: Record<string, unknown>;
}

/** GET /postulaciones/{id}: state, verified identity and saved section values. */
async function obtenerDetalle(id: string): Promise<Detalle> {
  return unwrap<Detalle>((await api.get(`/postulaciones/${id}`)).data);
}

async function catalogo(path: string): Promise<OpcionPais[]> {
  const { data } = await api.get(path);
  return unwrap<Array<{ id: number | string; nombre: string; codigo?: string }>>(data).map((item) => ({
    value: String(item.id),
    label: item.nombre,
    codigo: item.codigo,
  }));
}

export const httpGateway: PostulacionGateway = {
  async cargar() {
    let postulacion: PostulacionEstado;
    try {
      postulacion = unwrap<PostulacionEstado>((await api.get('/postulaciones/actual')).data);
    } catch (error) {
      if (!axios.isAxiosError(error) || error.response?.status !== 404) throw error;
      postulacion = unwrap<PostulacionEstado>((await api.post('/postulaciones')).data);
    }
    const detalle = await obtenerDetalle(postulacion.id);
    return { postulacion: detalle.postulacion, verificados: detalle.verificados };
  },

  async obtenerSeccion(id, seccion) {
    const valores = (await obtenerDetalle(id)).secciones?.[seccion];
    return valores ? (mapDates(valores, (iso) => isoToDmy(iso) || null, ISO_DATE) as never) : null;
  },

  async guardarSeccion(id, seccion, valores, version, completa) {
    const { method, path } = RUTAS[seccion];
    const payload = { ...(mapDates(valores, dmyToIso, DMY) as object), version, completa };
    return unwrap<PostulacionEstado>((await api[method](`/postulaciones/${id}/${path}`, payload)).data);
  },

  async confirmarCargo(id, cargoId, version) {
    const body = unwrap<{ postulacion: PostulacionEstado; elegibilidad: PostulacionEstado['elegibilidad'] }>(
      (await api.put(`/postulaciones/${id}/cargo`, { cargoId, confirmacion: true, version })).data,
    );
    if (!body.elegibilidad) throw new Error('Respuesta de elegibilidad incompleta.');
    return { postulacion: body.postulacion, elegibilidad: body.elegibilidad };
  },

  async subirDocumento(id, tipo, archivo, onProgress, signal) {
    const form = new FormData();
    form.append('archivo', archivo);
    form.append('tipo', tipo);
    const { data } = await api.post(`/postulaciones/${id}/documentos`, form, {
      signal,
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: (event) => {
        if (event.total) onProgress(Math.round((event.loaded / event.total) * 100));
      },
    });
    const doc = unwrap<{ id: string; nombreOriginal: string; tamanoBytes: number; urlDescarga: string }>(data);
    return {
      id: doc.id,
      name: doc.nombreOriginal,
      sizeBytes: doc.tamanoBytes,
      uploadedAt: new Date().toISOString(),
      url: doc.urlDescarga,
    };
  },

  async enviar(id, version) {
    const { data } = await api.post(
      `/postulaciones/${id}/enviar`,
      { aceptaConsentimiento: true, versionTerminos: '[PENDIENTE: versión de términos]', version },
      { headers: { 'Idempotency-Key': `${id}-${version}` } },
    );
    return unwrap(data);
  },

  catalogo: {
    paises: () => catalogo('/catalogos/paises'),
    provincias: (paisId) => catalogo(`/catalogos/paises/${paisId}/provincias`),
    cantones: (provinciaId) => catalogo(`/catalogos/provincias/${provinciaId}/cantones`),
  },
};

import type { ZodTypeAny } from 'zod';
import type { SelectOption } from '@/components/ui/SelectField';
import { SECCIONES, SECCIONES_POST_CARGO } from '../constants';
import type { SeccionId } from '../constants';
import {
  capacitacionSchema,
  domicilioSchema,
  experienciaSchema,
  familiaresSchema,
  formacionSchema,
  inhabilidadesSchema,
  personalSchema,
  publicacionesSchema,
  respaldoSchema,
} from '../schemas';
import type { PostulacionGateway } from './gateway';
import type { DatosVerificados, Elegibilidad, EstadoSeccion, PostulacionEstado, SeccionEditable, ValoresSeccion } from '../types';

/**
 * DEVELOPMENT ONLY. Simulates the SDD 7.5 contract in sessionStorage so the form can be
 * exercised before SRP-005 exposes the endpoints. Enabled with VITE_POSTULACION_MOCK=true.
 * Deterministic data: the applicant is "afiliado", so only the AFI position is eligible.
 */
const STORAGE_KEY = 'mock:postulacion';
const LATENCIA_MS = 600;

const VERIFICADOS: DatosVerificados = {
  cedula: '1710034065',
  nombres: 'María Fernanda',
  apellidos: 'Pérez Andrade',
  genero: 'Femenino',
  fechaNacimiento: '1980-05-14',
  nacionalidad: 'Ecuatoriana',
  correo: 'maria.perez@example.com',
};

const SCHEMAS: Record<SeccionEditable, ZodTypeAny> = {
  personal: personalSchema(VERIFICADOS.cedula),
  domicilio: domicilioSchema(VERIFICADOS.correo, () => true, (paisId) => paisId === '1'),
  familiares: familiaresSchema,
  cargo: respaldoSchema,
  formacion: formacionSchema,
  experiencia: experienciaSchema,
  publicaciones: publicacionesSchema,
  capacitacion: capacitacionSchema,
  inhabilidades: inhabilidadesSchema,
};

const MENSAJES_NO_ELEGIBLE: Record<number, string> = {
  2: 'Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector de los Jubilados ante el Directorio del BIESS, NO puede continuar, por cuanto no acredita jubilación.',
  3: 'Su postulación para el cargo de Miembro Principal y Alterno en Representación del Sector Empleador ante el Directorio del BIESS, NO puede continuar, por cuanto no es empleador.',
};

interface MockState {
  postulacion: PostulacionEstado;
  valores: Partial<ValoresSeccion>;
}

function estadoInicial(): MockState {
  const secciones = Object.fromEntries(SECCIONES.map(({ id }) => [id, 'SIN_INICIAR'])) as Record<SeccionId, EstadoSeccion>;
  return {
    postulacion: {
      id: '550e8400-e29b-41d4-a716-446655440000',
      estado: 'BORRADOR',
      version: 1,
      secciones,
      ultimoGuardado: null,
      elegibilidad: null,
      seccionReanudacion: 'personal',
    },
    valores: {},
  };
}

function leer(): MockState {
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    return raw ? (JSON.parse(raw) as MockState) : estadoInicial();
  } catch {
    return estadoInicial();
  }
}

function escribir(state: MockState): void {
  try {
    sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } catch {
    // Storage unavailable: the mock keeps working in memory for this call.
  }
}

const esperar = (ms = LATENCIA_MS) => new Promise((resolve) => setTimeout(resolve, ms));

function recalcularResumen(state: MockState): void {
  const { secciones } = state.postulacion;
  const editables = SECCIONES.filter(({ id }) => id !== 'resumen').map(({ id }) => id);
  const completas = editables.every((id) => secciones[id] === 'COMPLETA');
  secciones.resumen = completas ? 'INCOMPLETA' : 'SIN_INICIAR';
}

export const mockGateway: PostulacionGateway = {
  async cargar() {
    await esperar();
    const state = leer();
    escribir(state);
    return { postulacion: state.postulacion, verificados: VERIFICADOS };
  },

  async obtenerSeccion(_id, seccion) {
    await esperar(300);
    return (leer().valores[seccion] ?? null) as never;
  },

  async guardarSeccion(_id, seccion, valores, version, _completa) {
    await esperar();
    const state = leer();
    if (version !== state.postulacion.version) {
      throw Object.assign(new Error('Conflicto de versión'), { code: 'POS_CONFLICTO_VERSION', status: 409 });
    }
    state.valores = { ...state.valores, [seccion]: valores };
    state.postulacion.secciones[seccion] = SCHEMAS[seccion].safeParse(valores).success ? 'COMPLETA' : 'INCOMPLETA';
    state.postulacion.version += 1;
    state.postulacion.ultimoGuardado = new Date().toISOString();
    state.postulacion.seccionReanudacion = seccion;
    recalcularResumen(state);
    escribir(state);
    return state.postulacion;
  },

  async confirmarCargo(_id, cargoId) {
    await esperar(900);
    const state = leer();
    const elegible = cargoId === 1;
    const elegibilidad: Elegibilidad = {
      elegible,
      cargoId,
      mensaje: elegible ? 'Postulante afiliado al IESS.' : MENSAJES_NO_ELEGIBLE[cargoId] ?? '',
      fecha: new Date().toISOString(),
    };
    state.postulacion.elegibilidad = elegibilidad;
    state.postulacion.version += 1;
    if (!elegible) {
      state.postulacion.estado = 'NO_ELEGIBLE';
      SECCIONES_POST_CARGO.forEach((id) => {
        state.postulacion.secciones[id] = 'SIN_INICIAR';
      });
    }
    escribir(state);
    return { postulacion: state.postulacion, elegibilidad };
  },

  async subirDocumento(_id, _tipo, archivo, onProgress, signal) {
    for (let percent = 20; percent <= 100; percent += 20) {
      await esperar(250);
      if (signal.aborted) throw new DOMException('Cancelado', 'AbortError');
      onProgress(percent);
    }
    return {
      id: `doc-${Date.now()}`,
      name: archivo.name,
      sizeBytes: archivo.size,
      uploadedAt: new Date().toISOString(),
      url: URL.createObjectURL(archivo),
    };
  },

  async enviar() {
    await esperar(1200);
    const state = leer();
    state.postulacion.estado = 'ENVIADA';
    state.postulacion.secciones.resumen = 'COMPLETA';
    escribir(state);
    return {
      codigoUnico: 'BIESS-2026-AFI-000123',
      estado: 'ENVIADA',
      fechaEnvio: new Date().toISOString(),
      pdfUrl: '#',
    };
  },

  catalogo: {
    async paises() {
      await esperar(300);
      return PAISES;
    },
    async provincias(paisId) {
      await esperar(300);
      return paisId === '1' ? PROVINCIAS : [];
    },
    async cantones(provinciaId) {
      await esperar(300);
      return CANTONES[provinciaId] ?? [];
    },
  },
};

// Seed of SDD 16.3.
const o = (value: number, label: string): SelectOption => ({ value: String(value), label });
const PAISES = [
  { ...o(1, 'Ecuador'), codigo: 'EC' },
  { ...o(2, 'Colombia'), codigo: 'CO' },
  { ...o(3, 'Perú'), codigo: 'PE' },
  { ...o(4, 'Estados Unidos'), codigo: 'US' },
  { ...o(5, 'España'), codigo: 'ES' },
];
const PROVINCIAS = [
  'Azuay', 'Bolívar', 'Cañar', 'Carchi', 'Cotopaxi', 'Chimborazo', 'El Oro', 'Esmeraldas', 'Guayas',
  'Imbabura', 'Loja', 'Los Ríos', 'Manabí', 'Morona Santiago', 'Napo', 'Pastaza', 'Pichincha',
  'Tungurahua', 'Zamora Chinchipe', 'Galápagos', 'Sucumbíos', 'Orellana',
  'Santo Domingo de los Tsáchilas', 'Santa Elena',
].map((nombre, index) => o(index + 1, nombre));
const CANTONES: Record<string, SelectOption[]> = {
  '1': [o(101, 'Cuenca'), o(102, 'Girón'), o(103, 'Gualaceo')],
  '9': [o(901, 'Guayaquil'), o(902, 'Alfredo Baquerizo Moreno'), o(903, 'Balao')],
  '13': [o(1301, 'Portoviejo'), o(1302, 'Bolívar'), o(1303, 'Chone')],
  '17': [
    o(1701, 'Quito'), o(1702, 'Cayambe'), o(1703, 'Mejía'), o(1704, 'Pedro Moncayo'), o(1705, 'Rumiñahui'),
    o(1706, 'San Miguel de los Bancos'), o(1707, 'Pedro Vicente Maldonado'), o(1708, 'Puerto Quito'),
  ],
};

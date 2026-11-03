import type { SelectOption } from '@/components/ui/SelectField';
import type { SeccionId } from './constants';
import type {
  CapacitacionValues,
  DomicilioValues,
  ExperienciaValues,
  FamiliaresValues,
  FormacionValues,
  InhabilidadesValues,
  PersonalValues,
  PublicacionesValues,
  RespaldoValues,
} from './schemas';

export type EstadoSeccion = 'COMPLETA' | 'INCOMPLETA' | 'CON_ERRORES' | 'SIN_INICIAR';

export type EstadoPostulacion =
  | 'BORRADOR'
  | 'LISTA_PARA_ENVIO'
  | 'ENVIADA'
  | 'EN_REVISION'
  | 'NO_ELEGIBLE'
  | 'CERRADA';

export interface Elegibilidad {
  elegible: boolean;
  cargoId: number;
  /** Institutional message returned by the server (texts are [V], SDD 8.12 #4). */
  mensaje: string;
  fecha: string;
}

/** Server state of the application. The stepper is driven only by this (SDD 8.5.4). */
export interface PostulacionEstado {
  id: string;
  estado: EstadoPostulacion;
  version: number;
  secciones: Record<SeccionId, EstadoSeccion>;
  ultimoGuardado: string | null;
  elegibilidad: Elegibilidad | null;
  seccionReanudacion: SeccionId;
}

/** Data validated with the Registro Civil (RF-EF12-03). Read-only. */
export interface DatosVerificados {
  cedula: string;
  nombres: string;
  apellidos: string;
  genero: string;
  fechaNacimiento: string;
  nacionalidad: string;
  correo: string;
}

/** Values stored per editable section. */
export interface ValoresSeccion {
  personal: PersonalValues;
  domicilio: DomicilioValues;
  familiares: FamiliaresValues;
  cargo: RespaldoValues;
  formacion: FormacionValues;
  experiencia: ExperienciaValues;
  publicaciones: PublicacionesValues;
  capacitacion: CapacitacionValues;
  inhabilidades: InhabilidadesValues;
}

export type SeccionEditable = keyof ValoresSeccion;

export type TipoDocumento = 'RESPALDO_ORGANIZACIONAL' | 'TITULO' | 'EXPERIENCIA' | 'PUBLICACION' | 'CAPACITACION';

export interface EnvioResultado {
  codigoUnico: string;
  estado: EstadoPostulacion;
  fechaEnvio: string;
  pdfUrl: string;
}

/** Country option; `codigo` is ISO 3166-1 alpha-2 (Ecuador is recognized by "EC", not by id). */
export type OpcionPais = SelectOption & { codigo?: string };

export interface CatalogoGeografico {
  paises(): Promise<OpcionPais[]>;
  provincias(paisId: string): Promise<SelectOption[]>;
  cantones(provinciaId: string): Promise<SelectOption[]>;
}

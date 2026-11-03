import type { SelectOption } from '@/components/ui/SelectField';

/** Section order of SRS 3.5.1 / SDD 8.3. Elegibilidad is a result screen, not a section. */
export const SECCIONES = [
  { id: 'personal', titulo: 'Información personal', corto: 'Personal' },
  { id: 'domicilio', titulo: 'Información domiciliaria y de contacto', corto: 'Domicilio' },
  { id: 'familiares', titulo: 'Familiares en el IESS o el BIESS', corto: 'Familiares' },
  { id: 'cargo', titulo: 'Selección del cargo', corto: 'Cargo' },
  { id: 'formacion', titulo: 'Formación académica', corto: 'Formación' },
  { id: 'experiencia', titulo: 'Experiencia profesional', corto: 'Experiencia' },
  { id: 'publicaciones', titulo: 'Publicaciones e investigaciones', corto: 'Publicaciones' },
  { id: 'capacitacion', titulo: 'Capacitación', corto: 'Capacitación' },
  { id: 'inhabilidades', titulo: 'Inhabilidades', corto: 'Inhabilidades' },
  { id: 'resumen', titulo: 'Resumen y envío', corto: 'Resumen' },
] as const;

export type SeccionId = (typeof SECCIONES)[number]['id'];

export const TOTAL_SECCIONES = SECCIONES.length;

/** Sections that require a positive eligibility result (SDD 8.3). */
export const SECCIONES_POST_CARGO: SeccionId[] = [
  'formacion',
  'experiencia',
  'publicaciones',
  'capacitacion',
  'inhabilidades',
  'resumen',
];

/** SRS 7.6: 4 MB for the organizational support, 2 MB for the rest. */
export const MAX_MB_RESPALDO = 4;
export const MAX_MB_DOCUMENTO = 2;

export const SI_NO: SelectOption[] = [
  { value: 'SI', label: 'Sí' },
  { value: 'NO', label: 'No' },
];

const option = (value: string, label: string): SelectOption => ({ value, label });

// Enumerations from SDD 4.3.
export const ESTADO_CIVIL: SelectOption[] = [
  option('SOLTERO', 'Soltero/a'),
  option('CASADO', 'Casado/a'),
  option('UNION_DE_HECHO', 'Unión de hecho'),
  option('DIVORCIADO', 'Divorciado/a'),
  option('VIUDO', 'Viudo/a'),
  option('OTRO', 'Otro'),
];

/** `genero` enum of SDD 4.3, as returned by the API. */
export const GENERO: SelectOption[] = [
  option('MASCULINO', 'Masculino'),
  option('FEMENINO', 'Femenino'),
  option('OTRO', 'Otro'),
  option('PREFIERO_NO_DECIR', 'Prefiero no decirlo'),
];

/** RF-EF12-05: these states require the partner's data. */
export const ESTADOS_CON_PAREJA = ['CASADO', 'UNION_DE_HECHO'];

export const PARENTESCO: SelectOption[] = [
  option('CONVIVIENTE', 'Conviviente'),
  option('CONYUGE', 'Cónyuge'),
  option('CUNADO', 'Cuñado/a'),
  option('HERMANO', 'Hermano/a'),
  option('HIJO_POLITICO', 'Hijo/a político/a'),
  option('HIJO', 'Hijo/a'),
  option('MADRE', 'Madre'),
  option('NIETO', 'Nieto/a'),
  option('NUERA', 'Nuera'),
  option('PADRE', 'Padre'),
  option('PRIMO_HERMANO', 'Primo/a hermano/a'),
  option('SOBRINO', 'Sobrino/a'),
  option('SUEGRO', 'Suegro/a'),
  option('TIO', 'Tío/a'),
  option('YERNO', 'Yerno'),
];

export const INSTITUCION_FAMILIAR: SelectOption[] = [option('IESS', 'IESS'), option('BIESS', 'BIESS')];

export const NIVEL_INSTRUCCION: SelectOption[] = [
  option('TERCER_NIVEL', 'Tercer nivel'),
  option('ESPECIALIDAD', 'Especialidad'),
  option('MAESTRIA', 'Maestría'),
  option('PHD', 'Doctorado (PhD)'),
];

export const TIPO_INSTITUCION: SelectOption[] = [
  option('NACIONAL', 'Nacional'),
  option('EXTRANJERA', 'Extranjera'),
];

export const TIPO_EXPERIENCIA: SelectOption[] = [
  option('GENERAL', 'General'),
  option('DIRECCION', 'De dirección'),
];

export const TIPO_PUBLICACION: SelectOption[] = [
  option('PUBLICACION', 'Publicación'),
  option('TRABAJO_INVESTIGACION', 'Trabajo de investigación'),
];

export const TIPO_EVENTO: SelectOption[] = [
  option('CURSO', 'Curso'),
  option('SEMINARIO', 'Seminario'),
  option('TALLER', 'Taller'),
];

export function labelOf(options: SelectOption[], value: string | undefined | null): string {
  return options.find((item) => item.value === value)?.label ?? '';
}

/** Seed of SDD 16.3 (cargo table). Requirements per position are [V] in the bases. */
export const CARGOS = [
  {
    id: 1,
    codigo: 'AFI',
    nombre: 'Miembro Principal y Alterno en Representación del Sector de Afiliados',
    respaldo: 'una organización de trabajadores',
  },
  {
    id: 2,
    codigo: 'JUB',
    nombre: 'Miembro Principal y Alterno en Representación del Sector de los Jubilados',
    respaldo: 'una organización de jubilados',
  },
  {
    id: 3,
    codigo: 'EMP',
    nombre: 'Miembro Principal y Alterno en Representación del Sector Empleador',
    respaldo: 'una federación de cámaras',
  },
] as const;

export interface InhabilidadPregunta {
  id: number;
  codigo: string;
  texto: string;
  efecto: 'BLOQUEANTE' | 'ALERTA';
  /** Answer that triggers the blocking effect (1.9.1 blocks on "No"). */
  respuestaBloqueante?: 'SI' | 'NO';
  /** 1.9.7 asks for the removal details. */
  pideRemocion?: boolean;
}

/**
 * 13 questions from the seed (SDD 16.3, RF v1.4 1.9.1–1.9.13).
 * The literal legal text is pending legal review (SDD 8.12 #3).
 */
export const INHABILIDADES: InhabilidadPregunta[] = [
  { id: 1, codigo: '1.9.1', texto: '¿Es ecuatoriano/a y está en pleno goce de los derechos de participación política?', efecto: 'BLOQUEANTE', respuestaBloqueante: 'NO' },
  { id: 2, codigo: '1.9.2', texto: '¿Inhabilitado para ejercer el comercio?', efecto: 'ALERTA' },
  { id: 3, codigo: '1.9.3', texto: '¿En mora con el Estado o instituciones bajo control de la Superintendencia de Bancos?', efecto: 'ALERTA' },
  { id: 4, codigo: '1.9.4', texto: '¿Vínculo con sistema financiero, seguros, bolsas, casas de valores o fondos?', efecto: 'ALERTA' },
  { id: 5, codigo: '1.9.5', texto: '¿Funcionario o empleado del IESS-BIESS?', efecto: 'ALERTA' },
  { id: 6, codigo: '1.9.6', texto: '¿Interés propio o representación de terceros en aseguradoras o seguridad social?', efecto: 'ALERTA' },
  { id: 7, codigo: '1.9.7', texto: '¿Removido por Organismo de Control en los últimos 5 años?', efecto: 'ALERTA', pideRemocion: true },
  { id: 8, codigo: '1.9.8', texto: '¿Obligaciones pendientes con el SRI?', efecto: 'ALERTA' },
  { id: 9, codigo: '1.9.9', texto: '¿Castigo de obligaciones por institución financiera en los últimos 5 años?', efecto: 'ALERTA' },
  { id: 10, codigo: '1.9.10', texto: '¿Litigando contra el IESS-BIESS?', efecto: 'ALERTA' },
  { id: 11, codigo: '1.9.11', texto: '¿Procesado, enjuiciado o condenado por corrupción o crimen organizado?', efecto: 'ALERTA' },
  { id: 12, codigo: '1.9.12', texto: '¿Responsabilidades en firme de la Contraloría?', efecto: 'ALERTA' },
  { id: 13, codigo: '1.9.13', texto: '¿Registrado en la base de la UAFE?', efecto: 'ALERTA' },
];

/** SDD 8.5.8: five consents. Texts are pending legal review (SDD 8.12 #2). */
export const CONSENTIMIENTOS = (
  ['consentimiento1', 'consentimiento2', 'consentimiento3', 'consentimiento4', 'consentimiento5'] as const
).map((id, index) => ({ id, texto: `[PENDIENTE: texto jurídico del consentimiento ${index + 1}]` }));

/** ISO 3166-1 alpha-2 code used to recognize Ecuador in the country catalog. */
export const ECUADOR_ISO = 'EC';

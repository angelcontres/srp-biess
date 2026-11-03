import { formatFileSize } from '@/utils/format';
import type { UploadedFileInfo } from '@/components/ui/FileUpload';
import {
  ESTADO_CIVIL,
  INHABILIDADES,
  labelOf,
  NIVEL_INSTRUCCION,
  PARENTESCO,
  TIPO_EVENTO,
  TIPO_EXPERIENCIA,
  TIPO_PUBLICACION,
} from '../constants';
import type { SeccionEditable, ValoresSeccion } from '../types';
import { fechaLegible } from '../utils/fechas';

export type Fila = [etiqueta: string, valor: string];

const siNo = (value: string) => (value === 'SI' ? 'Sí' : value === 'NO' ? 'No' : 'Sin responder');
const doc = (documento: UploadedFileInfo | null | undefined) =>
  documento ? `${documento.name} (${formatFileSize(documento.sizeBytes)})` : 'Sin documento';

/** Read-only rows for the Resumen review (SDD 8.5.8). */
export function filasDeSeccion<K extends SeccionEditable>(seccion: K, valores: ValoresSeccion[K] | null): Fila[] {
  if (!valores) return [['Estado', 'Sin datos guardados']];
  const v = valores as ValoresSeccion[SeccionEditable];
  switch (seccion) {
    case 'personal': {
      const p = v as ValoresSeccion['personal'];
      const filas: Fila[] = [['Estado civil', labelOf(ESTADO_CIVIL, p.estadoCivil) || 'Sin responder']];
      if (p.parejaNombres) filas.push(['Pareja', `${p.parejaNombres}, cédula ${p.parejaCedula}`]);
      return filas;
    }
    case 'domicilio': {
      const d = v as ValoresSeccion['domicilio'];
      return [
        ['Dirección', [d.callePrincipal, d.numero, d.calleSecundaria].filter(Boolean).join(', ')],
        ['Sector y referencia', [d.sector, d.referencia].filter(Boolean).join(', ') || 'Sin datos'],
        ['Teléfono celular', d.telefonoCelular],
        ['Correo secundario', d.correoAlterno || 'No registrado'],
      ];
    }
    case 'familiares': {
      const f = v as ValoresSeccion['familiares'];
      return [
        ['Tiene familiares en el IESS o el BIESS', siNo(f.tieneFamiliares)],
        ...f.familiares.map((item): Fila => [labelOf(PARENTESCO, item.parentesco), `${item.nombresApellidos}, ${item.institucion}`]),
      ];
    }
    case 'cargo': {
      const c = v as ValoresSeccion['cargo'];
      const filas: Fila[] = [['Respaldo organizacional', siNo(c.tieneRespaldo)]];
      if (c.tieneRespaldo === 'SI') filas.push(['Organización', c.nombreOrganizacion], ['Documento', doc(c.documento)]);
      return filas;
    }
    case 'formacion':
      return (v as ValoresSeccion['formacion']).formaciones.map((item): Fila => [
        labelOf(NIVEL_INSTRUCCION, item.nivel),
        `${item.titulo}, ${item.institucion}. Documento: ${doc(item.documento)}`,
      ]);
    case 'experiencia':
      return (v as ValoresSeccion['experiencia']).experiencias.map((item): Fila => [
        `Experiencia ${labelOf(TIPO_EXPERIENCIA, item.tipo).toLowerCase()}`,
        `${item.cargo}, ${item.institucion}, desde el ${fechaLegible(item.fechaInicio)}. Documento: ${doc(item.documento)}`,
      ]);
    case 'publicaciones': {
      const p = v as ValoresSeccion['publicaciones'];
      return [
        ['Tiene publicaciones', siNo(p.tienePublicaciones)],
        ...p.publicaciones.map((item): Fila => [
          labelOf(TIPO_PUBLICACION, item.tipo),
          `${item.titulo}. Documento: ${doc(item.documento)}`,
        ]),
      ];
    }
    case 'capacitacion': {
      const items = (v as ValoresSeccion['capacitacion']).capacitaciones;
      if (items.length === 0) return [['Capacitaciones', 'No registró capacitaciones']];
      return items.map((item): Fila => [
        labelOf(TIPO_EVENTO, item.tipoEvento),
        `${item.nombre}, ${item.horas} horas. Documento: ${doc(item.documento)}`,
      ]);
    }
    case 'inhabilidades':
      return (v as ValoresSeccion['inhabilidades']).respuestas.map((respuesta, index): Fila => [
        `${INHABILIDADES[index]?.codigo ?? ''} ${INHABILIDADES[index]?.texto ?? ''}`,
        siNo(respuesta),
      ]);
    default:
      return [];
  }
}

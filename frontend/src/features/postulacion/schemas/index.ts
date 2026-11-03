import { z } from 'zod';
import { ESTADOS_CON_PAREJA, INHABILIDADES } from '../constants';
import {
  documento,
  fecha,
  MENSAJE_TELEFONO,
  opcional,
  requerido,
  seleccion,
  siNo,
  TELEFONO,
  validarRango,
} from './comunes';

/* 1. Información personal (RF-EF12-03/05). Verified data is not part of the form. */
export const personalSchema = (cedulaPostulante: string) =>
  z
    .object({
      estadoCivil: seleccion('Seleccione su estado civil.'),
      parejaNombres: opcional(),
      parejaCedula: opcional(10),
    })
    .superRefine((values, ctx) => {
      if (!ESTADOS_CON_PAREJA.includes(values.estadoCivil)) return;
      if (!values.parejaNombres) {
        ctx.addIssue({ code: 'custom', path: ['parejaNombres'], message: 'Ingrese los nombres y apellidos de su pareja.' });
      }
      if (!values.parejaCedula) {
        ctx.addIssue({ code: 'custom', path: ['parejaCedula'], message: 'Ingrese la cédula de su pareja.' });
      } else if (!/^\d{10}$/.test(values.parejaCedula)) {
        ctx.addIssue({
          code: 'custom',
          path: ['parejaCedula'],
          message: 'La cédula debe tener 10 números. Revise que no tenga letras ni espacios.',
        });
      } else if (values.parejaCedula === cedulaPostulante) {
        ctx.addIssue({
          code: 'custom',
          path: ['parejaCedula'],
          message: 'La cédula de su pareja debe ser distinta de la suya.',
        });
      }
    });
export type PersonalValues = z.input<ReturnType<typeof personalSchema>>;

/* 2. Domicilio y contacto (RF-EF12-06/07). */
/**
 * `requiereCanton` returns false when the selected province has no cantons in the
 * catalog: RF-EF12-07 asks for a notice in that case, not for blocking the save.
 */
export const domicilioSchema = (
  correoPrincipal: string,
  requiereCanton: () => boolean = () => true,
  /** Ecuador is identified through the catalog (ISO "EC"), not by a fixed id. */
  esEcuador: (paisId: string) => boolean = () => false,
) =>
  z
    .object({
      paisId: seleccion('Seleccione el país de residencia.'),
      provinciaId: z.string(),
      cantonId: z.string(),
      provinciaManual: opcional(120),
      ciudadManual: opcional(120),
      telefonoCelular: requerido('Ingrese su teléfono celular.', 16).regex(TELEFONO, MENSAJE_TELEFONO),
      correoAlterno: z
        .string()
        .trim()
        .refine((value) => value === '' || z.string().email().safeParse(value).success, {
          message: 'Ingrese un correo válido, por ejemplo nombre@dominio.com.',
        })
        .refine((value) => value.toLowerCase() !== correoPrincipal.toLowerCase(), {
          message: 'El correo secundario debe ser distinto del correo principal.',
        }),
      callePrincipal: requerido('Ingrese la calle principal.', 150),
      numero: requerido('Ingrese el número de casa o edificio.', 20),
      calleSecundaria: opcional(150),
      sector: opcional(120),
      referencia: opcional(255),
      telefonoDomicilio: z
        .string()
        .trim()
        .refine((value) => value === '' || TELEFONO.test(value), { message: MENSAJE_TELEFONO }),
    })
    .superRefine((values, ctx) => {
      if (esEcuador(values.paisId)) {
        if (!values.provinciaId) ctx.addIssue({ code: 'custom', path: ['provinciaId'], message: 'Seleccione la provincia.' });
        if (!values.cantonId && requiereCanton()) {
          ctx.addIssue({ code: 'custom', path: ['cantonId'], message: 'Seleccione el cantón.' });
        }
      } else if (values.paisId) {
        if (!values.provinciaManual) ctx.addIssue({ code: 'custom', path: ['provinciaManual'], message: 'Ingrese la provincia o estado.' });
        if (!values.ciudadManual) ctx.addIssue({ code: 'custom', path: ['ciudadManual'], message: 'Ingrese la ciudad.' });
      }
    });
export type DomicilioValues = z.input<ReturnType<typeof domicilioSchema>>;

/* 3. Familiares (RF-EF12-08). */
export const familiarSchema = z.object({
  id: z.string(),
  nombresApellidos: requerido('Ingrese los nombres y apellidos del familiar.', 200),
  parentesco: seleccion('Seleccione el parentesco.'),
  institucion: seleccion('Seleccione la institución.'),
  area: requerido('Ingrese el área en la que trabaja.', 150),
  cargo: requerido('Ingrese el cargo que ocupa.', 150),
});
export type FamiliarValues = z.input<typeof familiarSchema>;

export const familiaresSchema = z
  .object({
    tieneFamiliares: siNo('Indique si tiene familiares que trabajen en el IESS o el BIESS.'),
    familiares: z.array(familiarSchema),
  })
  .superRefine((values, ctx) => {
    if (values.tieneFamiliares === 'SI' && values.familiares.length === 0) {
      ctx.addIssue({ code: 'custom', path: ['familiares'], message: 'Agregue al menos un familiar o responda “No”.' });
    }
  });
export type FamiliaresValues = z.input<typeof familiaresSchema>;

/* 4. Cargo: respaldo organizacional after eligibility (RF-EF12-11). */
export const respaldoSchema = z
  .object({
    tieneRespaldo: siNo('Indique si cuenta con respaldo organizacional.'),
    nombreOrganizacion: opcional(200),
    documento: documento('Adjunte el documento de respaldo en PDF.').or(z.null()),
  })
  .superRefine((values, ctx) => {
    if (values.tieneRespaldo !== 'SI') return;
    if (!values.nombreOrganizacion) {
      ctx.addIssue({ code: 'custom', path: ['nombreOrganizacion'], message: 'Ingrese el nombre de la organización.' });
    }
    if (!values.documento) {
      ctx.addIssue({ code: 'custom', path: ['documento'], message: 'Adjunte el documento de respaldo en PDF.' });
    }
  });
export type RespaldoValues = z.input<typeof respaldoSchema>;

/* 5. Formación académica (RF-EF12-12). */
export const formacionItemSchema = z.object({
  id: z.string(),
  nivel: seleccion('Seleccione el nivel de instrucción.'),
  institucion: requerido('Ingrese la institución educativa.', 200),
  tipoInstitucion: seleccion('Seleccione si la institución es nacional o extranjera.'),
  titulo: requerido('Ingrese el título obtenido.', 200),
  area: requerido('Ingrese el área de estudios.', 150),
  numeroSenescyt: requerido('Ingrese el número de registro de la SENESCYT.', 50),
  fechaSenescyt: fecha('Ingrese la fecha de registro de la SENESCYT.'),
  documento: documento('Adjunte el título en PDF.'),
});
export type FormacionItem = z.input<typeof formacionItemSchema>;

export const formacionSchema = z.object({
  formaciones: z.array(formacionItemSchema).min(1, 'Agregue al menos una formación académica.'),
});
export type FormacionValues = z.input<typeof formacionSchema>;

/* 6. Experiencia profesional (RF-EF12-13). */
export const experienciaItemSchema = z
  .object({
    id: z.string(),
    tipo: seleccion('Seleccione el tipo de experiencia.'),
    cargo: requerido('Ingrese el cargo o puesto.', 150),
    institucion: requerido('Ingrese la institución o empresa.', 200),
    fechaInicio: fecha('Ingrese la fecha de inicio.'),
    actualmente: z.boolean(),
    fechaFin: z.string(),
    actividades: requerido('Ingrese las actividades relacionadas con el cargo.', 500),
    funciones: requerido('Ingrese la descripción de funciones.', 1000),
    documento: documento('Adjunte el certificado de experiencia en PDF.'),
  })
  .superRefine((values, ctx) => {
    if (values.actualmente) return;
    const resultado = fecha('Ingrese la fecha de fin o marque “Trabajo aquí actualmente”.').safeParse(values.fechaFin);
    if (!resultado.success) {
      ctx.addIssue({ code: 'custom', path: ['fechaFin'], message: resultado.error.issues[0]?.message ?? '' });
      return;
    }
    validarRango(values.fechaInicio, values.fechaFin, ctx, ['fechaFin']);
  });
export type ExperienciaItem = z.input<typeof experienciaItemSchema>;

export const experienciaSchema = z.object({ experiencias: z.array(experienciaItemSchema) });
export type ExperienciaValues = z.input<typeof experienciaSchema>;

/* 7. Publicaciones (RF-EF12-14). */
export const publicacionItemSchema = z.object({
  id: z.string(),
  titulo: requerido('Ingrese el título de la publicación.', 300),
  tipo: seleccion('Seleccione el tipo de publicación.'),
  medio: requerido('Ingrese el medio de publicación.', 200),
  fecha: fecha('Ingrese la fecha de publicación.'),
  relacion: requerido('Indique la relación con su formación o ejercicio profesional.', 500),
  documento: documento('Adjunte la publicación en PDF.'),
});
export type PublicacionItem = z.input<typeof publicacionItemSchema>;

export const publicacionesSchema = z
  .object({
    tienePublicaciones: siNo('Indique si cuenta con publicaciones o investigaciones académicas.'),
    publicaciones: z.array(publicacionItemSchema),
  })
  .superRefine((values, ctx) => {
    if (values.tienePublicaciones === 'SI' && values.publicaciones.length === 0) {
      ctx.addIssue({ code: 'custom', path: ['publicaciones'], message: 'Agregue al menos una publicación o responda “No”.' });
    }
  });
export type PublicacionesValues = z.input<typeof publicacionesSchema>;

/* 8. Capacitación (RF-EF12-15): optional section, complete items. */
export const capacitacionItemSchema = z
  .object({
    id: z.string(),
    tipoEvento: seleccion('Seleccione el tipo de evento.'),
    nombre: requerido('Ingrese el nombre del evento.', 200),
    institucion: requerido('Ingrese la institución que organizó el evento.', 200),
    fechaInicio: fecha('Ingrese la fecha de inicio.'),
    fechaFin: fecha('Ingrese la fecha de fin.'),
    horas: requerido('Ingrese el número de horas.', 5).regex(
      /^[1-9]\d*$/,
      'Ingrese un número de horas mayor que cero, sin decimales.',
    ),
    documento: documento('Adjunte el certificado en PDF.'),
  })
  .superRefine((values, ctx) => validarRango(values.fechaInicio, values.fechaFin, ctx, ['fechaFin']));
export type CapacitacionItem = z.input<typeof capacitacionItemSchema>;

export const capacitacionSchema = z.object({ capacitaciones: z.array(capacitacionItemSchema) });
export type CapacitacionValues = z.input<typeof capacitacionSchema>;

/* 9. Inhabilidades (RF-EF12-16, SRS 7.4). */
export const inhabilidadesSchema = z
  .object({
    respuestas: z.array(siNo('Responda esta pregunta.')).length(INHABILIDADES.length),
    remocionInstitucion: opcional(200),
    remocionCargo: opcional(150),
    remocionFechaFin: z.string(),
    remocionMotivo: opcional(500),
  })
  .superRefine((values, ctx) => {
    const indice = INHABILIDADES.findIndex((pregunta) => pregunta.pideRemocion);
    if (values.respuestas[indice] !== 'SI') return;
    if (!values.remocionInstitucion) ctx.addIssue({ code: 'custom', path: ['remocionInstitucion'], message: 'Ingrese la institución.' });
    if (!values.remocionCargo) ctx.addIssue({ code: 'custom', path: ['remocionCargo'], message: 'Ingrese el cargo.' });
    const fin = fecha('Ingrese la fecha de finalización.').safeParse(values.remocionFechaFin);
    if (!fin.success) ctx.addIssue({ code: 'custom', path: ['remocionFechaFin'], message: fin.error.issues[0]?.message ?? '' });
    if (!values.remocionMotivo) ctx.addIssue({ code: 'custom', path: ['remocionMotivo'], message: 'Ingrese el motivo de la remoción.' });
  });
export type InhabilidadesValues = z.input<typeof inhabilidadesSchema>;

/* 10. Resumen: five mandatory consents (RF-EF12-18). */
export const resumenSchema = z.object({
  consentimiento1: z.literal(true, { errorMap: () => ({ message: 'Acepte el consentimiento 1 para enviar su postulación.' }) }),
  consentimiento2: z.literal(true, { errorMap: () => ({ message: 'Acepte el consentimiento 2 para enviar su postulación.' }) }),
  consentimiento3: z.literal(true, { errorMap: () => ({ message: 'Acepte el consentimiento 3 para enviar su postulación.' }) }),
  consentimiento4: z.literal(true, { errorMap: () => ({ message: 'Acepte el consentimiento 4 para enviar su postulación.' }) }),
  consentimiento5: z.literal(true, { errorMap: () => ({ message: 'Acepte el consentimiento 5 para enviar su postulación.' }) }),
});
export type ResumenValues = z.input<typeof resumenSchema>;


export { default as SECCIONES_REGLAS } from './secciones_reglas.json';

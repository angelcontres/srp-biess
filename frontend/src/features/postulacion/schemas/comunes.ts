import { z } from 'zod';
import type { UploadedFileInfo } from '@/components/ui/FileUpload';
import { todayIso } from '@/utils/format';
import { dmyToIso } from '../utils/fechas';

/** Required trimmed text. Messages follow SDD 8.10: "Ingrese [dato]." */
export const requerido = (mensaje: string, max = 255) =>
  z.string().trim().min(1, mensaje).max(max, `Use como máximo ${max} caracteres.`);

export const opcional = (max = 255) => z.string().trim().max(max, `Use como máximo ${max} caracteres.`);

export const seleccion = (mensaje: string) => z.string().min(1, mensaje);

export const siNo = (mensaje: string) =>
  z.enum(['SI', 'NO', '']).refine((value) => value !== '', mensaje);

export const cedula = (mensaje: string) =>
  requerido(mensaje, 10).regex(
    /^\d{10}$/,
    'La cédula debe tener 10 números. Revise que no tenga letras ni espacios.',
  );

/** National (0991234567, 022345678) or international (+593991234567) phone. */
export const TELEFONO = /^(\+\d{8,15}|0\d{8,9})$/;
export const MENSAJE_TELEFONO =
  'Ingrese el teléfono con 9 o 10 números, por ejemplo 0991234567, o con código de país, por ejemplo +593991234567.';

/** dd/mm/aaaa text that must be a real date and, by default, not in the future. */
export const fecha = (mensaje: string, { permitirFutura = false } = {}) =>
  z.string().superRefine((value, ctx) => {
    if (value.trim() === '') {
      ctx.addIssue({ code: z.ZodIssueCode.custom, message: mensaje });
      return;
    }
    const iso = dmyToIso(value);
    if (!iso) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'Ingrese la fecha con el formato dd/mm/aaaa, por ejemplo 15/01/2020.',
      });
      return;
    }
    if (!permitirFutura && iso > todayIso()) {
      ctx.addIssue({ code: z.ZodIssueCode.custom, message: 'La fecha no puede ser posterior a hoy.' });
    }
  });

/** Uploaded PDF reference; the file itself never enters the form state or the draft. */
export const documento = (mensaje = 'Adjunte el documento PDF.') =>
  z
    .custom<UploadedFileInfo | null>(
      (value) => value === null || (typeof value === 'object' && value !== null && 'id' in value),
    )
    .refine((value) => value !== null, mensaje);

/** Adds an issue when end < start (both as valid dd/mm/aaaa). */
export function validarRango(
  inicio: string,
  fin: string,
  ctx: z.RefinementCtx,
  path: (string | number)[],
) {
  const inicioIso = dmyToIso(inicio);
  const finIso = dmyToIso(fin);
  if (inicioIso && finIso && finIso < inicioIso) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      path,
      message: 'La fecha de fin debe ser posterior a la fecha de inicio.',
    });
  }
}

/** Client id for repeatable items before the server assigns one. */
export function nuevoId(): string {
  return typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `tmp-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

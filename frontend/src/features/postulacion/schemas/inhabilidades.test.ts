import { describe, expect, it } from 'vitest';
import { todayIso } from '@/utils/format';
import { INHABILIDADES } from '../constants';
import { isoToDmy, restarAnios } from '../utils/fechas';
import { inhabilidadesSchema } from './index';

const indiceRemocion = INHABILIDADES.findIndex((pregunta) => pregunta.pideRemocion);

function conRemocion(fechaFinIso: string) {
  return {
    respuestas: INHABILIDADES.map((_, i) => (i === indiceRemocion ? 'SI' : 'NO')),
    remocionInstitucion: 'Contraloría General del Estado',
    remocionCargo: 'Director',
    remocionFechaFin: isoToDmy(fechaFinIso),
    remocionMotivo: 'Motivo de prueba',
  };
}

const errorDeFecha = (fechaFinIso: string) =>
  inhabilidadesSchema
    .safeParse(conRemocion(fechaFinIso))
    .error?.issues.find((issue) => issue.path[0] === 'remocionFechaFin')?.message;

describe('Inhabilidades: fecha de finalización de la remoción (1.9.7, últimos 5 años)', () => {
  const hoy = todayIso();
  const limite = restarAnios(hoy, 5);

  it('acepta una fecha dentro de los últimos 5 años y el límite exacto', () => {
    expect(errorDeFecha(restarAnios(hoy, 2))).toBeUndefined();
    expect(errorDeFecha(limite)).toBeUndefined();
  });

  it('rechaza una fecha anterior al rango e indica desde cuándo', () => {
    const mensaje = errorDeFecha(restarAnios(hoy, 6));
    expect(mensaje).toContain('dentro de los últimos 5 años');
    expect(mensaje).toContain(isoToDmy(limite));
  });

  it('sigue rechazando una fecha futura', () => {
    expect(errorDeFecha(`${Number(hoy.slice(0, 4)) + 1}${hoy.slice(4)}`)).toBeDefined();
  });

  it('con respuesta "No" no exige ni valida la fecha', () => {
    const valores = { ...conRemocion(restarAnios(hoy, 10)), respuestas: INHABILIDADES.map(() => 'NO') };
    expect(inhabilidadesSchema.safeParse(valores).success).toBe(true);
  });
});

describe('restarAnios', () => {
  it('resta años y ajusta el 29 de febrero', () => {
    expect(restarAnios('2026-11-03', 5)).toBe('2021-11-03');
    expect(restarAnios('2028-02-29', 5)).toBe('2023-02-28');
    expect(restarAnios('2028-02-29', 4)).toBe('2024-02-29');
  });
});

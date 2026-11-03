import { describe, expect, it } from 'vitest';
import { calcularTiempo, formatTiempo, haySuperposicion, totalReferencial } from './experiencia';

const texto = (inicio: string, fin: string) => {
  const tiempo = calcularTiempo(inicio, fin);
  return tiempo ? formatTiempo(tiempo) : null;
};

describe('calcularTiempo (SRS 7.5)', () => {
  it('calcula años, meses y días anclados al día de inicio', () => {
    expect(texto('2020-01-15', '2023-03-10')).toBe('3 años, 1 mes y 23 días');
  });

  it('ajusta el año bisiesto y el fin de mes', () => {
    expect(texto('2020-02-29', '2021-02-28')).toBe('1 año, 0 meses y 0 días');
    expect(texto('2023-01-31', '2023-02-28')).toBe('0 años, 1 mes y 0 días');
  });

  it('acepta inicio igual a fin y rechaza fin anterior al inicio', () => {
    expect(texto('2023-05-05', '2023-05-05')).toBe('0 años, 0 meses y 0 días');
    expect(texto('2023-05-05', '2023-05-01')).toBeNull();
  });
});

describe('total y superposición', () => {
  it('convierte el total con 365 días por año y 30 por mes', () => {
    expect(totalReferencial([{ inicioIso: '2020-01-01', finIso: '2020-03-01' }])).toEqual({ anios: 0, meses: 2, dias: 0 });
  });

  it('detecta períodos superpuestos aunque no sean consecutivos', () => {
    expect(
      haySuperposicion([
        { inicioIso: '2010-01-01', finIso: '2020-01-01' },
        { inicioIso: '2011-01-01', finIso: '2011-02-01' },
        { inicioIso: '2015-01-01', finIso: '2016-01-01' },
      ]),
    ).toBe(true);
    expect(
      haySuperposicion([
        { inicioIso: '2010-01-01', finIso: '2011-01-01' },
        { inicioIso: '2012-01-01', finIso: '2013-01-01' },
      ]),
    ).toBe(false);
  });
});

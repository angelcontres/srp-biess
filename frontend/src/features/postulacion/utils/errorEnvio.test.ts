import { AxiosError, AxiosHeaders, type AxiosResponse } from 'axios';
import { describe, expect, it } from 'vitest';
import { describirErrorEnvio } from './errorEnvio';

function errorHttp(status: number, data: unknown): AxiosError {
  const config = { headers: new AxiosHeaders() };
  const response = { status, data, statusText: '', headers: {}, config } as AxiosResponse;
  return new AxiosError('fallo', String(status), config, undefined, response);
}

describe('describirErrorEnvio (SDD 8.5.8)', () => {
  it('422 POS_INCOMPLETA lista lo que falta', () => {
    const e = describirErrorEnvio(
      errorHttp(422, { code: 'POS_INCOMPLETA', faltantes: ['Complete la sección Inhabilidades.', 3] }),
    );
    expect(e.titulo).toBe('La postulación no fue enviada.');
    expect(e.faltantes).toEqual(['Complete la sección Inhabilidades.']);
    expect(e.variante).toBe('danger');
  });

  it('409 POS_NO_LISTA_PARA_ENVIO no invita a reintentar y ofrece Mi postulación', () => {
    const e = describirErrorEnvio(errorHttp(409, { code: 'POS_NO_LISTA_PARA_ENVIO' }));
    expect(e.detalle).not.toMatch(/Vuelva a presionar/);
    expect(e.verMiPostulacion).toBe(true);
  });

  it('409 POS_YA_ENVIADA es informativo', () => {
    const e = describirErrorEnvio(errorHttp(409, { code: 'POS_YA_ENVIADA' }));
    expect(e.variante).toBe('info');
    expect(e.titulo).toBe('Su postulación ya fue enviada.');
  });

  it('503 POS_PDF_NO_GENERADO usa el mensaje del SRS', () => {
    expect(describirErrorEnvio(errorHttp(503, { code: 'POS_PDF_NO_GENERADO' })).detalle).toBe(
      'No fue posible generar su formulario. Intente nuevamente.',
    );
  });

  it('sin conexión lo indica y un 500 mantiene el mensaje genérico', () => {
    expect(describirErrorEnvio(new AxiosError('Network Error', 'ERR_NETWORK')).detalle).toMatch(/No se pudo conectar/);
    expect(describirErrorEnvio(errorHttp(500, { code: 'ERROR_INTERNO' })).detalle).toMatch(/Ocurrió un problema en el sistema/);
  });
});

import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { mockGateway } from '../services/mockGateway';
import type { PostulacionGateway } from '../services';
import type { PostulacionEstado } from '../types';
import MiPostulacionPage from './MiPostulacionPage';

const ESPERA = { timeout: 4000 };

function conEstado(cambios: Partial<PostulacionEstado>): PostulacionGateway {
  return {
    ...mockGateway,
    async cargar() {
      const datos = await mockGateway.cargar();
      return { ...datos, postulacion: { ...datos.postulacion, ...cambios } };
    },
  };
}

function abrir(gateway: PostulacionGateway) {
  render(
    <MemoryRouter>
      <MiPostulacionPage gateway={gateway} />
    </MemoryRouter>,
  );
}

beforeEach(() => sessionStorage.clear());
afterEach(cleanup);

describe('MiPostulacionPage (SDD 8.5.10)', () => {
  it('muestra un borrador como no enviado, con acceso para continuar', async () => {
    abrir(conEstado({ estado: 'BORRADOR', elegibilidad: null }));
    expect(await screen.findByText('Borrador — no enviada', undefined, ESPERA)).toBeTruthy();
    expect(document.title).toBe('Mi postulación | Postulación BIESS');
    expect(screen.getByRole('heading', { level: 1, name: 'Mi postulación' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Continuar con su postulación' })).toBeTruthy();
    expect(screen.getByText('Se asigna al enviar la postulación.')).toBeTruthy();
  });

  it('muestra el resultado de no elegibilidad con motivo y fecha, sin acción de continuar', async () => {
    abrir(
      conEstado({
        estado: 'NO_ELEGIBLE',
        elegibilidad: { elegible: false, cargoId: 2, mensaje: 'Mensaje institucional de cierre.', fecha: '2026-11-03T18:12:42Z' },
      }),
    );
    expect(await screen.findByText('No elegible', undefined, ESPERA)).toBeTruthy();
    expect(screen.getByText('No cumple los requisitos de elegibilidad para este cargo')).toBeTruthy();
    expect(screen.getByText('Mensaje institucional de cierre.')).toBeTruthy();
    expect(screen.getByText(/Fecha y hora del resultado: 3 de noviembre de 2026.*13:12/)).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Continuar con su postulación' })).toBeNull();
  });
});

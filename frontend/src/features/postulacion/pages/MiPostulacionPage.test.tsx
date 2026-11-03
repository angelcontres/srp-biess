import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi } from 'vitest';
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
    // No code and no PDF: "No aplica" (code and form), never "se asigna al enviar".
    expect(screen.getAllByText('No aplica.')).toHaveLength(2);
    expect(screen.queryByText('Se asigna al enviar la postulación.')).toBeNull();
    expect(screen.queryByRole('button', { name: 'Descargar PDF' })).toBeNull();
  });

  it('enviada: muestra código, fecha, historial y descarga el PDF', async () => {
    const descargarFormulario = vi.fn(() => Promise.resolve(new Blob(['%PDF-1.7'])));
    URL.createObjectURL = vi.fn(() => 'blob:pdf');
    URL.revokeObjectURL = vi.fn();
    const gateway: PostulacionGateway = {
      ...conEstado({ estado: 'ENVIADA', codigoUnico: 'BIESS-2026-AFI-000123', fechaEnvio: '2026-11-03T21:05:00Z' }),
      descargarFormulario,
      historial: () =>
        Promise.resolve([
          { estado: 'ENVIADA', fecha: '2026-11-03T21:05:00Z' },
          { estado: 'LISTA_PARA_ENVIO', fecha: '2026-11-03T21:04:00Z' },
          { estado: 'BORRADOR', fecha: '2026-11-01T15:00:00Z' },
        ]),
    };
    abrir(gateway);
    expect(await screen.findByText('BIESS-2026-AFI-000123', undefined, ESPERA)).toBeTruthy();
    // Submission date in the current state (and again in the history), in Ecuador time.
    expect(screen.getAllByText(/3 de noviembre de 2026.*16:05/).length).toBeGreaterThanOrEqual(1);

    const items = await screen.findAllByRole('listitem');
    expect(items.map((li) => li.querySelector('span')?.textContent)).toEqual([
      'Enviada',
      'Lista para envío',
      'Borrador — no enviada',
    ]);

    await userEvent.click(screen.getByRole('button', { name: 'Descargar PDF' }));
    await waitFor(() => expect(descargarFormulario).toHaveBeenCalledTimes(1));
  });

  it('en borrador no ofrece descargar el PDF', async () => {
    abrir(conEstado({ estado: 'BORRADOR', elegibilidad: null }));
    expect(await screen.findByText('El formulario en PDF se genera al enviar la postulación.', undefined, ESPERA)).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Descargar PDF' })).toBeNull();
  });
});

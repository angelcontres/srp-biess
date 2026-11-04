import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { cleanup, render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { validatePdf } from '@/components/ui/FileUpload';
import { mockGateway } from '../services/mockGateway';
import type { PostulacionGateway } from '../services';
import PostulacionPage from './PostulacionPage';

const ESPERA = { timeout: 4000 };
const estadoCivil = () => screen.getByRole('combobox', { name: /Estado civil/ });

async function abrirPersonal(gateway: PostulacionGateway = mockGateway) {
  const user = userEvent.setup();
  // The views link to "Mi postulación", so they need a router.
  render(
    <MemoryRouter>
      <PostulacionPage gateway={gateway} />
    </MemoryRouter>,
  );
  await screen.findByRole('heading', { level: 1, name: 'Información personal' }, ESPERA);
  await screen.findByRole('combobox', { name: /Estado civil/ }, ESPERA);
  return user;
}

beforeEach(() => {
  sessionStorage.clear();
  window.scrollTo = () => undefined; // not implemented by jsdom
});
afterEach(cleanup);

describe('PostulacionPage (SRP-013)', () => {
  it('muestra los datos verificados en solo lectura', async () => {
    await abrirPersonal();
    expect(screen.getByLabelText(/Cédula de identidad/)).toHaveProperty('readOnly', true);
    expect(document.title).toBe('Sección 1 de 10: Información personal | Postulación BIESS');
  });

  it('"Guardar y continuar" exige los datos de pareja y muestra el resumen de errores', async () => {
    const user = await abrirPersonal();
    await user.selectOptions(estadoCivil(), 'CASADO');
    await user.click(screen.getByRole('button', { name: 'Guardar y continuar' }));

    const resumen = await screen.findByText('Revise los siguientes datos:');
    const lista = within(resumen.parentElement as HTMLElement);
    expect(lista.getByRole('link', { name: 'Ingrese los nombres y apellidos de su pareja.' })).toBeTruthy();
    expect(lista.getByRole('link', { name: 'Ingrese la cédula de su pareja.' })).toBeTruthy();
  });

  it('"Guardar progreso" guarda una sección incompleta sin perder lo ingresado', async () => {
    const user = await abrirPersonal();
    await user.selectOptions(estadoCivil(), 'CASADO');
    await user.type(screen.getByLabelText(/Nombres y apellidos de la pareja/), 'Juan Gómez');
    await user.click(screen.getByRole('button', { name: 'Guardar progreso' }));

    expect(await screen.findByText(/Progreso guardado a las/, undefined, ESPERA)).toBeTruthy();
    expect(screen.getByLabelText(/Nombres y apellidos de la pareja/)).toHaveProperty('value', 'Juan Gómez');
  });

  it('un error del servidor no borra los datos del formulario', async () => {
    const fallido: PostulacionGateway = {
      ...mockGateway,
      guardarSeccion: () => Promise.reject(Object.assign(new Error('500'), { isAxiosError: false })),
    };
    const user = await abrirPersonal(fallido);
    await user.selectOptions(estadoCivil(), 'SOLTERO');
    await user.click(screen.getByRole('button', { name: 'Guardar progreso' }));

    expect(await screen.findByText('No se pudo guardar.', undefined, ESPERA)).toBeTruthy();
    expect(screen.getByText(/No se guardaron los cambios/)).toBeTruthy();
    expect(estadoCivil()).toHaveProperty('value', 'SOLTERO');
  });

  it('pide confirmación al cambiar de sección con cambios sin guardar', async () => {
    const user = await abrirPersonal();
    await user.selectOptions(estadoCivil(), 'SOLTERO');
    const nav = screen.getAllByRole('navigation', { name: 'Secciones de la postulación' })[0] as HTMLElement;
    await user.click(within(nav).getAllByRole('button', { name: /Domicilio/ })[0] as HTMLElement);

    expect(await screen.findByRole('dialog', { name: 'Tiene cambios sin guardar en esta sección.' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Seguir editando' })).toBe(document.activeElement);
  });

  it('una postulación NO_ELEGIBLE se muestra en solo lectura y sin acciones de guardado', async () => {
    const cerrada: PostulacionGateway = {
      ...mockGateway,
      async cargar() {
        const datos = await mockGateway.cargar();
        return { ...datos, postulacion: { ...datos.postulacion, estado: 'NO_ELEGIBLE' } };
      },
    };
    await abrirPersonal(cerrada);
    expect(screen.getByText('Su postulación ya no admite cambios.')).toBeTruthy();
    expect(estadoCivil().matches(':disabled')).toBe(true); // disabled through its fieldset
    expect(screen.queryByRole('button', { name: 'Guardar y continuar' })).toBeNull();
    expect(screen.queryByRole('button', { name: /Guardar progreso/ })).toBeNull();
  });

  it('en escritorio el estado de guardado aparece una sola vez, en la columna de secciones', async () => {
    const original = window.matchMedia;
    window.matchMedia = ((query: string) => ({
      matches: query === '(min-width: 1024px)',
      media: query,
      addEventListener: () => undefined,
      removeEventListener: () => undefined,
    })) as unknown as typeof window.matchMedia;
    try {
      await abrirPersonal();
      const estados = await screen.findAllByRole('status');
      const guardado = estados.filter((el) => /guard/i.test(el.textContent ?? ''));
      expect(guardado).toHaveLength(1);
      expect(guardado[0]?.closest('aside')).not.toBeNull();
    } finally {
      window.matchMedia = original;
    }
  });

  it('las secciones posteriores a Cargo aparecen como no disponibles sin elegibilidad', async () => {
    await abrirPersonal();
    const nav = screen.getAllByRole('navigation', { name: 'Secciones de la postulación' })[0] as HTMLElement;
    expect(within(nav).getAllByText('No disponible').length).toBeGreaterThan(0);
  });
});

describe('Cargo y Experiencia', () => {
  it('confirma el cargo en un modal y muestra la elegibilidad en pantalla propia', async () => {
    const user = await abrirPersonal();
    const nav = screen.getAllByRole('navigation', { name: 'Secciones de la postulación' })[0] as HTMLElement;
    await user.click(within(nav).getAllByRole('button', { name: /Cargo/ })[0] as HTMLElement);
    await user.click(await screen.findByRole('radio', { name: /Sector de Afiliados/ }, ESPERA));
    await user.click(screen.getByRole('button', { name: 'Confirmar cargo' }));

    const dialogo = await screen.findByRole('dialog', { name: '¿Confirmar el cargo seleccionado?' });
    expect(within(dialogo).getByRole('button', { name: 'Revisar otra vez' })).toBe(document.activeElement);
    await user.click(within(dialogo).getByRole('button', { name: 'Confirmar cargo' }));

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Cumple los requisitos de elegibilidad' }, ESPERA),
    ).toBeTruthy();
  });

  it('calcula el tiempo de experiencia en años, meses y días al ingresar fechas', async () => {
    sessionStorage.setItem(
      'mock:postulacion',
      JSON.stringify({
        postulacion: {
          id: 'p1',
          estado: 'BORRADOR',
          version: 1,
          secciones: {},
          ultimoGuardado: null,
          elegibilidad: { elegible: true, cargoId: 1, mensaje: 'Postulante afiliado al IESS.', fecha: '2026-11-01T10:00:00Z' },
          seccionReanudacion: 'experiencia',
        },
        valores: {},
      }),
    );
    const user = userEvent.setup();
    render(<PostulacionPage gateway={mockGateway} />);
    await user.click(await screen.findByRole('button', { name: 'Agregar experiencia' }, ESPERA));

    const dialogo = await screen.findByRole('dialog', { name: 'Agregar experiencia' });
    await user.type(within(dialogo).getByRole('textbox', { name: 'Fecha de inicio' }), '15012020');
    await user.type(within(dialogo).getByRole('textbox', { name: 'Fecha de fin' }), '10032023');
    expect(within(dialogo).getByText('3 años, 1 mes y 23 días')).toBeTruthy();
    expect(within(dialogo).getByText('Calculado automáticamente')).toBeTruthy();
  });
});

describe('Domicilio: cascada país, provincia y cantón', () => {
  function sembrarDomicilio() {
    const domicilio = {
      paisId: '1', provinciaId: '17', cantonId: '1701', provinciaManual: '', ciudadManual: '',
      telefonoCelular: '0991234567', correoAlterno: '', callePrincipal: 'Av. Amazonas', numero: 'N34-120',
      calleSecundaria: '', sector: '', referencia: '', telefonoDomicilio: '',
    };
    sessionStorage.setItem(
      'mock:postulacion',
      JSON.stringify({
        postulacion: {
          id: 'p1', estado: 'BORRADOR', version: 1, secciones: {}, ultimoGuardado: null,
          elegibilidad: null, seccionReanudacion: 'domicilio',
        },
        valores: { domicilio },
      }),
    );
  }

  it('recupera el cantón guardado según la provincia', async () => {
    sembrarDomicilio();
    render(<PostulacionPage gateway={mockGateway} />);
    await screen.findByRole('option', { name: 'Quito' }, ESPERA);
    await waitFor(() => expect(screen.getByRole('combobox', { name: /Cantón/ })).toHaveProperty('value', '1701'));
    expect(screen.getByRole('combobox', { name: /Provincia/ })).toHaveProperty('value', '17');
  });

  it('al cambiar de provincia limpia el cantón y, sin cantones, permite guardar', async () => {
    sembrarDomicilio();
    const user = userEvent.setup();
    render(<PostulacionPage gateway={mockGateway} />);
    await screen.findByRole('option', { name: 'Quito' }, ESPERA);

    await user.selectOptions(screen.getByRole('combobox', { name: /Provincia/ }), '2');
    expect(await screen.findByText(/No hay cantones registrados para esta provincia/, undefined, ESPERA)).toBeTruthy();
    expect(screen.getByRole('combobox', { name: /Cantón/ })).toHaveProperty('value', '');

    await user.click(screen.getByRole('button', { name: 'Guardar y continuar' }));
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Familiares en el IESS o el BIESS' }, ESPERA),
    ).toBeTruthy();
  });
});

describe('validatePdf (SDD 8.6.3)', () => {
  const archivo = (name: string, size: number, type: string) => ({ name, size, type }) as File;

  it('rechaza archivos que superan el límite con un mensaje específico', () => {
    expect(validatePdf(archivo('titulo.pdf', 3.4 * 1024 * 1024, 'application/pdf'), 2)).toBe(
      'El archivo pesa 3,4 MB. El tamaño máximo permitido es 2 MB.',
    );
    expect(validatePdf(archivo('respaldo.pdf', 3.4 * 1024 * 1024, 'application/pdf'), 4)).toBeNull();
  });

  it('rechaza archivos que no son PDF', () => {
    expect(validatePdf(archivo('titulo.docx', 1000, 'application/msword'), 2)).toBe(
      'El archivo no es PDF. Seleccione un archivo en formato PDF.',
    );
  });
});

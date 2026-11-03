import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import App from './App';
import { SessionProvider } from '@/features/auth/session/SessionProvider';
import { expireSession, getAccessToken, setAccessToken } from '@/store/authStore';

const mocks = vi.hoisted(() => ({
  refresh: vi.fn<() => Promise<string | null>>(),
  get: vi.fn(),
  post: vi.fn(),
}));

vi.mock('@/services/api/client', async (importOriginal) => {
  const original = await importOriginal<typeof import('@/services/api/client')>();
  return {
    ...original,
    refreshAccessToken: mocks.refresh,
    api: { ...original.api, get: mocks.get, post: mocks.post },
  };
});

// The form has its own tests; here only its heading matters.
vi.mock('@/features/postulacion/pages/PostulacionPage', () => ({
  default: () => <h1>Formulario de postulación</h1>,
}));

function renderEn(ruta: string) {
  return render(
    <MemoryRouter initialEntries={[ruta]}>
      <SessionProvider>
        <App />
      </SessionProvider>
    </MemoryRouter>,
  );
}

function conSesion() {
  mocks.refresh.mockImplementation(async () => {
    setAccessToken('token-valido');
    return 'token-valido';
  });
  mocks.get.mockResolvedValue({ data: { data: { email: 'postulante@biess.gob.ec' } } });
}

beforeEach(() => {
  setAccessToken(null);
  sessionStorage.clear();
  mocks.refresh.mockResolvedValue(null);
  mocks.post.mockResolvedValue({ data: {} });
});
afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

describe('Rutas protegidas y sesión (SRP-015)', () => {
  it('sin sesión, /formulario redirige al inicio de sesión', async () => {
    renderEn('/formulario');
    expect(await screen.findByRole('heading', { level: 1, name: 'Iniciar sesión' })).toBeTruthy();
    expect(screen.queryByText('Formulario de postulación')).toBeNull();
  });

  it('con sesión restaurada por la cookie, /formulario muestra el formulario', async () => {
    conSesion();
    renderEn('/formulario');
    expect(await screen.findByRole('heading', { name: 'Formulario de postulación' })).toBeTruthy();
    expect(screen.getByText('postulante@biess.gob.ec')).toBeTruthy();
  });

  it('con sesión, el inicio de sesión lleva directo al formulario', async () => {
    conSesion();
    renderEn('/ingresar');
    expect(await screen.findByRole('heading', { name: 'Formulario de postulación' })).toBeTruthy();
  });

  it('al iniciar sesión vuelve a la página protegida solicitada', async () => {
    const user = userEvent.setup();
    mocks.post.mockResolvedValueOnce({
      data: { data: { access_token: 'nuevo', email: 'postulante@biess.gob.ec' } },
    });
    renderEn('/formulario');
    await user.type(await screen.findByLabelText(/Cédula de identidad/), '0926687856');
    await user.type(screen.getByLabelText(/Contraseña/, { selector: 'input' }), 'Clave2026*');
    await user.click(screen.getByRole('button', { name: /Iniciar sesión|Ingresar/ }));
    expect(await screen.findByRole('heading', { name: 'Formulario de postulación' })).toBeTruthy();
    expect(getAccessToken()).toBe('nuevo');
  });

  it('si la sesión expira, lleva al inicio de sesión con un aviso y borra los borradores', async () => {
    conSesion();
    sessionStorage.setItem('draft:test:personal', '{}');
    renderEn('/formulario');
    await screen.findByRole('heading', { name: 'Formulario de postulación' });

    act(() => expireSession());

    expect(await screen.findByRole('heading', { level: 1, name: 'Iniciar sesión' })).toBeTruthy();
    expect(screen.getByText(/Su sesión expiró/)).toBeTruthy();
    expect(screen.queryByText('postulante@biess.gob.ec')).toBeNull();
    expect(sessionStorage.getItem('draft:test:personal')).toBeNull();
  });

  it('cerrar sesión llama a /auth/logout, limpia el token y vuelve a la portada', async () => {
    const user = userEvent.setup();
    conSesion();
    renderEn('/formulario');
    await screen.findByRole('heading', { name: 'Formulario de postulación' });

    await user.click(screen.getByRole('button', { name: /Salir/ }));

    expect(mocks.post).toHaveBeenCalledWith('/auth/logout');
    expect(getAccessToken()).toBeNull();
    expect(screen.queryByText('Formulario de postulación')).toBeNull();
    expect(screen.queryByText('postulante@biess.gob.ec')).toBeNull();
  });
});

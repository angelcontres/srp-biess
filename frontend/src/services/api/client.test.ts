import { afterEach, describe, expect, it, vi } from 'vitest';
import axios from 'axios';
import { refreshAccessToken } from './client';
import { getAccessToken, onSessionExpired, setAccessToken } from '@/store/authStore';

afterEach(() => {
  vi.restoreAllMocks();
  setAccessToken(null);
});

describe('refreshAccessToken (SRP-015)', () => {
  it('guarda el token nuevo y comparte una sola llamada entre peticiones simultáneas', async () => {
    const post = vi.spyOn(axios, 'post').mockResolvedValue({ data: { data: { access_token: 'nuevo' } } });
    const [a, b] = await Promise.all([refreshAccessToken(), refreshAccessToken()]);
    expect([a, b]).toEqual(['nuevo', 'nuevo']);
    expect(post).toHaveBeenCalledTimes(1);
    expect(getAccessToken()).toBe('nuevo');
  });

  it('si falla con una sesión abierta, avisa que la sesión expiró', async () => {
    vi.spyOn(axios, 'post').mockRejectedValue(new Error('401'));
    const expirada = vi.fn();
    const quitar = onSessionExpired(expirada);
    setAccessToken('vencido');

    expect(await refreshAccessToken()).toBeNull();
    expect(expirada).toHaveBeenCalledTimes(1);
    expect(getAccessToken()).toBeNull();
    quitar();
  });

  it('si falla al restaurar sin sesión previa, no avisa', async () => {
    vi.spyOn(axios, 'post').mockRejectedValue(new Error('401'));
    const expirada = vi.fn();
    const quitar = onSessionExpired(expirada);

    expect(await refreshAccessToken()).toBeNull();
    expect(expirada).not.toHaveBeenCalled();
    quitar();
  });
});

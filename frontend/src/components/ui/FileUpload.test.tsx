import { AxiosError, AxiosHeaders, type AxiosResponse } from 'axios';
import { useState } from 'react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import FileUpload from './FileUpload';
import type { UploadedFileInfo } from './FileUpload';

afterEach(cleanup);

function errorHttp(status: number, data: unknown): AxiosError {
  const config = { headers: new AxiosHeaders() };
  const response = { status, data, statusText: '', headers: {}, config } as AxiosResponse;
  return new AxiosError('fallo', String(status), config, undefined, response);
}

const pdf = () => new File(['%PDF-1.7'], 'titulo.pdf', { type: 'application/pdf' });
const previo: UploadedFileInfo = {
  id: 'doc-1',
  name: 'anterior.pdf',
  sizeBytes: 100,
  uploadedAt: '2026-11-03T18:00:00Z',
};

function Campo(props: {
  onUpload: () => Promise<UploadedFileInfo>;
  onRemove?: () => Promise<void>;
  inicial?: UploadedFileInfo | null;
}) {
  const [value, setValue] = useState<UploadedFileInfo | null>(props.inicial ?? null);
  return (
    <>
      <FileUpload
        label="Título"
        maxSizeMB={2}
        value={value}
        onChange={setValue}
        onUpload={props.onUpload}
        onRemove={props.onRemove}
      />
      <p data-testid="valor">{value?.name ?? 'vacío'}</p>
    </>
  );
}

function elegir(file: File) {
  const input = document.querySelector('input[type="file"]') as HTMLInputElement;
  fireEvent.change(input, { target: { files: [file] } });
}

describe('FileUpload (SRP-017b)', () => {
  it('un 413 del servidor muestra su mensaje y no ofrece reintentar', async () => {
    const onUpload = () =>
      Promise.reject(
        errorHttp(413, { code: 'DOC_TOTAL_EXCEDIDO', error: 'Sus documentos superan el tamaño total permitido.' }),
      );
    render(<Campo onUpload={onUpload} />);
    elegir(pdf());
    expect(await screen.findByText('Sus documentos superan el tamaño total permitido.')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Reintentar' })).toBeNull();
    expect(screen.queryByText(/problema de conexión/)).toBeNull();
  });

  it('el 413 del límite de cuerpo, sin JSON, muestra un mensaje de tamaño', async () => {
    render(<Campo onUpload={() => Promise.reject(errorHttp(413, 'Request Entity Too Large'))} />);
    elegir(pdf());
    expect(await screen.findByText('El archivo supera el tamaño máximo permitido para este documento.')).toBeTruthy();
  });

  it('sin respuesta del servidor ofrece reintentar por conexión', async () => {
    render(<Campo onUpload={() => Promise.reject(new AxiosError('Network Error', 'ERR_NETWORK'))} />);
    elegir(pdf());
    expect(await screen.findByText(/problema de conexión/)).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Reintentar' })).toBeTruthy();
  });

  it('al reemplazar elimina el documento anterior en el servidor', async () => {
    const onRemove = vi.fn(() => Promise.resolve());
    const nuevo: UploadedFileInfo = {
      id: 'doc-2',
      name: 'titulo.pdf',
      sizeBytes: 8,
      uploadedAt: '2026-11-03T18:05:00Z',
    };
    render(<Campo inicial={previo} onUpload={() => Promise.resolve(nuevo)} onRemove={onRemove} />);
    elegir(pdf());
    await waitFor(() => expect(screen.getByTestId('valor').textContent).toBe('titulo.pdf'));
    expect(onRemove).toHaveBeenCalledTimes(1);
  });

  it('eliminar un documento que ya no existe (404) limpia el campo', async () => {
    render(
      <Campo
        inicial={previo}
        onUpload={() => Promise.reject(new Error())}
        onRemove={() => Promise.reject(errorHttp(404, { code: 'DOC_NO_ENCONTRADO' }))}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Eliminar anterior.pdf' }));
    await waitFor(() => expect(screen.getByTestId('valor').textContent).toBe('vacío'));
  });

  it('si eliminar falla por otro motivo, conserva el documento y avisa', async () => {
    render(
      <Campo
        inicial={previo}
        onUpload={() => Promise.reject(new Error())}
        onRemove={() => Promise.reject(errorHttp(409, { code: 'POS_CERRADA' }))}
      />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Eliminar anterior.pdf' }));
    expect(await screen.findByText('No se pudo eliminar el documento. Reintente.')).toBeTruthy();
    expect(screen.getByTestId('valor').textContent).toBe('anterior.pdf');
  });
});

import { useId, useRef, useState } from 'react';
import { CircleAlert, FileText, RotateCw, Upload } from 'lucide-react';
import Button from './Button';
import FieldError, { RequiredMark } from './FieldMessages';
import { normalizeError } from '@/services/api/client';
import { formatDateTime, formatFileSize, LOCALE } from '@/utils/format';

export interface UploadedFileInfo {
  id: string;
  name: string;
  sizeBytes: number;
  uploadedAt: string;
  url?: string;
}

export type UploadHandler = (
  file: File,
  onProgress: (percent: number) => void,
  signal: AbortSignal,
) => Promise<UploadedFileInfo>;

export interface FileUploadProps {
  label: string;
  /** What the document must contain; shown before the control. */
  contentHint?: string;
  maxSizeMB: number;
  value: UploadedFileInfo | null;
  onChange: (value: UploadedFileInfo | null) => void;
  onUpload: UploadHandler;
  /** Fetches the stored PDF; the Blob is saved with the attachment name. */
  onDownload?: () => Promise<Blob>;
  /** Deletes the stored PDF. The field is cleared only after the server confirms. */
  onRemove?: () => Promise<void>;
  /** Validation error coming from the form (for example, a required document). */
  error?: string;
  required?: boolean;
  id?: string;
}

type UploadState =
  | { kind: 'idle' }
  | { kind: 'uploading'; fileName: string; percent: number }
  | { kind: 'invalid'; message: string }
  | { kind: 'failed'; file: File; message: string };

const sizeFormatter = new Intl.NumberFormat(LOCALE, { maximumFractionDigits: 1 });

/** Checks type and size before uploading (SDD 8.6.3). Returns a message or null. */
export function validatePdf(file: File, maxSizeMB: number): string | null {
  const isPdf = file.type === 'application/pdf' || file.name.toLowerCase().endsWith('.pdf');
  if (!isPdf) return 'El archivo no es PDF. Seleccione un archivo en formato PDF.';
  if (file.size > maxSizeMB * 1024 * 1024) {
    const sizeMB = sizeFormatter.format(file.size / (1024 * 1024));
    return `El archivo pesa ${sizeMB} MB. El tamaño máximo permitido es ${maxSizeMB} MB.`;
  }
  return null;
}

const CONNECTION_FAILURE = 'No se pudo subir el archivo por un problema de conexión. Reintente.';

/**
 * Upload failure: a 4xx rejection (415 not a PDF, 413 too large, 409 closed application) shows
 * the server message and asks for another file; only network or 5xx failures offer a retry.
 */
function describeUploadFailure(error: unknown): { retry: boolean; message: string } {
  const problem = normalizeError(error);
  if (!problem.status) return { retry: true, message: CONNECTION_FAILURE };
  if (problem.status >= 500) {
    return { retry: true, message: 'No se pudo subir el archivo por un problema del sistema. Reintente.' };
  }
  // 413 from the body limit arrives without the API envelope.
  if (problem.status === 413 && problem.code === '413') {
    return { retry: false, message: 'El archivo supera el tamaño máximo permitido para este documento.' };
  }
  return { retry: false, message: problem.message };
}

/** PDF attachment with the five states of DESIGN file-upload.md. */
export default function FileUpload({
  label,
  contentHint,
  maxSizeMB,
  value,
  onChange,
  onUpload,
  onDownload,
  onRemove,
  error,
  required,
  id,
}: FileUploadProps) {
  const generatedId = useId();
  const fieldId = id ?? `file-${generatedId}`;
  const hintId = `${fieldId}-hint`;
  const errorId = `${fieldId}-error`;
  const actionErrorId = `${fieldId}-action-error`;
  const inputRef = useRef<HTMLInputElement>(null);
  const abortRef = useRef<AbortController | null>(null);
  const [state, setState] = useState<UploadState>({ kind: 'idle' });
  const [busy, setBusy] = useState<'download' | 'remove' | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const upload = async (file: File) => {
    setActionError(null);
    const invalid = validatePdf(file, maxSizeMB);
    if (invalid) {
      setState({ kind: 'invalid', message: invalid });
      return;
    }
    const controller = new AbortController();
    abortRef.current = controller;
    setState({ kind: 'uploading', fileName: file.name, percent: 0 });
    try {
      const uploaded = await onUpload(
        file,
        (percent) => setState({ kind: 'uploading', fileName: file.name, percent }),
        controller.signal,
      );
      // Replacing: the previous PDF is deleted on the server so it stops counting toward the
      // 50 MB quota. Best effort: the new file is already stored either way.
      if (value && onRemove) void onRemove().catch(() => undefined);
      onChange(uploaded);
      setState({ kind: 'idle' });
    } catch (error) {
      if (controller.signal.aborted) {
        setState({ kind: 'idle' });
      } else {
        const failure = describeUploadFailure(error);
        setState(failure.retry ? { kind: 'failed', file, message: failure.message } : { kind: 'invalid', message: failure.message });
      }
    } finally {
      abortRef.current = null;
    }
  };

  const download = async () => {
    if (!onDownload) return;
    setActionError(null);
    setBusy('download');
    try {
      const blob = await onDownload();
      const objectUrl = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = objectUrl;
      link.download = value?.name ?? 'documento.pdf';
      document.body.append(link);
      link.click();
      link.remove();
      window.setTimeout(() => URL.revokeObjectURL(objectUrl), 0);
    } catch {
      setActionError('No se pudo descargar el documento. Reintente.');
    } finally {
      setBusy(null);
    }
  };

  const remove = async () => {
    setActionError(null);
    if (!onRemove) {
      onChange(null);
      setState({ kind: 'idle' });
      return;
    }
    setBusy('remove');
    try {
      await onRemove();
      onChange(null);
      setState({ kind: 'idle' });
    } catch (error) {
      if (normalizeError(error).status === 404) {
        // Already gone on the server (DOC_NO_ENCONTRADO): clear the stale reference.
        onChange(null);
        setState({ kind: 'idle' });
        return;
      }
      setActionError('No se pudo eliminar el documento. Reintente.');
    } finally {
      setBusy(null);
    }
  };

  const openPicker = () => inputRef.current?.click();
  const pickerLabel = value ? 'Reemplazar archivo' : 'Seleccionar archivo PDF';
  const hintText = `Formato aceptado: .pdf. Tamaño máximo: ${maxSizeMB} MB.${contentHint ? ` ${contentHint}` : ''}`;

  return (
    <div id={fieldId} tabIndex={-1} className="outline-none">
      <p className="text-base font-medium text-neutral-900">
        {label}
        {required ? <RequiredMark /> : null}
      </p>
      <p id={hintId} className="mt-1 text-sm text-muted">
        {hintText}
      </p>

      <input
        ref={inputRef}
        type="file"
        accept="application/pdf,.pdf"
        tabIndex={-1}
        aria-hidden="true"
        className="sr-only"
        onChange={(event) => {
          const file = event.target.files?.[0];
          event.target.value = '';
          if (file) void upload(file);
        }}
      />

      <div className="mt-2">
        {state.kind === 'uploading' ? (
          <div className="rounded-md border border-line-strong p-3">
            <p className="break-all text-base text-neutral-900">{state.fileName}</p>
            <div
              role="progressbar"
              aria-label={`Subiendo ${state.fileName}`}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-valuenow={state.percent}
              className="mt-2 h-1 w-full bg-line"
            >
              <div className="h-full bg-primary" style={{ width: `${state.percent}%` }} />
            </div>
            <div className="mt-2 flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm tabular-nums text-muted" aria-live="polite">
                {state.percent} %
              </p>
              <Button variant="tertiary" onClick={() => abortRef.current?.abort()}>
                Cancelar
              </Button>
            </div>
          </div>
        ) : value ? (
          <div className="flex flex-col gap-3 rounded-md border border-line-strong p-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex min-w-0 items-start gap-3">
              <FileText size={20} strokeWidth={2} aria-hidden="true" className="mt-0.5 shrink-0 text-primary" />
              <div className="min-w-0">
                <p className="break-all text-base text-neutral-900">{value.name}</p>
                <p className="text-sm tabular-nums text-muted">
                  {formatFileSize(value.sizeBytes)}, cargado el {formatDateTime(value.uploadedAt)}
                </p>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2">
              {onDownload ? (
                <Button
                  variant="tertiary"
                  onClick={() => void download()}
                  isLoading={busy === 'download'}
                  loadingText="Descargando"
                  aria-label={`Descargar ${value.name}`}
                >
                  Descargar
                </Button>
              ) : value.url ? (
                <a href={value.url} target="_blank" rel="noreferrer" className="inline-flex min-h-11 items-center">
                  Ver<span className="sr-only"> {value.name}</span>
                </a>
              ) : null}
              <Button variant="tertiary" onClick={openPicker} aria-label={`Reemplazar ${value.name}`}>
                Reemplazar
              </Button>
              <Button
                variant="tertiary"
                onClick={() => void remove()}
                isLoading={busy === 'remove'}
                loadingText="Eliminando"
                aria-label={`Eliminar ${value.name}`}
              >
                Eliminar
              </Button>
            </div>
          </div>
        ) : (
          <Button
            variant="secondary"
            onClick={openPicker}
            leftIcon={<Upload size={20} strokeWidth={2} aria-hidden="true" />}
            aria-describedby={hintId}
          >
            {state.kind === 'invalid' ? 'Seleccionar otro archivo' : pickerLabel}
          </Button>
        )}
      </div>

      {state.kind === 'invalid' ? <FieldError error={state.message} errorId={`${fieldId}-invalid`} /> : null}

      {actionError ? <FieldError error={actionError} errorId={actionErrorId} /> : null}

      {state.kind === 'failed' ? (
        <div role="alert" className="mt-2 flex flex-col gap-2 rounded-md border border-danger border-l-4 bg-danger-50 p-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="flex gap-2 text-base text-neutral-900">
            <CircleAlert size={20} strokeWidth={2} aria-hidden="true" className="shrink-0 text-danger" />
            {state.message}
          </p>
          <Button
            variant="secondary"
            onClick={() => void upload(state.file)}
            leftIcon={<RotateCw size={20} strokeWidth={2} aria-hidden="true" />}
          >
            Reintentar
          </Button>
        </div>
      ) : null}

      {state.kind !== 'invalid' ? <FieldError error={error} errorId={errorId} /> : null}
    </div>
  );
}

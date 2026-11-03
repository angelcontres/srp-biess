import { useState } from 'react';
import type { FormEvent } from 'react';
import { ArrowLeft } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import TextField from '@/components/ui/TextField';
import { RequiredNote } from '@/components/ui/FieldMessages';
import { api, normalizeError } from '@/services/api/client';

export interface ForgotPasswordPageProps {
  onNavigateToLogin?: () => void;
  onNavigateToReset?: (token?: string) => void;
}

export default function ForgotPasswordPage({ onNavigateToLogin, onNavigateToReset }: ForgotPasswordPageProps) {
  const [identificador, setIdentificador] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successInfo, setSuccessInfo] = useState<{ mensaje: string; tokenDev?: string } | null>(null);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const id = identificador.trim();
    if (!id) return;

    setIsLoading(true);
    setErrorMessage(null);
    setSuccessInfo(null);

    try {
      const response = await api.post<{ data: { mensaje: string; tokenDev?: string } }>(
        '/auth/recuperar-password',
        { identificador: id }
      );

      const data = response.data?.data;
      setSuccessInfo({
        mensaje: data?.mensaje || 'Si los datos corresponden a una cuenta registrada, se enviarán las instrucciones.',
        tokenDev: data?.tokenDev,
      });
    } catch (error) {
      const normalized = normalizeError(error);
      setErrorMessage(normalized.message || 'No se pudo procesar la solicitud. Intente nuevamente.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-12">
      <div className="mx-auto max-w-md rounded-lg border border-line-strong bg-surface p-6 sm:p-8">
        <h1 className="text-xl font-bold leading-tight sm:text-2xl text-neutral-900">
          Recuperación de contraseña
        </h1>
        <p className="mt-2 text-base text-neutral-700">
          Ingrese su número de cédula o correo electrónico para recibir las instrucciones de restablecimiento.
        </p>

        {!successInfo ? (
          <form onSubmit={handleSubmit} className="mt-6 space-y-6" noValidate>
            <RequiredNote />
            <TextField
              label="Cédula de identidad o correo"
              value={identificador}
              onChange={(e) => setIdentificador(e.target.value)}
              required
              disabled={isLoading}
              autoComplete="username"
              helperText="Ingrese los 10 dígitos de su cédula o la dirección de correo con la que se registró."
            />

            <Button
              type="submit"
              fullWidth
              isLoading={isLoading}
              disabled={identificador.trim().length === 0}
              loadingText="Enviando solicitud…"
            >
              Enviar enlace de recuperación
            </Button>
          </form>
        ) : (
          <div className="mt-6 space-y-6">
            <Alert variant="info" title="Solicitud enviada">
              {successInfo.mensaje}
            </Alert>

            {successInfo.tokenDev ? (
              <div className="rounded-md border border-line bg-surface-muted p-4 space-y-3">
                <p className="text-xs font-semibold text-neutral-700">MODO DEMOSTRACIÓN (TOKEN DEV):</p>
                <p className="break-all font-mono text-xs text-neutral-900 select-all">{successInfo.tokenDev}</p>
                <Button
                  type="button"
                  variant="secondary"
                  fullWidth
                  onClick={() => onNavigateToReset?.(successInfo.tokenDev)}
                >
                  Continuar a Restablecer Contraseña
                </Button>
              </div>
            ) : null}

            <Button
              type="button"
              fullWidth
              variant="secondary"
              onClick={onNavigateToLogin}
            >
              Volver al inicio de sesión
            </Button>
          </div>
        )}

        {errorMessage ? (
          <Alert variant="danger" title="Error en la solicitud" className="mt-6">
            {errorMessage}
          </Alert>
        ) : null}

        <div className="mt-6 border-t border-line pt-4 text-center">
          <button
            type="button"
            onClick={onNavigateToLogin}
            className="inline-flex items-center gap-1.5 text-sm font-medium text-primary underline underline-offset-4 hover:text-primary-dark"
          >
            <ArrowLeft size={16} aria-hidden="true" />
            Volver a iniciar sesión
          </button>
        </div>
      </div>
    </div>
  );
}

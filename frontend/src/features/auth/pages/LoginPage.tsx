import { useState } from 'react';
import type { FormEvent } from 'react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import TextField from '@/components/ui/TextField';
import { api, normalizeError } from '@/services/api/client';

interface CedulaValidationResponse {
  message?: string;
}

export default function LoginPage() {
  const [cedula, setCedula] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsLoading(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      const { data } = await api.post<CedulaValidationResponse>('/auth/validate-cedula', {
        cedula,
      });
      setSuccessMessage(data.message ?? 'Cédula válida autorizada para postular.');
    } catch (error) {
      setErrorMessage(normalizeError(error).message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-12">
      <div className="mx-auto max-w-lg rounded-lg border border-line-strong bg-surface p-4 sm:p-6">
        <h1 className="text-xl font-bold leading-tight sm:text-2xl">Portal de postulaciones</h1>
        <p className="mt-2 max-w-prose text-base text-neutral-700">
          Ingrese su cédula de identidad para verificar su registro conforme a la Resolución C.D.
          701.
        </p>

        <form onSubmit={handleSubmit} className="mt-6 space-y-6" noValidate>
          <TextField
            label="Cédula de identidad"
            value={cedula}
            onChange={(event) => setCedula(event.target.value.replace(/\D/g, ''))}
            maxLength={10}
            inputMode="numeric"
            autoComplete="username"
            required
            className="tabular-nums"
            helperText="Ingrese los 10 números de su cédula."
          />

          <Button
            type="submit"
            fullWidth
            isLoading={isLoading}
            disabled={cedula.length !== 10}
            loadingText="Validando cédula…"
          >
            Ingresar
          </Button>

          {cedula.length !== 10 ? (
            <p className="text-sm text-muted">
              El botón se habilita cuando la cédula tiene 10 números.
            </p>
          ) : null}
        </form>

        {successMessage ? (
          <Alert variant="success" title="Cédula válida" className="mt-6">
            {successMessage}
          </Alert>
        ) : null}

        {errorMessage ? (
          <Alert variant="danger" title="No fue posible validar la cédula" className="mt-6">
            {errorMessage}
          </Alert>
        ) : null}
      </div>
    </div>
  );
}

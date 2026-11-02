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
    <div className="flex min-h-[calc(100vh-64px)] items-center justify-center p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-200 bg-white p-8 shadow-lg">
        <div className="mb-6 text-center">
          <h2 className="text-2xl font-bold text-neutral-900">Portal de Postulaciones</h2>
          <p className="mt-1 text-sm text-neutral-900/60">
            Ingrese su cédula de identidad para verificar su registro conforme a la Resolución C.D.
            701
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4" noValidate>
          <TextField
            label="Cédula de Identidad (10 dígitos)"
            value={cedula}
            onChange={(event) => setCedula(event.target.value.replace(/\D/g, ''))}
            placeholder="Ej. 0926687856"
            maxLength={10}
            inputMode="numeric"
            autoComplete="off"
            required
            className="text-center font-mono text-lg tracking-widest"
            helperText="Solo dígitos. El sistema valida el algoritmo Módulo 10."
          />

          <Button type="submit" size="lg" fullWidth isLoading={isLoading} disabled={cedula.length !== 10} loadingText="Validando Módulo 10...">
            Verificar e Ingresar
          </Button>
        </form>

        {successMessage ? (
          <Alert variant="success" title="Cédula válida" className="mt-4">
            {successMessage}
          </Alert>
        ) : null}

        {errorMessage ? (
          <Alert variant="danger" title="Error de validación" className="mt-4">
            {errorMessage}
          </Alert>
        ) : null}
      </div>
    </div>
  );
}

import { useState } from 'react';
import type { FormEvent } from 'react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import TextField from '@/components/ui/TextField';
import LockoutNotice from '@/features/auth/components/LockoutNotice';
import { useCountdown } from '@/features/auth/hooks/useCountdown';
import { api, DEFAULT_LOCKOUT_SECONDS, normalizeError } from '@/services/api/client';

export type ValidationType = 'FECHA_EXPEDICION' | 'CODIGO_DACTILAR';

export interface RegisterPageProps {
  /** Callback to switch to the login view. */
  onNavigateToLogin?: () => void;
}

interface ValidarIdentidadResponse {
  identidadValidada: boolean;
  tokenValidacion?: string;
  datos?: {
    nombres?: string;
    apellidos?: string;
    genero?: string;
    fechaNacimiento?: string;
    nacionalidad?: string;
  };
}

/** Converts `dd/mm/aaaa` to the `YYYY-MM-DD` expected by the API; null if the date is invalid. */
export function toIsoDate(value: string): string | null {
  const match = /^(\d{2})\/(\d{2})\/(\d{4})$/.exec(value.trim());
  if (!match) return null;
  const [, dd, mm, yyyy] = match;
  const date = new Date(Date.UTC(Number(yyyy), Number(mm) - 1, Number(dd)));
  const isSameDate =
    date.getUTCFullYear() === Number(yyyy) &&
    date.getUTCMonth() === Number(mm) - 1 &&
    date.getUTCDate() === Number(dd);
  return isSameDate ? `${yyyy}-${mm}-${dd}` : null;
}

export default function RegisterPage({ onNavigateToLogin }: RegisterPageProps) {
  const [cedula, setCedula] = useState('');
  const [tipoValidacion, setTipoValidacion] = useState<ValidationType>('FECHA_EXPEDICION');
  const [valorValidacion, setValorValidacion] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isLocked, setIsLocked] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [lockoutMessage, setLockoutMessage] = useState<string | undefined>(undefined);
  const [fechaError, setFechaError] = useState<string | undefined>(undefined);
  /** Token for step 2 of the registration (SDD 7.4.2); valid for 10 minutes. */
  const [tokenValidacion, setTokenValidacion] = useState<string | null>(null);
  const [verifiedData, setVerifiedData] = useState<ValidarIdentidadResponse['datos'] | null>(null);

  // Temporizador del bloqueo HTTP 423: arranca con Retry-After (600 s si no llega)
  const countdown = useCountdown({
    initialSeconds: DEFAULT_LOCKOUT_SECONDS,
    onFinish: () => {
      setIsLocked(false);
      setErrorMessage(null);
    },
  });

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (isLocked) return;

    let valor = valorValidacion.trim();
    if (tipoValidacion === 'FECHA_EXPEDICION') {
      const isoDate = toIsoDate(valor);
      if (!isoDate) {
        setFechaError('Ingrese una fecha válida con el formato dd/mm/aaaa.');
        return;
      }
      valor = isoDate;
    }

    setIsLoading(true);
    setErrorMessage(null);
    setFechaError(undefined);
    setVerifiedData(null);
    setTokenValidacion(null);

    try {
      // Intenta consumir el endpoint oficial de validación de identidad (SDD 7.4.1)
      // El backend envuelve la respuesta en { success, data }
      const { data } = await api.post<{ data: ValidarIdentidadResponse }>(
        '/auth/registro/validar-identidad',
        { cedula, tipoValidacion, valor }
      );

      setTokenValidacion(data.data.tokenValidacion ?? null);
      setVerifiedData(data.data.datos ?? {});
    } catch (error) {
      const normalized = normalizeError(error);

      // Si el servidor responde 423 (bloqueo por 3 intentos fallidos de validación)
      if (normalized.status === 423) {
        setIsLocked(true);
        setLockoutMessage(normalized.message);
        countdown.start(normalized.retryAfterSeconds ?? DEFAULT_LOCKOUT_SECONDS);
      } else {
        // Mensaje genérico según RF-EF11-02 para no revelar cuál dato falló
        setErrorMessage(
          normalized.message ||
          'La información ingresada es incorrecta. Verifique su cédula y dato de expedición o código dactilar.'
        );
      }
    } finally {
      setIsLoading(false);
    }
  };

  const isFormValid =
    cedula.length === 10 &&
    valorValidacion.trim().length > 0 &&
    !isLocked;

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-12">
      <div className="mx-auto max-w-md rounded-lg border border-line-strong bg-surface p-6 sm:p-8">
        <h1 className="text-xl font-bold leading-tight sm:text-2xl text-neutral-900">
          Registro de postulante
        </h1>
        <p className="mt-2 text-base text-neutral-700">
          Valide su identidad conforme a la Resolución C.D. 701 para crear su cuenta y postular al
          concurso.
        </p>

        {isLocked ? (
          <LockoutNotice
            formattedTime={countdown.formattedTime}
            isLocked={isLocked}
            title="Registro bloqueado temporalmente"
            message={lockoutMessage}
            className="mt-6"
          />
        ) : null}

        <form onSubmit={handleSubmit} className="mt-6 space-y-6" noValidate>
          <TextField
            label="Cédula de identidad"
            value={cedula}
            onChange={(e) => setCedula(e.target.value.replace(/\D/g, ''))}
            maxLength={10}
            inputMode="numeric"
            autoComplete="username"
            required
            disabled={isLocked || isLoading}
            className="tabular-nums"
            helperText="Ingrese los 10 dígitos de su cédula de identidad."
          />

          <fieldset className="space-y-3">
            <legend className="text-base font-medium text-neutral-900">
              Método de validación <span className="font-normal text-muted">(obligatorio)</span>
            </legend>
            <div className="flex flex-col gap-2 sm:flex-row sm:gap-6">
              <label className="flex items-center gap-2 text-base text-neutral-900 cursor-pointer">
                <input
                  type="radio"
                  name="tipoValidacion"
                  value="FECHA_EXPEDICION"
                  checked={tipoValidacion === 'FECHA_EXPEDICION'}
                  onChange={() => {
                    setTipoValidacion('FECHA_EXPEDICION');
                    setValorValidacion('');
                    setFechaError(undefined);
                  }}
                  disabled={isLocked || isLoading}
                  className="h-4 w-4 rounded-full border-line-input text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
                />
                Fecha de expedición
              </label>

              <label className="flex items-center gap-2 text-base text-neutral-900 cursor-pointer">
                <input
                  type="radio"
                  name="tipoValidacion"
                  value="CODIGO_DACTILAR"
                  checked={tipoValidacion === 'CODIGO_DACTILAR'}
                  onChange={() => {
                    setTipoValidacion('CODIGO_DACTILAR');
                    setValorValidacion('');
                  }}
                  disabled={isLocked || isLoading}
                  className="h-4 w-4 rounded-full border-line-input text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
                />
                Código dactilar
              </label>
            </div>
          </fieldset>

          {tipoValidacion === 'FECHA_EXPEDICION' ? (
            <TextField
              label="Fecha de expedición"
              value={valorValidacion}
              onChange={(e) => {
                setValorValidacion(e.target.value);
                setFechaError(undefined);
              }}
              error={fechaError}
              inputMode="numeric"
              placeholder="Ej. 15/06/2018"
              maxLength={10}
              required
              disabled={isLocked || isLoading}
              className="tabular-nums"
              helperText="Formato dd/mm/aaaa tal como consta en su cédula."
            />
          ) : (
            <TextField
              label="Código dactilar"
              value={valorValidacion}
              onChange={(e) => setValorValidacion(e.target.value.toUpperCase())}
              placeholder="Ej. A1234B5678"
              maxLength={10}
              required
              disabled={isLocked || isLoading}
              helperText="10 caracteres alfanuméricos ubicados al reverso de su cédula."
            />
          )}

          <Button
            type="submit"
            fullWidth
            isLoading={isLoading}
            disabled={!isFormValid}
            loadingText="Validando identidad…"
          >
            Validar identidad
          </Button>
        </form>

        {errorMessage ? (
          <Alert variant="danger" title="Error de validación" className="mt-6">
            {errorMessage}
          </Alert>
        ) : null}

        {verifiedData && tokenValidacion ? (
          <Alert variant="success" title="Identidad confirmada" className="mt-6">
            <p className="font-semibold text-neutral-900">
              {verifiedData.nombres} {verifiedData.apellidos ?? ''}
            </p>
            <p className="mt-1 text-sm text-neutral-700">
              {verifiedData.nacionalidad ? `Nacionalidad: ${verifiedData.nacionalidad}. ` : null}
              Su identidad ha sido verificada
              ante el Registro Civil. Ya puede continuar con el registro de su correo y contraseña.
            </p>
          </Alert>
        ) : null}

        <div className="mt-6 border-t border-line pt-4 text-center">
          <p className="text-sm text-muted">
            ¿Ya tiene una cuenta creada?{' '}
            <button
              type="button"
              onClick={onNavigateToLogin}
              className="font-medium text-primary underline underline-offset-4 hover:text-primary-dark"
            >
              Iniciar sesión
            </button>
          </p>
        </div>
      </div>
    </div>
  );
}

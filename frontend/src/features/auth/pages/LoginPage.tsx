import { useState } from 'react';
import type { FormEvent } from 'react';
import { Eye, EyeOff } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import TextField from '@/components/ui/TextField';
import LockoutNotice from '@/features/auth/components/LockoutNotice';
import { useCountdown } from '@/features/auth/hooks/useCountdown';
import { api, DEFAULT_LOCKOUT_SECONDS, normalizeError } from '@/services/api/client';
import { setAccessToken } from '@/store/authStore';
import { RequiredNote } from '@/components/ui/FieldMessages';

export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  postulante_id: string;
  cedula: string;
  email: string;
}

export interface LoginPageProps {
  /** Callback to switch to the registration view. */
  onNavigateToRegister?: () => void;
  /** Callback invoked upon successful authentication. */
  onLoginSuccess?: (userData: LoginResponse) => void;
  /** Notice shown above the form, e.g. after the session expired (SRP-015). */
  aviso?: string;
}

export default function LoginPage({ onNavigateToRegister, onLoginSuccess, aviso }: LoginPageProps) {
  const [cedula, setCedula] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isLocked, setIsLocked] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

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

    setIsLoading(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      const response = await api.post<{ data: LoginResponse }>('/auth/login', {
        cedula,
        password,
      });

      setSuccessMessage('Sesión iniciada correctamente. Redirigiendo a su postulación…');
      if (response.data?.data) {
        setAccessToken(response.data.data.access_token);
        onLoginSuccess?.(response.data.data);
      }
    } catch (error) {
      const normalized = normalizeError(error);

      // Bloqueo temporal por 3 intentos fallidos (HTTP 423 Locked - EF-11 / SRP-012)
      if (normalized.status === 423) {
        setIsLocked(true);
        countdown.start(normalized.retryAfterSeconds ?? DEFAULT_LOCKOUT_SECONDS);
      } else if (normalized.status === 401) {
        // Feedback visual instantáneo para credenciales erróneas (RF-EF11-05)
        setErrorMessage('Sus credenciales de ingreso son incorrectas. Revise su cédula y contraseña.');
      } else {
        setErrorMessage(normalized.message || 'Error al iniciar sesión. Intente nuevamente.');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const isFormValid = cedula.length === 10 && password.length > 0 && !isLocked;

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-12">
      <div className="mx-auto max-w-md rounded-lg border border-line-strong bg-surface p-6 sm:p-8">
        <h1 className="text-xl font-bold leading-tight text-neutral-900 sm:text-2xl">
          Iniciar sesión
        </h1>
        <p className="mt-2 text-base text-neutral-700">
          Ingrese su cédula de identidad y su contraseña para acceder al sistema de postulación.
        </p>

        {aviso ? (
          <Alert variant="info" title={aviso} className="mt-6" />
        ) : null}

        {isLocked ? (
          <LockoutNotice
            formattedTime={countdown.formattedTime}
            isLocked={isLocked}
            className="mt-6"
          />
        ) : null}

        <form onSubmit={handleSubmit} className="mt-6 space-y-6" noValidate>
          <RequiredNote />
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
            helperText="Ingrese los 10 números de su cédula de identidad."
          />

          <TextField
            label="Contraseña"
            type={showPassword ? 'text' : 'password'}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            required
            disabled={isLocked || isLoading}
            helperText="Ingrese su contraseña institucional."
            trailingAdornment={
              <button
                type="button"
                onClick={() => setShowPassword((prev) => !prev)}
                disabled={isLocked || isLoading}
                aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
                aria-pressed={showPassword}
                className="inline-flex h-11 w-11 items-center justify-center rounded-md text-muted hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark disabled:cursor-not-allowed disabled:text-muted"
              >
                {showPassword ? (
                  <EyeOff size={20} strokeWidth={2} aria-hidden="true" />
                ) : (
                  <Eye size={20} strokeWidth={2} aria-hidden="true" />
                )}
              </button>
            }
          />

          <Button
            type="submit"
            fullWidth
            isLoading={isLoading}
            disabled={!isFormValid}
            loadingText="Verificando credenciales…"
          >
            Ingresar
          </Button>
        </form>

        {errorMessage ? (
          <Alert variant="danger" title="Acceso denegado" className="mt-6">
            {errorMessage}
          </Alert>
        ) : null}

        {successMessage ? (
          <Alert variant="success" title="Acceso concedido" className="mt-6">
            {successMessage}
          </Alert>
        ) : null}

        <div className="mt-6 border-t border-line pt-4 text-center">
          <p className="text-sm text-muted">
            ¿Aún no se ha registrado?{' '}
            <button
              type="button"
              onClick={onNavigateToRegister}
              className="font-medium text-primary underline underline-offset-4 hover:text-primary-dark"
            >
              Registrarse para postular
            </button>
          </p>
        </div>
      </div>
    </div>
  );
}

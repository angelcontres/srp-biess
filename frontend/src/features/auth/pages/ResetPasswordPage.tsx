import { useState, useEffect } from 'react';
import type { FormEvent } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Eye, EyeOff } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import TextField from '@/components/ui/TextField';
import { RequiredNote } from '@/components/ui/FieldMessages';
import { api, normalizeError } from '@/services/api/client';

export interface ResetPasswordPageProps {
  onNavigateToLogin?: () => void;
}

export default function ResetPasswordPage({ onNavigateToLogin }: ResetPasswordPageProps) {
  const [searchParams] = useSearchParams();
  const [token, setToken] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  useEffect(() => {
    const queryToken = searchParams.get('token');
    if (queryToken) {
      setToken(queryToken);
    }
  }, [searchParams]);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setErrorMessage(null);

    const t = token.trim();
    if (!t) {
      setErrorMessage('Debe proporcionar el token de recuperación.');
      return;
    }

    if (password.length < 10) {
      setErrorMessage('La nueva contraseña debe tener al menos 10 caracteres.');
      return;
    }

    if (password !== confirmPassword) {
      setErrorMessage('Las contraseñas no coinciden.');
      return;
    }

    setIsLoading(true);

    try {
      await api.post('/auth/restablecer-password', {
        token: t,
        password,
      });

      setIsSuccess(true);
    } catch (error) {
      const normalized = normalizeError(error);
      setErrorMessage(normalized.message || 'No se pudo restablecer la contraseña. El enlace puede haber expirado.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-12">
      <div className="mx-auto max-w-md rounded-lg border border-line-strong bg-surface p-6 sm:p-8">
        <h1 className="text-xl font-bold leading-tight sm:text-2xl text-neutral-900">
          Restablecer contraseña
        </h1>
        <p className="mt-2 text-base text-neutral-700">
          Defina una nueva contraseña segura para acceder a su postulación.
        </p>

        {!isSuccess ? (
          <form onSubmit={handleSubmit} className="mt-6 space-y-6" noValidate>
            <RequiredNote />

            {!searchParams.get('token') ? (
              <TextField
                label="Token de recuperación"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                required
                disabled={isLoading}
                helperText="Pegue el código recibido o copie el token temporal."
              />
            ) : null}

            <TextField
              label="Nueva contraseña"
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="new-password"
              required
              disabled={isLoading}
              helperText="Mínimo 10 caracteres, combinando al menos 3 clases (mayúsculas, minúsculas, números o caracteres especiales)."
              trailingAdornment={
                <button
                  type="button"
                  onClick={() => setShowPassword((prev) => !prev)}
                  disabled={isLoading}
                  aria-label={showPassword ? 'Ocultar contraseña' : 'Mostrar contraseña'}
                  aria-pressed={showPassword}
                  className="inline-flex h-11 w-11 items-center justify-center rounded-md text-muted hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
                >
                  {showPassword ? (
                    <EyeOff size={20} strokeWidth={2} aria-hidden="true" />
                  ) : (
                    <Eye size={20} strokeWidth={2} aria-hidden="true" />
                  )}
                </button>
              }
            />

            <TextField
              label="Confirmar nueva contraseña"
              type={showConfirmPassword ? 'text' : 'password'}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
              required
              disabled={isLoading}
              helperText="Repita la contraseña para confirmar."
              trailingAdornment={
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword((prev) => !prev)}
                  disabled={isLoading}
                  aria-label={showConfirmPassword ? 'Ocultar confirmación' : 'Mostrar confirmación'}
                  aria-pressed={showConfirmPassword}
                  className="inline-flex h-11 w-11 items-center justify-center rounded-md text-muted hover:text-neutral-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-dark"
                >
                  {showConfirmPassword ? (
                    <EyeOff size={20} strokeWidth={2} aria-hidden="true" />
                  ) : (
                    <Eye size={20} strokeWidth={2} aria-hidden="true" />
                  )}
                </button>
              }
            />

            {errorMessage ? (
              <Alert variant="danger" title="Error">
                {errorMessage}
              </Alert>
            ) : null}

            <Button
              type="submit"
              fullWidth
              isLoading={isLoading}
              disabled={password.length < 10 || confirmPassword.length < 10 || !token.trim()}
              loadingText="Actualizando contraseña…"
            >
              Guardar nueva contraseña
            </Button>
          </form>
        ) : (
          <div className="mt-6 space-y-6">
            <Alert variant="success" title="Contraseña actualizada">
              Su contraseña ha sido modificada exitosamente. Ya puede iniciar sesión con sus nuevas credenciales.
            </Alert>
            <Button
              type="button"
              fullWidth
              onClick={onNavigateToLogin}
            >
              Iniciar sesión
            </Button>
          </div>
        )}

        {!isSuccess ? (
          <div className="mt-6 border-t border-line pt-4 text-center">
            <button
              type="button"
              onClick={onNavigateToLogin}
              className="text-sm font-medium text-primary underline underline-offset-4 hover:text-primary-dark"
            >
              Cancelar y volver a iniciar sesión
            </button>
          </div>
        ) : null}
      </div>
    </div>
  );
}

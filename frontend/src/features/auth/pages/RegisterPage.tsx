import { useState } from 'react';
import type { FormEvent } from 'react';
import { Eye, EyeOff, CheckCircle2, ArrowLeft } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import TextField from '@/components/ui/TextField';
import LockoutNotice from '@/features/auth/components/LockoutNotice';
import { useCountdown } from '@/features/auth/hooks/useCountdown';
import { api, DEFAULT_LOCKOUT_SECONDS, normalizeError } from '@/services/api/client';
import { RequiredMark, RequiredNote } from '@/components/ui/FieldMessages';

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

function validarPoliticaPassword(password: string, cedula: string, correo: string): string | null {
  const p = password.trim();
  if (p.length < 10 || p.length > 128) {
    return 'La contraseña debe tener al menos 10 caracteres.';
  }
  if (cedula && p.toLowerCase() === cedula.toLowerCase()) {
    return 'La contraseña no puede ser idéntica a su número de cédula.';
  }
  if (correo && p.toLowerCase() === correo.toLowerCase()) {
    return 'La contraseña no puede ser idéntica a su correo electrónico.';
  }
  let clases = 0;
  if (/[A-Z]/.test(p)) clases++;
  if (/[a-z]/.test(p)) clases++;
  if (/[0-9]/.test(p)) clases++;
  if (/[^A-Za-z0-9]/.test(p)) clases++;

  if (clases < 3) {
    return 'La contraseña debe incluir al menos 3 de las 4 clases requeridas (mayúsculas, minúsculas, números y caracteres especiales).';
  }
  return null;
}

export default function RegisterPage({ onNavigateToLogin }: RegisterPageProps) {
  // Paso 1: Validación de identidad
  const [cedula, setCedula] = useState('');
  const [tipoValidacion, setTipoValidacion] = useState<ValidationType>('FECHA_EXPEDICION');
  const [valorValidacion, setValorValidacion] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [isLocked, setIsLocked] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [lockoutMessage, setLockoutMessage] = useState<string | undefined>(undefined);
  const [fechaError, setFechaError] = useState<string | undefined>(undefined);
  const [tokenValidacion, setTokenValidacion] = useState<string | null>(null);
  const [verifiedData, setVerifiedData] = useState<ValidarIdentidadResponse['datos'] | null>(null);

  // Paso 2: Creación de cuenta
  const [correo, setCorreo] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isRegistering, setIsRegistering] = useState(false);
  const [registerSuccess, setRegisterSuccess] = useState(false);
  const [registerError, setRegisterError] = useState<string | null>(null);

  // Temporizador del bloqueo HTTP 423
  const countdown = useCountdown({
    initialSeconds: DEFAULT_LOCKOUT_SECONDS,
    onFinish: () => {
      setIsLocked(false);
      setErrorMessage(null);
    },
  });

  const handleValidarIdentidad = async (event: FormEvent<HTMLFormElement>) => {
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
    setRegisterError(null);

    try {
      const { data } = await api.post<{ data: ValidarIdentidadResponse }>(
        '/auth/registro/validar-identidad',
        { cedula, tipoValidacion, valor }
      );

      setTokenValidacion(data.data.tokenValidacion ?? null);
      setVerifiedData(data.data.datos ?? {});
    } catch (error) {
      const normalized = normalizeError(error);

      if (normalized.status === 423) {
        setIsLocked(true);
        setLockoutMessage(normalized.message);
        countdown.start(normalized.retryAfterSeconds ?? DEFAULT_LOCKOUT_SECONDS);
      } else {
        setErrorMessage(
          normalized.message ||
          'La información ingresada es incorrecta. Verifique su cédula y dato de expedición o código dactilar.'
        );
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleCrearCuenta = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!tokenValidacion) return;

    setRegisterError(null);

    const errorPolitica = validarPoliticaPassword(password, cedula, correo);
    if (errorPolitica) {
      setRegisterError(errorPolitica);
      return;
    }

    if (password !== confirmPassword) {
      setRegisterError('Las contraseñas ingresadas no coinciden.');
      return;
    }

    setIsRegistering(true);

    try {
      await api.post('/auth/registro', {
        tokenValidacion,
        correo: correo.trim().toLowerCase(),
        password,
      });

      setRegisterSuccess(true);
    } catch (error) {
      const normalized = normalizeError(error);
      setRegisterError(normalized.message || 'No se pudo crear su cuenta. Intente nuevamente.');
    } finally {
      setIsRegistering(false);
    }
  };

  const handleReiniciarPaso1 = () => {
    setTokenValidacion(null);
    setVerifiedData(null);
    setCorreo('');
    setPassword('');
    setConfirmPassword('');
    setRegisterError(null);
  };

  const isFormValidPaso1 =
    cedula.length === 10 &&
    valorValidacion.trim().length > 0 &&
    !isLocked;

  const isFormValidPaso2 =
    correo.trim().length > 0 &&
    password.length >= 10 &&
    confirmPassword.length >= 10;

  return (
    <div className="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 sm:py-12">
      <div className="mx-auto max-w-md rounded-lg border border-line-strong bg-surface p-6 sm:p-8">
        <h1 className="text-xl font-bold leading-tight sm:text-2xl text-neutral-900">
          Registro de postulante
        </h1>
        <p className="mt-2 text-base text-neutral-700">
          {tokenValidacion
            ? 'Complete sus credenciales para habilitar el acceso al sistema de postulación.'
            : 'Valide su identidad conforme a la Resolución C.D. 701 para crear su cuenta y postular al concurso.'}
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

        {/* ============================================================== */}
        {/* PASO 1: Validación de Identidad (Registro Civil)              */}
        {/* ============================================================== */}
        {!tokenValidacion ? (
          <form onSubmit={handleValidarIdentidad} className="mt-6 space-y-6" noValidate>
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
              helperText="Ingrese los 10 dígitos de su cédula de identidad."
            />

            <fieldset className="space-y-3">
              <legend className="text-base font-medium text-neutral-900">
                Método de validación<RequiredMark />
              </legend>
              <div className="flex flex-col gap-2 sm:flex-row sm:gap-6">
                <label className="flex min-h-11 cursor-pointer items-center gap-2 text-base text-neutral-900">
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

                <label className="flex min-h-11 cursor-pointer items-center gap-2 text-base text-neutral-900">
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
              disabled={!isFormValidPaso1}
              loadingText="Validando identidad…"
            >
              Validar identidad
            </Button>
          </form>
        ) : null}

        {errorMessage ? (
          <Alert variant="danger" title="Error de validación" className="mt-6">
            {errorMessage}
          </Alert>
        ) : null}

        {/* ============================================================== */}
        {/* PASO 2: Resumen de Identidad Verificada y Credenciales        */}
        {/* ============================================================== */}
        {tokenValidacion && verifiedData && !registerSuccess ? (
          <div className="mt-6 space-y-6">
            <div className="rounded-md border border-line bg-surface-muted p-4">
              <div className="flex items-start gap-3">
                <CheckCircle2 className="h-5 w-5 flex-shrink-0 text-success mt-0.5" aria-hidden="true" />
                <div className="space-y-1 text-sm">
                  <p className="font-semibold text-neutral-900">
                    {verifiedData.nombres} {verifiedData.apellidos ?? ''}
                  </p>
                  <p className="text-neutral-700">
                    Cédula: <span className="tabular-nums font-medium">{cedula}</span>
                  </p>
                  {verifiedData.nacionalidad ? (
                    <p className="text-neutral-700">Nacionalidad: {verifiedData.nacionalidad}</p>
                  ) : null}
                </div>
              </div>
            </div>

            <form onSubmit={handleCrearCuenta} className="space-y-6" noValidate>
              <RequiredNote />

              <TextField
                label="Correo electrónico"
                type="email"
                value={correo}
                onChange={(e) => setCorreo(e.target.value)}
                autoComplete="email"
                required
                disabled={isRegistering}
                placeholder="ejemplo@correo.com"
                helperText="Se enviarán las notificaciones del concurso a este correo."
              />

              <TextField
                label="Contraseña"
                type={showPassword ? 'text' : 'password'}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete="new-password"
                required
                disabled={isRegistering}
                helperText="Mínimo 10 caracteres, combinando al menos 3 clases (mayúsculas, minúsculas, números o símbolos)."
                trailingAdornment={
                  <button
                    type="button"
                    onClick={() => setShowPassword((prev) => !prev)}
                    disabled={isRegistering}
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
                label="Confirmar contraseña"
                type={showConfirmPassword ? 'text' : 'password'}
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                autoComplete="new-password"
                required
                disabled={isRegistering}
                helperText="Repita la contraseña para confirmar."
                trailingAdornment={
                  <button
                    type="button"
                    onClick={() => setShowConfirmPassword((prev) => !prev)}
                    disabled={isRegistering}
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

              {registerError ? (
                <Alert variant="danger" title="Error en registro">
                  {registerError}
                </Alert>
              ) : null}

              <div className="space-y-3">
                <Button
                  type="submit"
                  fullWidth
                  isLoading={isRegistering}
                  disabled={!isFormValidPaso2}
                  loadingText="Creando su cuenta…"
                >
                  Crear cuenta
                </Button>

                <button
                  type="button"
                  onClick={handleReiniciarPaso1}
                  disabled={isRegistering}
                  className="flex w-full items-center justify-center gap-1.5 text-sm font-medium text-neutral-700 hover:text-neutral-900 underline underline-offset-4"
                >
                  <ArrowLeft size={16} aria-hidden="true" />
                  Corregir cédula o método de validación
                </button>
              </div>
            </form>
          </div>
        ) : null}

        {/* Éxito de Registro */}
        {registerSuccess ? (
          <div className="mt-6 space-y-6">
            <Alert variant="success" title="Cuenta creada exitosamente">
              Su cuenta ha sido registrada en el Sistema de Postulación al Directorio del BIESS. Ya puede acceder con su número de cédula y la contraseña establecida.
            </Alert>
            <Button
              type="button"
              fullWidth
              onClick={onNavigateToLogin}
            >
              Iniciar sesión ahora
            </Button>
          </div>
        ) : null}

        {!registerSuccess ? (
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
        ) : null}
      </div>
    </div>
  );
}

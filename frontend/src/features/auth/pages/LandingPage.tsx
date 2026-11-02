import { useEffect } from 'react';
import { LogIn, UserPlus } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';

export interface LandingPageProps {
  /** Callback to switch to the login view. */
  onNavigateToLogin?: () => void;
  /** Callback to switch to the registration view. */
  onNavigateToRegister?: () => void;
  /**
   * Official institutional photograph supplied by BIESS. Stock images are not
   * allowed (DESIGN.md); until the asset exists the panel shows a solid color.
   */
  imageSrc?: string;
  /** Shows the institutional notice when the application phase is closed (SDD 8.5.1). */
  isPhaseClosed?: boolean;
}

export default function LandingPage({
  onNavigateToLogin,
  onNavigateToRegister,
  imageSrc,
  isPhaseClosed = false,
}: LandingPageProps) {
  useEffect(() => {
    document.title = 'Inicio | Postulación BIESS';
  }, []);

  return (
    <div className="mx-auto w-full max-w-7xl px-4 py-8 sm:px-6 sm:py-12">
      <div className="grid overflow-hidden rounded-lg border border-line-strong bg-surface lg:grid-cols-2">
        {/* Panel de presentación: fotografía institucional con filtro sólido */}
        <section
          aria-labelledby="landing-title"
          className="relative isolate flex min-h-80 flex-col justify-end bg-primary-dark p-6 text-white sm:p-8 lg:min-h-full"
        >
          {imageSrc ? (
            <>
              <img
                src={imageSrc}
                alt=""
                className="absolute inset-0 -z-20 h-full w-full object-cover grayscale"
              />
              {/* Filtro: color sólido con opacidad, sin degradado ni desenfoque */}
              <div className="absolute inset-0 -z-10 bg-primary-dark/80" aria-hidden="true" />
            </>
          ) : (
            <p className="mb-auto pb-8 text-sm text-white">
              [PENDIENTE: fotografía institucional oficial del BIESS]
            </p>
          )}

          <div className="space-y-4">
            <p className="text-base font-medium text-white">Banco del IESS</p>
            <h1 id="landing-title" className="text-3xl font-bold leading-tight text-white">
              Concurso Público de Méritos y Oposición, Directorio del BIESS
            </h1>
            <p className="text-base leading-normal text-white">
              [PENDIENTE: texto breve del proceso y requisitos del concurso]
            </p>
          </div>
        </section>

        {/* Panel de opciones de acceso */}
        <section
          aria-labelledby="landing-options-title"
          className="flex flex-col justify-center space-y-8 p-6 sm:p-8 lg:p-12"
        >
          <div className="space-y-2">
            <h2 id="landing-options-title" className="text-lg font-semibold text-neutral-900">
              ¿Cómo desea continuar?
            </h2>
            <p className="text-base text-neutral-700">
              Para postular, ingrese con su cuenta o regístrese validando su identidad con su
              cédula.
            </p>
          </div>

          {isPhaseClosed ? (
            <Alert variant="info" title="Fase de postulación cerrada">
              [PENDIENTE: mensaje institucional de fase cerrada y fechas del concurso]
            </Alert>
          ) : null}

          <div className="space-y-6">
            <div className="space-y-3">
              <h3 className="text-base font-semibold text-neutral-900">Ya tengo una cuenta</h3>
              <Button
                variant="primary"
                fullWidth
                leftIcon={<LogIn size={20} strokeWidth={2} aria-hidden="true" />}
                onClick={onNavigateToLogin}
              >
                Iniciar sesión
              </Button>
            </div>

            <div className="space-y-3 border-t border-line pt-6">
              <h3 className="text-base font-semibold text-neutral-900">Es mi primera vez</h3>
              <p className="text-sm text-muted">
                Necesitará su cédula y la fecha de expedición o el código dactilar.
              </p>
              <Button
                variant="secondary"
                fullWidth
                leftIcon={<UserPlus size={20} strokeWidth={2} aria-hidden="true" />}
                onClick={onNavigateToRegister}
              >
                Registrarse
              </Button>
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}

import { useEffect } from 'react';
import { LogIn, UserPlus, Shield, CheckCircle, Clock, FileText } from 'lucide-react';
import Alert from '@/components/ui/Alert';
import Button from '@/components/ui/Button';
import iconoBiess from '@/assets/icono-biess.svg';

export interface LandingPageProps {
  /** Callback to switch to the login view. */
  onNavigateToLogin?: () => void;
  /** Callback to switch to the registration view. */
  onNavigateToRegister?: () => void;
  /**
   * Official institutional photograph supplied by BIESS. Stock images are not
   * allowed (DESIGN.md); when provided, a dark overlay is applied for text contrast.
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
    <div className="min-h-screen lg:grid lg:grid-cols-12">
      {/* Panel izquierdo: información del concurso - fondo azul institucional */}
      <section
        aria-labelledby="landing-title"
        className="col-span-12 relative flex flex-col justify-start bg-primary-dark p-6 lg:col-span-7 lg:p-12"
      >
        {imageSrc ? (
          <>
            <img
              src={imageSrc}
              alt=""
              className="absolute inset-0 -z-20 h-full w-full object-cover grayscale"
            />
            {/* Filtro oscuro para contraste con texto blanco */}
            <div className="absolute inset-0 -z-10 bg-primary-dark/75" aria-hidden="true" />
          </>
        ) : null}

        <div className="space-y-8">
          <div className="space-y-4">
            <h1 id="landing-title" className="text-[30px] font-bold leading-[1.2] text-white">
              Concurso Público de Méritos y Oposición BIESS
            </h1>
            <p className="text-base leading-normal text-white">
              Selección de Miembros Principales y Alternos del Directorio del BIESS (Resolución No. C.D. 701)
            </p>
          </div>

          <div className="space-y-6">
            <h2 className="text-lg font-semibold text-white">Requisitos Clave</h2>
            <div className="grid gap-8 sm:grid-cols-2">
              <div className="relative rounded-md border border-line-strong bg-surface p-4 pl-20">
                <div className="absolute -left-6 top-1/2 -translate-y-1/2 flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-white border border-line-strong">
                  <Shield className="h-8 w-8 text-primary" aria-hidden="true" />
                </div>
                <div>
                  <h3 className="text-base font-semibold text-text">Nacionalidad</h3>
                  <p className="text-sm text-muted">Poseer nacionalidad ecuatoriana.</p>
                </div>
              </div>
              <div className="relative rounded-md border border-line-strong bg-surface p-4 pl-20">
                <div className="absolute -left-6 top-1/2 -translate-y-1/2 flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-white border border-line-strong">
                  <CheckCircle className="h-8 w-8 text-primary" aria-hidden="true" />
                </div>
                <div>
                  <h3 className="text-base font-semibold text-text">Derechos Políticos</h3>
                  <p className="text-sm text-muted">Estar en pleno goce de derechos de participación política y sin inhabilidades legales.</p>
                </div>
              </div>
              <div className="relative rounded-md border border-line-strong bg-surface p-4 pl-20">
                <div className="absolute -left-6 top-1/2 -translate-y-1/2 flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-white border border-line-strong">
                  <FileText className="h-8 w-8 text-primary" aria-hidden="true" />
                </div>
                <div>
                  <h3 className="text-base font-semibold text-text">Formación y Experiencia</h3>
                  <p className="text-sm text-muted">Título de Tercer o Cuarto Nivel registrado en SENESCYT y trayectoria acreditada en Finanzas, Mercado de Capitales, Administración de Empresas, Economía, Derecho o áreas afines.</p>
                </div>
              </div>
              <div className="relative rounded-md border border-line-strong bg-surface p-4 pl-20">
                <div className="absolute -left-6 top-1/2 -translate-y-1/2 flex h-16 w-16 shrink-0 items-center justify-center rounded-full bg-white border border-line-strong">
                  <Clock className="h-8 w-8 text-primary" aria-hidden="true" />
                </div>
                <div>
                  <h3 className="text-base font-semibold text-text">Condición del Sector</h3>
                  <p className="text-sm text-muted">Acreditar la calidad activa de afiliado, jubilado o empleador del IESS según el sector de postulación.</p>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-6">
            <h2 className="text-lg font-semibold text-white">Pasos del Proceso</h2>
            <ol className="space-y-6">
              <li className="flex gap-6">
                <span className="flex h-16 w-16 flex-shrink-0 items-center justify-center rounded-full bg-primary text-lg font-semibold text-white">1</span>
                <div>
                  <h3 className="text-base font-semibold text-white">Registro</h3>
                  <p className="text-sm text-white/80">Valida tu identidad ingresando tu cédula y fecha de expedición o código dactilar.</p>
                </div>
              </li>
              <li className="flex gap-6">
                <span className="flex h-16 w-16 flex-shrink-0 items-center justify-center rounded-full bg-primary text-lg font-semibold text-white">2</span>
                <div>
                  <h3 className="text-base font-semibold text-white">Formulario Digital</h3>
                  <p className="text-sm text-white/80">Completa tus datos personales, experiencia laboral y adjunta los respaldos en PDF.</p>
                </div>
              </li>
              <li className="flex gap-6">
                <span className="flex h-16 w-16 flex-shrink-0 items-center justify-center rounded-full bg-primary text-lg font-semibold text-white">3</span>
                <div>
                  <h3 className="text-base font-semibold text-white">Envío Digital</h3>
                  <p className="text-sm text-white/80">Acepta la Declaración Jurada, genera tu Código Único de Postulación y envía el formulario.</p>
                </div>
              </li>
              <li className="flex gap-6">
                <span className="flex h-16 w-16 flex-shrink-0 items-center justify-center rounded-full bg-primary text-lg font-semibold text-white">4</span>
                <div>
                  <h3 className="text-base font-semibold text-white">Entrega Física</h3>
                  <p className="text-sm text-white/80">Imprime el formulario completado, coloca tu firma manuscrita y entrégalo físicamente en las dependencias del IESS.</p>
                </div>
              </li>
            </ol>
          </div>
        </div>
      </section>

      {/* Panel derecho: opciones de acceso - fondo blanco */}
      <section
        aria-labelledby="landing-options-title"
        className="col-span-12 flex flex-col justify-center space-y-8 bg-surface p-6 lg:col-span-5 lg:p-12"
      >
        {/* Logo BIESS centrado */}
        <div className="flex flex-col items-center space-y-4">
          <img
            src={iconoBiess}
            alt="Logo oficial BIESS"
            className="h-32 w-32 object-contain"
          />
          <p className="text-[30px] font-bold text-text">BIESS</p>
        </div>

        <div className="space-y-3">
          <h2 id="landing-options-title" className="text-[24px] font-semibold text-text">
            Acceso a la Plataforma
          </h2>
          <p className="text-lg text-muted">
            Para postular, ingrese con su cuenta o regístrese validando su identidad con su
            cédula.
          </p>
        </div>

        {isPhaseClosed ? (
          <Alert variant="info" title="Fase de postulación cerrada">
            [PENDIENTE: mensaje institucional de fase cerrada y fechas del concurso]
          </Alert>
        ) : null}

        <div className="space-y-8">
          <div className="space-y-4">
            <h3 className="text-lg font-semibold text-text">Ya tengo una cuenta</h3>
            <Button
              variant="primary"
              fullWidth
              leftIcon={<LogIn size={24} strokeWidth={2} aria-hidden="true" />}
              onClick={onNavigateToLogin}
            >
              Iniciar sesión
            </Button>
          </div>

          <div className="space-y-4 border-t border-line pt-8">
            <h3 className="text-lg font-semibold text-text">Es mi primera vez</h3>
            <p className="text-base text-muted">
              Necesitará su cédula y la fecha de expedición o el código dactilar.
            </p>
            <Button
              variant="secondary"
              fullWidth
              leftIcon={<UserPlus size={24} strokeWidth={2} aria-hidden="true" />}
              onClick={onNavigateToRegister}
            >
              Registrarse
            </Button>
          </div>
        </div>
      </section>
    </div>
  );
}

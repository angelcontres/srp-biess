import { lazy, Suspense, useEffect, useRef } from 'react';
import { Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import Navbar from '@/components/layout/Navbar';
import Skeleton from '@/components/ui/Skeleton';
import LandingPage from '@/features/auth/pages/LandingPage';
import LoginPage from '@/features/auth/pages/LoginPage';
import RegisterPage from '@/features/auth/pages/RegisterPage';
import ForgotPasswordPage from '@/features/auth/pages/ForgotPasswordPage';
import ResetPasswordPage from '@/features/auth/pages/ResetPasswordPage';
import { useSession } from '@/features/auth/session/SessionProvider';
import { ProtectedRoute, PublicOnlyRoute, RUTAS } from '@/features/auth/session/routes';
import type { EstadoRedireccion } from '@/features/auth/session/routes';

// SDD 6.13: the form is code-split; it loads only after authentication.
const PostulacionPage = lazy(() => import('@/features/postulacion/pages/PostulacionPage'));
const MiPostulacionPage = lazy(() => import('@/features/postulacion/pages/MiPostulacionPage'));

const footerLinks = [
  { label: 'Ayuda', href: 'https://www.biess.fin.ec/servicio-al-cliente' },
  {
    label: 'Políticas de privacidad',
    href: 'https://www.biess.fin.ec/politicas-de-privacidad',
  },
  { label: 'Bases legales del concurso', href: 'https://www.biess.fin.ec/transparencia' },
];

const AVISO_SESION_EXPIRADA = 'Su sesión expiró. Ingrese nuevamente para continuar con su postulación.';

/** Moves focus to the new page's heading after client-side navigation (not on the first load). */
function useFocoAlNavegar() {
  const { pathname } = useLocation();
  const primeraCarga = useRef(true);
  useEffect(() => {
    if (primeraCarga.current) {
      primeraCarga.current = false;
      return;
    }
    const titulo = document.querySelector<HTMLElement>('#main-content h1');
    if (!titulo) return;
    if (!titulo.hasAttribute('tabindex')) titulo.setAttribute('tabindex', '-1');
    titulo.focus({ preventScroll: false });
  }, [pathname]);
}

function App() {
  const { user, expired, signIn, signOut } = useSession();
  const navigate = useNavigate();
  const location = useLocation();
  useFocoAlNavegar();

  const handleLogout = async () => {
    await signOut();
    navigate(RUTAS.inicio, { replace: true });
  };

  const destinoTrasIngreso = (location.state as EstadoRedireccion | null)?.desde ?? RUTAS.formulario;
  const isLandingPage = location.pathname === RUTAS.inicio;

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:bg-surface focus:p-2 focus:text-primary"
      >
        Saltar al contenido
      </a>

      {!isLandingPage && <Navbar userName={user} onLogout={() => void handleLogout()} />}

      <main id="main-content" className="flex-1">
        <Routes>
          <Route element={<PublicOnlyRoute />}>
            <Route
              path={RUTAS.inicio}
              element={
                <LandingPage
                  onNavigateToLogin={() => navigate(RUTAS.ingresar)}
                  onNavigateToRegister={() => navigate(RUTAS.registro)}
                />
              }
            />
            <Route
              path={RUTAS.ingresar}
              element={
                <LoginPage
                  aviso={expired ? AVISO_SESION_EXPIRADA : undefined}
                  onNavigateToRegister={() => navigate(RUTAS.registro)}
                  onNavigateToForgotPassword={() => navigate(RUTAS.recuperarPassword)}
                  onLoginSuccess={(datos) => {
                    navigate(destinoTrasIngreso, { replace: true });
                    signIn(datos.email);
                  }}
                />
              }
            />
            <Route path={RUTAS.registro} element={<RegisterPage onNavigateToLogin={() => navigate(RUTAS.ingresar)} />} />
            <Route
              path={RUTAS.recuperarPassword}
              element={
                <ForgotPasswordPage
                  onNavigateToLogin={() => navigate(RUTAS.ingresar)}
                  onNavigateToReset={(tok) =>
                    navigate(tok ? `${RUTAS.restablecerPassword}?token=${encodeURIComponent(tok)}` : RUTAS.restablecerPassword)
                  }
                />
              }
            />
            <Route
              path={RUTAS.restablecerPassword}
              element={<ResetPasswordPage onNavigateToLogin={() => navigate(RUTAS.ingresar)} />}
            />
          </Route>

          <Route element={<ProtectedRoute />}>
            <Route
              path={RUTAS.formulario}
              element={
                // RF-EF11-10: an authenticated applicant goes straight to the form (SRP-013).
                <Suspense
                  fallback={
                    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6">
                      <Skeleton label="Cargando su postulación" />
                    </div>
                  }
                >
                  <PostulacionPage />
                </Suspense>
              }
            />
            <Route
              path={RUTAS.miPostulacion}
              element={
                // SDD 8.5.10: state, eligibility result and, once sent, code and PDF.
                <Suspense
                  fallback={
                    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6">
                      <Skeleton label="Cargando su postulación" />
                    </div>
                  }
                >
                  <MiPostulacionPage />
                </Suspense>
              }
            />
          </Route>

          <Route path="*" element={<Navigate to={RUTAS.inicio} replace />} />
        </Routes>
      </main>

      <footer className="border-t border-line-strong bg-surface px-4 py-6 text-sm text-muted sm:px-6">
        <div className="mx-auto flex max-w-7xl flex-col gap-3">
          <nav aria-label="Enlaces institucionales">
            <ul className="flex flex-wrap gap-x-6 gap-y-2">
              {footerLinks.map((link) => (
                <li key={link.href}>
                  <a
                    href={link.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline underline-offset-4"
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </nav>
          <p>
            Banco del Instituto Ecuatoriano de Seguridad Social (BIESS). Todos los derechos
            reservados.
          </p>
        </div>
      </footer>
    </div>
  );
}

export default App;

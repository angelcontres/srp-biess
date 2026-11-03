import { Navigate, Outlet, useLocation } from 'react-router-dom';
import Skeleton from '@/components/ui/Skeleton';
import { useSession } from './SessionProvider';

export const RUTAS = {
  inicio: '/',
  ingresar: '/ingresar',
  registro: '/registro',
  recuperarPassword: '/recuperar-password',
  restablecerPassword: '/restablecer-password',
  formulario: '/formulario',
  miPostulacion: '/mi-postulacion',
} as const;

/** Navigation state: the protected page to return to after signing in. */
export interface EstadoRedireccion {
  desde?: string;
}

function CargandoSesion() {
  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6">
      <Skeleton label="Verificando su sesión" />
    </div>
  );
}

/** Pages that require a session: without one, redirects to sign in and comes back afterwards. */
export function ProtectedRoute() {
  const { status } = useSession();
  const location = useLocation();
  if (status === 'restoring') return <CargandoSesion />;
  if (status === 'anonymous') {
    const estado: EstadoRedireccion = { desde: location.pathname };
    return <Navigate to={RUTAS.ingresar} replace state={estado} />;
  }
  return <Outlet />;
}

/** Landing, sign-in and registration: a signed-in applicant goes straight to the form (RF-EF11-10). */
export function PublicOnlyRoute() {
  const { status } = useSession();
  if (status === 'restoring') return <CargandoSesion />;
  if (status === 'authenticated') return <Navigate to={RUTAS.formulario} replace />;
  return <Outlet />;
}

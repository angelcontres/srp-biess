import { LogOut } from 'lucide-react';
import { Link } from 'react-router-dom';
import { RUTAS } from '@/features/auth/session/routes';
import iconoBiess from '@/assets/icono-biess.svg';

export interface NavbarProps {
  /** Name or cédula of the authenticated user; hidden when there is no session. */
  userName?: string;
  onLogout?: () => void;
}

export default function Navbar({ userName, onLogout }: NavbarProps) {
  return (
    <>
      <header className="bg-primary text-white">
        <div className="mx-auto flex min-h-16 max-w-7xl flex-wrap items-center justify-between gap-x-4 gap-y-2 px-4 py-2 sm:flex-nowrap sm:px-6">
          <div className="flex min-w-0 items-center gap-3">
            <img
              src={iconoBiess}
              alt="Logo oficial BIESS"
              className="h-10 w-10 shrink-0 object-contain"
            />
            <div className="min-w-0">
              <span className="block text-sm font-medium leading-none text-white/80">
                Banco del IESS
              </span>
              <p className="text-balance text-base font-semibold leading-snug text-white sm:text-lg">
                Concurso Público de Méritos y Oposición
              </p>
            </div>
          </div>

          {userName ? (
            // < 640 px: link and "Salir" drop to a second row so the title keeps its width (WCAG 1.4.10).
            <div className="flex w-full items-center justify-between gap-3 sm:w-auto sm:shrink-0 sm:justify-start">
              <Link
                to={RUTAS.miPostulacion}
                className="inline-flex h-11 items-center text-base text-white underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
              >
                Mi postulación
              </Link>
              <span className="hidden text-base sm:inline">{userName}</span>
              <button
                type="button"
                onClick={onLogout}
                className="inline-flex h-11 items-center gap-2 rounded-md border border-white px-4 text-base font-semibold text-white hover:bg-primary-dark focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
              >
                <LogOut size={20} strokeWidth={2} aria-hidden="true" />
                Salir
              </button>
            </div>
          ) : null}
        </div>
      </header>
      <div className="h-1 w-full bg-secondary" aria-hidden="true" />
    </>
  );
}

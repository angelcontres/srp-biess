import { LogOut } from 'lucide-react';
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
        <div className="mx-auto flex min-h-16 max-w-7xl items-center justify-between gap-4 px-4 py-2 sm:px-6">
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
            <div className="flex shrink-0 items-center gap-3">
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

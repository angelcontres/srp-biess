import { ShieldCheck } from 'lucide-react';

export default function Navbar() {
  return (
    <header className="bg-primary text-white shadow-md">
      <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-3 sm:px-6 lg:px-8">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary font-bold text-primary shadow">
            <ShieldCheck aria-hidden="true" className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-lg font-bold leading-tight tracking-wide">BIESS</h1>
            <p className="text-xs font-medium tracking-wider text-secondary">
              BANCO DEL IESS · CONCURSO PÚBLICO
            </p>
          </div>
        </div>
        <div className="hidden text-xs text-white/70 md:block">
          Resolución No. C.D. 701 | EF-11 &amp; EF-12
        </div>
      </div>
    </header>
  );
}

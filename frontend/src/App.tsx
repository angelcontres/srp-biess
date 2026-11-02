import Navbar from '@/components/layout/Navbar';
import LoginPage from '@/features/auth/pages/LoginPage';

const footerLinks = [
  { label: 'Ayuda', href: '#ayuda' },
  { label: 'Políticas de privacidad', href: '#privacidad' },
  { label: 'Soporte técnico', href: '#soporte' },
  { label: 'Bases legales del concurso', href: '#bases-legales' },
];

function App() {
  return (
    <div className="flex min-h-screen flex-col bg-background">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:bg-surface focus:p-2 focus:text-primary"
      >
        Saltar al contenido
      </a>

      <Navbar />

      <main id="main-content" className="flex-1">
        <LoginPage />
      </main>

      <footer className="border-t border-line-strong bg-surface px-4 py-6 text-sm text-muted sm:px-6">
        <div className="mx-auto flex max-w-7xl flex-col gap-3">
          <nav aria-label="Enlaces institucionales">
            <ul className="flex flex-wrap gap-x-6 gap-y-2">
              {footerLinks.map((link) => (
                <li key={link.href}>
                  <a href={link.href} className="underline underline-offset-4">
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

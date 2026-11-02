import { useState, useEffect } from 'react';
import Navbar from '@/components/layout/Navbar';
import LoginPage from '@/features/auth/pages/LoginPage';
import RegisterPage from '@/features/auth/pages/RegisterPage';
import { clearSession } from '@/store/authStore';
import { api, refreshAccessToken } from './services/api/client';

const footerLinks = [
  { label: 'Ayuda', href: '#ayuda' },
  { label: 'Políticas de privacidad', href: '#privacidad' },
  { label: 'Soporte técnico', href: '#soporte' },
  { label: 'Bases legales del concurso', href: '#bases-legales' },
];

type AuthView = 'login' | 'register';

function App() {
  const [currentView, setCurrentView] = useState<AuthView>('login');
  const [authenticatedUser, setAuthenticatedUser] = useState<string | undefined>(undefined);
  const [isRestoring, setIsRestoring] = useState(true);


  useEffect(() => {
    (async () => {
      const token = await refreshAccessToken();
      if (token) {
        try {
          const { data } = await api.get<{ data: { email: string } }>('/auth/me');
          setAuthenticatedUser(data.data.email);
        } catch {
          // la sesión no es válida
        }
      }
      setIsRestoring(false);
    })();
  }, []);

  if (isRestoring) return null; // o un spinner, para no mostrar el login un instante



  const handleLogout = async () => {
    try {
      await api.post('/auth/logout');
    } catch {
      // aunque falle, la sesión local se limpia igual
    } finally {
      clearSession();
    }
    setAuthenticatedUser(undefined);
    setCurrentView('login');
  };

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:bg-surface focus:p-2 focus:text-primary"
      >
        Saltar al contenido
      </a>

      <Navbar userName={authenticatedUser} onLogout={handleLogout} />

      <main id="main-content" className="flex-1">
        {currentView === 'login' ? (
          <LoginPage
            onNavigateToRegister={() => setCurrentView('register')}
            onLoginSuccess={(user) => {
              if (typeof user?.email === 'string') {
                setAuthenticatedUser(user.email);
              }
            }}
          />
        ) : (
          <RegisterPage onNavigateToLogin={() => setCurrentView('login')} />
        )}
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

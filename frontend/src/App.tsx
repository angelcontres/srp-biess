import Navbar from '@/components/layout/Navbar';
import LoginPage from '@/features/auth/pages/LoginPage';

function App() {
  return (
    <div className="flex min-h-screen flex-col bg-slate-50">
      <Navbar />
      <main className="flex-1">
        <LoginPage />
      </main>
      <footer className="border-t border-slate-200 py-4 text-center text-xs text-slate-500">
        © 2026 Banco del Instituto Ecuatoriano de Seguridad Social (BIESS). Todos los derechos
        reservados.
      </footer>
    </div>
  );
}

export default App;

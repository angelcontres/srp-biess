import React from 'react';
import Navbar from './components/Navbar';
import Login from './pages/Login';

function App() {
  return (
    <div className="min-h-screen bg-slate-50 flex flex-col">
      <Navbar />
      <main className="flex-1">
        <Login />
      </main>
      <footer className="py-4 text-center text-xs text-slate-500 border-t border-slate-200">
        © 2026 Banco del Instituto Ecuatoriano de Seguridad Social (BIESS). Todos los derechos reservados.
      </footer>
    </div>
  );
}

export default App;

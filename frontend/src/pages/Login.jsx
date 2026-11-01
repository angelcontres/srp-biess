import React, { useState } from 'react';
import api from '../services/api';

export default function Login() {
  const [cedula, setCedula] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const res = await api.post('/auth/validate-cedula', { cedula });
      setResult(res);
    } catch (err) {
      setError(err?.error || 'Error al validar cédula ecuatoriana');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-64px)] flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-lg border border-slate-200 p-8 max-w-md w-full">
        <div className="text-center mb-6">
          <h2 className="text-2xl font-bold text-slate-800">Portal de Postulaciones</h2>
          <p className="text-sm text-slate-500 mt-1">
            Ingrese su cédula de identidad para verificar su registro conforme a la Resolución C.D. 701
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-semibold text-slate-700 mb-1">
              Cédula de Identidad (10 dígitos)
            </label>
            <input
              type="text"
              maxLength="10"
              value={cedula}
              onChange={(e) => setCedula(e.target.value.replace(/\D/g, ''))}
              placeholder="Ej. 0926687856"
              className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-800 focus:outline-none font-mono text-center text-lg tracking-widest"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading || cedula.length !== 10}
            className="w-full py-2.5 px-4 rounded-lg bg-blue-900 hover:bg-blue-800 text-white font-medium shadow transition duration-150 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? 'Validando Módulo 10...' : 'Verificar e Ingresar'}
          </button>
        </form>

        {result && (
          <div className="mt-4 p-3 bg-emerald-50 border border-emerald-300 rounded-lg text-emerald-800 text-sm">
            <p className="font-semibold">✓ {result.message}</p>
            <p className="text-xs text-emerald-600 mt-1">Cédula válida autorizada para postular.</p>
          </div>
        )}

        {error && (
          <div className="mt-4 p-3 bg-rose-50 border border-rose-300 rounded-lg text-rose-800 text-sm">
            <p className="font-semibold">✗ Error de Validación</p>
            <p className="text-xs text-rose-600 mt-1">{error}</p>
          </div>
        )}
      </div>
    </div>
  );
}

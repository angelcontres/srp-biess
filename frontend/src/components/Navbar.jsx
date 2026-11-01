import React from 'react';

export default function Navbar() {
  return (
    <header className="bg-gradient-to-r from-blue-900 to-indigo-950 text-white shadow-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3 flex items-center justify-between">
        <div className="flex items-center space-x-3">
          <div className="h-10 w-10 rounded-lg bg-amber-500 flex items-center justify-center font-bold text-blue-950 text-xl shadow">
            B
          </div>
          <div>
            <h1 className="font-bold text-lg leading-tight tracking-wide">BIESS</h1>
            <p className="text-xs text-amber-300 font-medium tracking-wider">BANCO DEL IESS - CONCURSO PÚBLICO</p>
          </div>
        </div>
        <div className="text-xs text-slate-300 hidden md:block">
          Resolución No. C.D. 701 | EF-11 & EF-12
        </div>
      </div>
    </header>
  );
}

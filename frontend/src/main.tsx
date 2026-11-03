import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from '@/App';
import { SessionProvider } from '@/features/auth/session/SessionProvider';
import '@/index.css';

const rootElement = document.getElementById('root');

if (!rootElement) {
  throw new Error('Root element "#root" was not found in index.html');
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <BrowserRouter>
      <SessionProvider>
        <App />
      </SessionProvider>
    </BrowserRouter>
  </React.StrictMode>,
);

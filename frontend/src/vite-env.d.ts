/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL?: string;
  /** "true" enables the development adapter of the application form (SRP-013). */
  readonly VITE_POSTULACION_MOCK?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

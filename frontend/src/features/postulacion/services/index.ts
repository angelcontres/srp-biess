import { httpGateway } from './httpGateway';
import { mockGateway } from './mockGateway';
import type { PostulacionGateway } from './gateway';

/** VITE_POSTULACION_MOCK=true selects the development adapter (never in production builds). */
export const postulacionGateway: PostulacionGateway =
  import.meta.env.DEV && import.meta.env.VITE_POSTULACION_MOCK === 'true' ? mockGateway : httpGateway;

export type { PostulacionGateway } from './gateway';
export { CONFLICTO_VERSION } from './gateway';

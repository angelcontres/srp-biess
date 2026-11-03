import { useEffect, useState } from 'react';

/** Tracks a CSS media query; false where matchMedia is unavailable (tests, SSR). */
export function useMediaQuery(query: string): boolean {
  const leer = () => typeof window !== 'undefined' && typeof window.matchMedia === 'function' && window.matchMedia(query).matches;
  const [coincide, setCoincide] = useState(leer);
  useEffect(() => {
    if (typeof window.matchMedia !== 'function') return undefined;
    const media = window.matchMedia(query);
    const actualizar = () => setCoincide(media.matches);
    actualizar();
    media.addEventListener('change', actualizar);
    return () => media.removeEventListener('change', actualizar);
  }, [query]);
  return coincide;
}

/** Tailwind `lg:` (DESIGN layout.md: stepper as a left column from 1024 px). */
export const ESCRITORIO = '(min-width: 1024px)';

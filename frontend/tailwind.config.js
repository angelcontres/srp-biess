/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    // Reemplaza la paleta por defecto: solo existen los tokens de DESIGN.md.
    colors: {
      transparent: 'transparent',
      current: 'currentColor',
      white: '#FFFFFF',
      primary: { DEFAULT: '#0B5CAD', dark: '#083F78' },
      secondary: '#F2B705',
      background: '#F8FAFC',
      surface: '#FFFFFF',
      muted: '#475569',
      neutral: {
        50: '#F8FAFC',
        100: '#F1F5F9',
        600: '#475569',
        700: '#334155',
        900: '#0F172A',
      },
      line: { DEFAULT: '#E2E8F0', strong: '#7F8DA1', input: '#64748B' },
      info: { DEFAULT: '#0B5CAD', 50: '#EFF6FF' },
      success: { DEFAULT: '#1E7B45', 50: '#ECFDF3' },
      warning: { DEFAULT: '#B45309', 50: '#FFFBEB' },
      danger: { DEFAULT: '#B42318', 50: '#FEF3F2' },
    },
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', '"Segoe UI"', 'Roboto', 'sans-serif'],
      },
      boxShadow: {
        sm: '0 1px 2px rgba(15,23,42,0.08)',
        md: '0 4px 8px rgba(15,23,42,0.12)',
        lg: '0 12px 24px rgba(15,23,42,0.18)',
      },
    },
  },
  plugins: [],
};

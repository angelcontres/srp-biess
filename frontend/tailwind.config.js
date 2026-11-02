/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: { DEFAULT: '#0B5CAD', dark: '#083F78' },
        secondary: '#F2B705',
        success: '#1E7B45',
        warning: '#B45309',
        danger: '#B42318',
        neutral: { 50: '#F8FAFC', 900: '#0F172A' },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

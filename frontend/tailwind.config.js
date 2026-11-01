/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        biess: {
          blue: '#003876',
          gold: '#C59B27',
          lightBlue: '#005CB9',
          dark: '#001A33'
        }
      }
    },
  },
  plugins: [],
}

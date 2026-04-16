/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        honeytrap: {
          bg: '#0a0a1a',
          card: '#1a1a2e',
          border: '#2a2a4a',
          green: '#4ecca3',
          red: '#e84545',
          blue: '#3282b8',
          yellow: '#f5c518',
          muted: '#6c6c8a',
          text: '#e0e0e0',
        },
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      boxShadow: {
        glow: '0 0 15px rgba(78, 204, 163, 0.15)',
        'glow-red': '0 0 15px rgba(232, 69, 69, 0.15)',
      },
    },
  },
  plugins: [],
};
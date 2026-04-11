/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        mono: ['JetBrains Mono', 'Fira Code', 'Cascadia Code', 'Consolas', 'monospace'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      colors: {
        // Dark developer theme
        surface: {
          '0': '#0d0d0f',  // deepest bg
          '1': '#111114',  // sidebar
          '2': '#18181c',  // main bg
          '3': '#1e1e24',  // cards/messages
          '4': '#26262e',  // hover states
          '5': '#2e2e38',  // borders
        },
        accent: {
          DEFAULT: '#7c6af7',  // purple
          dim: '#5a4fd1',
          muted: '#3b3260',
          text: '#a89bf8',
        },
        online: '#22c55e',
        offline: '#6b7280',
        error: '#f87171',
        warning: '#fbbf24',
      },
      animation: {
        'blink': 'blink 1s step-end infinite',
        'fade-in': 'fadeIn 0.15s ease-out',
        'slide-up': 'slideUp 0.15s ease-out',
      },
      keyframes: {
        blink: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0' },
        },
        fadeIn: {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        slideUp: {
          from: { transform: 'translateY(4px)', opacity: '0' },
          to: { transform: 'translateY(0)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}

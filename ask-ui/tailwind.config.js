/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      },
      colors: {
        brand: {
          50:  '#f4f3f9',
          100: '#e9e8f3',
          200: '#d3d1e7',
          300: '#bcb9db',
          400: '#9f9bc9',
          500: '#736EAE',
          600: '#5f5b92',
          700: '#4a4775',
          800: '#373559',
          900: '#24223c',
          DEFAULT: '#736EAE',
        },
      },
      animation: {
        'bounce-dot': 'bounce 1s infinite',
        'fade-in': 'fadeIn 0.2s ease-out',
        'slide-up': 'slideUp 0.25s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(8px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
      typography: (theme) => ({
        DEFAULT: {
          css: {
            color: theme('colors.gray.800'),
            a: {
              color: theme('colors.brand.500'),
              textDecoration: 'none',
              fontWeight: '500',
              '&:hover': {
                color: theme('colors.brand.600'),
                textDecoration: 'underline',
              },
            },
            'h1,h2,h3,h4': {
              color: theme('colors.gray.900'),
              fontWeight: '600',
            },
            code: {
              backgroundColor: theme('colors.gray.100'),
              color: theme('colors.gray.800'),
              padding: '0.2em 0.4em',
              borderRadius: '0.25rem',
              fontSize: '0.875em',
              fontWeight: '400',
            },
            'code::before': { content: '""' },
            'code::after': { content: '""' },
            blockquote: {
              borderLeftColor: theme('colors.brand.300'),
              color: theme('colors.gray.600'),
              fontStyle: 'normal',
            },
          },
        },
        sm: {
          css: {
            fontSize: '0.9rem',
            lineHeight: '1.7',
          },
        },
        dark: {
          css: {
            color: theme('colors.gray.300'),
            a: { color: theme('colors.brand.400') },
            'h1,h2,h3,h4': { color: theme('colors.gray.100'), fontWeight: '600' },
            strong: { color: theme('colors.gray.100') },
            li: { color: theme('colors.gray.300') },
            p: { color: theme('colors.gray.300') },
            code: {
              backgroundColor: theme('colors.gray.700'),
              color: theme('colors.gray.200'),
            },
            pre: {
              backgroundColor: theme('colors.gray.800'),
              borderColor: theme('colors.gray.700'),
            },
            blockquote: {
              borderLeftColor: theme('colors.brand.700'),
              color: theme('colors.gray.400'),
            },
          },
        },
      }),
    },
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}

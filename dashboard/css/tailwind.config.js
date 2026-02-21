// Configuration Tailwind CSS pour Nourriture Solidaire Dashboard
// Ce fichier définit toutes les personnalisations du thème

const config = {
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Couleurs primaires (vert - thème principal)
        primary: {
          50: '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#22c55e',
          600: '#16a34a',
          700: '#15803d',
          800: '#166534',
          900: '#14532d',
          950: '#052e16',
        },
        // Couleurs secondaires
        secondary: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
          950: '#020617',
        },
        // Couleurs de succès
        success: {
          50: '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          300: '#6ee7b7',
          400: '#34d399',
          500: '#10b981',
          600: '#059669',
          700: '#047857',
          800: '#065f46',
          900: '#064e3b',
        },
        // Couleurs d'erreur
        error: {
          50: '#fef2f2',
          100: '#fee2e2',
          200: '#fecaca',
          300: '#fca5a5',
          400: '#f87171',
          500: '#ef4444',
          600: '#dc2626',
          700: '#b91c1c',
          800: '#991b1b',
          900: '#7f1d1d',
        },
        // Couleurs d'avertissement
        warning: {
          50: '#fffbeb',
          100: '#fef3c7',
          200: '#fde68a',
          300: '#fcd34d',
          400: '#fbbf24',
          500: '#f59e0b',
          600: '#d97706',
          700: '#b45309',
          800: '#92400e',
          900: '#78350f',
        },
        // Couleurs d'information
        info: {
          50: '#eff6ff',
          100: '#dbeafe',
          200: '#bfdbfe',
          300: '#93c5fd',
          400: '#60a5fa',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
          800: '#1e40af',
          900: '#1e3a8a',
        },
        // Couleurs de fond
        background: {
          light: '#ffffff',
          DEFAULT: '#f9fafb',
          dark: '#111827',
          darker: '#0f172a',
        },
        // Couleurs de surface
        surface: {
          light: '#ffffff',
          DEFAULT: '#f3f4f6',
          dark: '#1f2937',
          darker: '#111827',
        },
        // Couleurs de texte
        text: {
          primary: '#111827',
          secondary: '#4b5563',
          muted: '#9ca3af',
          light: '#d1d5db',
          inverse: '#ffffff',
        },
        // Couleurs de bordure
        border: {
          light: '#e5e7eb',
          DEFAULT: '#d1d5db',
          dark: '#374151',
          darker: '#1f2937',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '0.875rem' }],
      },
      spacing: {
        '4.5': '1.125rem',
        '5.5': '1.375rem',
        '6.5': '1.625rem',
        '7.5': '1.875rem',
        '8.5': '2.125rem',
        '9.5': '2.375rem',
        '10.5': '2.625rem',
        '11': '2.75rem',
        '11.5': '2.875rem',
        '12.5': '3.125rem',
        '13': '3.25rem',
        '14.5': '3.625rem',
        '15': '3.75rem',
        '18': '4.5rem',
        '22': '5.5rem',
        '25': '6.25rem',
        '30': '7.5rem',
        '35': '8.75rem',
        '45': '11.25rem',
        '50': '12.5rem',
        '55': '13.75rem',
        '60': '15rem',
        '65': '16.25rem',
        '70': '17.5rem',
        '75': '18.75rem',
        '90': '22.5rem',
        '94': '23.5rem',
        '100': '25rem',
        '115': '28.75rem',
        '125': '31.25rem',
        '150': '37.5rem',
        '180': '45rem',
        '203': '50.75rem',
      },
      maxWidth: {
        '2.5xl': '45rem',
        '8xl': '90rem',
      },
      minWidth: {
        '22.5': '5.625rem',
        '75': '18.75rem',
      },
      zIndex: {
        1: '1',
        99: '99',
        999: '999',
        9999: '9999',
        99999: '99999',
      },
      opacity: {
        15: '0.15',
        35: '0.35',
        65: '0.65',
      },
      boxShadow: {
        'card': '0px 1px 3px rgba(0, 0, 0, 0.08)',
        'card-hover': '0px 4px 12px rgba(0, 0, 0, 0.12)',
        'dropdown': '0px 4px 16px rgba(0, 0, 0, 0.12)',
        'modal': '0px 8px 32px rgba(0, 0, 0, 0.16)',
        'sidebar': '4px 0px 16px rgba(0, 0, 0, 0.08)',
        'input': '0px 1px 2px rgba(0, 0, 0, 0.05)',
        'input-focus': '0px 0px 0px 3px rgba(34, 197, 94, 0.15)',
        'button': '0px 1px 2px rgba(0, 0, 0, 0.05)',
        'button-hover': '0px 2px 4px rgba(0, 0, 0, 0.1)',
      },
      borderRadius: {
        'sm': '0.25rem',
        'DEFAULT': '0.375rem',
        'md': '0.5rem',
        'lg': '0.75rem',
        'xl': '1rem',
        '2xl': '1.25rem',
        '3xl': '1.5rem',
      },
      transitionDuration: {
        '250': '250ms',
        '350': '350ms',
        '400': '400ms',
      },
      animation: {
        'spin-slow': 'spin 3s linear infinite',
        'pulse-slow': 'pulse 3s ease-in-out infinite',
        'bounce-slow': 'bounce 2s infinite',
        'fade-in': 'fadeIn 0.3s ease-out',
        'fade-out': 'fadeOut 0.3s ease-out',
        'slide-in-up': 'slideInUp 0.3s ease-out',
        'slide-in-down': 'slideInDown 0.3s ease-out',
        'slide-in-left': 'slideInLeft 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'shake': 'shake 0.5s ease-in-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        fadeOut: {
          '0%': { opacity: '1' },
          '100%': { opacity: '0' },
        },
        slideInUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideInLeft: {
          '0%': { transform: 'translateX(-10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideInRight: {
          '0%': { transform: 'translateX(10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        shake: {
          '0%, 100%': { transform: 'translateX(0)' },
          '10%, 30%, 50%, 70%, 90%': { transform: 'translateX(-4px)' },
          '20%, 40%, 60%, 80%': { transform: 'translateX(4px)' },
        },
      },
    },
  },
};

// Export pour utilisation avec Tailwind CLI ou build tools
if (typeof module !== 'undefined') {
  module.exports = config;
}

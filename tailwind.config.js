/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html"],
  theme: {
    extend: {
      colors: {
        pulse: {
          black: "#000000",
          dark: "#121212",
          "bg-start": "#1A1A1A",
          volt: "#CCFF00",
          gray: "#E0E0E0",
          muted: "rgba(224, 224, 224, 0.4)",
          card: "#181818",
        },
      },
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};

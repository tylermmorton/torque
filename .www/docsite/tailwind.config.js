/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: "jit",
  darkMode: "class",
  content: ["./**/*.{html,ts,tsx,js,jsx}"],
  plugins: [require("@tailwindcss/typography")],
  theme: {
    extend: {
      colors: {
        black: "#0A0A0F",
        white: "#f8f1ff",
        raisin: "#171623",
        "raisin-dark": "#13121c",
        "raisin-light": "#211f32",
        "raisin-hover": "#3b3b54",
        "raisin-border": "#4C4C6B",
        "raisin-warning": "#EEC170",
        "raisin-error": "#ED474A",
      },
      fontFamily: {
        "fira-code": ["Fira Code", "monospace"],
      },
    },
  },
};

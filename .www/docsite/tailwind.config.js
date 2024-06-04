/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: "jit",
  darkMode: "class",
  content: ["./**/*.{html,vue,ts,tsx,js,jsx}"],
  plugins: [require("@tailwindcss/typography")],
};

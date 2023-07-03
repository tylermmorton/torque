/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: 'jit',
  content: ["./**/*.{html,vue,ts,tsx,js,jsx}"],
  plugins: [require("@tailwindcss/typography")],
};

/** @type {import('tailwindcss').Config} */
module.exports = {
  mode: 'jit',
  purge: ["./**/*.{html,vue,ts,tsx,js,jsx}"],
  content: ["./**/*.{html,vue,ts,tsx,js,jsx}"],
  plugins: [require("@tailwindcss/typography")],
};

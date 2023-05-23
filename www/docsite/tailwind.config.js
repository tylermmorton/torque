/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./**/*.{html,vue,ts,tsx,js,jsx}"],
  theme: {
    extend: {
      colors: {
        text: "#d0d0d9",
        background: "#191f2f",
        highlight: "#a061c5",
        lowlight: "#4b465e",
        border: "#3f3f3f",
      },

      borderColor: (theme) => ({
        DEFAULT: "#3f3f3f",
        ...theme("colors"),
      }),

      typography: {
        DEFAULT: {
          css: {
            color: "#d0d0d9",
            a: {
              color: "#d0d0d9",
              "&:hover": {
                color: "#9a5fc3",
              },
            },
            "--tw-prose-code": "#92be7c",
            "--tw-prose-headings": "#d0d0d9",
          },
        },
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
};

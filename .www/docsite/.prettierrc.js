/** @type {import("prettier").Config} */
const config = {
  overrides: [
    {
      files: "*.tmpl.html",
      parser: "go-template",
    },
  ],
};

module.exports = config;

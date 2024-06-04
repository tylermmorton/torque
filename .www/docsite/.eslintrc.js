module.exports = {
  overrides: [
    {
      // Configuration for .tmpl.html files to be used in conjunction with
      // prettier and the prettier-plugin-go-template defaults
      // https://yeonjuan.github.io/html-eslint/docs
      files: ["*.tmpl.html"],
      parser: "@html-eslint/parser",
      plugins: ["@html-eslint" /* "@htmx-eslint" */],
      extends: ["plugin:@html-eslint/recommended"],
      rules: {
        // all of these styling rules are handled by prettier
        "@html-eslint/element-newline": "off",
        "@html-eslint/indent": "off",
        "@html-eslint/no-extra-spacing-attrs": "off",
        "@html-eslint/require-closing-tags": [
          "error",
          { selfClosing: "always" },
        ],

        // actual linting rules
        "@html-eslint/quotes": ["error", "double"],
        "@html-eslint/id-naming-convention": ["error", "kebab-case"],
        "@html-eslint/require-doctype": "off",
        "@html-eslint/require-img-alt": "warn",
      },
    },
  ],
};

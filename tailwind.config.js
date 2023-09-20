/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.html"],
  safelist: [
    {
      pattern: /text-.+-(|500|700)/,
      variants: ["hover"],
    },
  ],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/forms")],
};

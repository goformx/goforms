import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import json from "@eslint/json";
import css from "@eslint/css";
import { defineConfig } from "eslint/config";
import prettier from "eslint-plugin-prettier";

export default defineConfig([
  {
    files: ["**/*.{js,mjs,cjs,ts}"],
    plugins: { js },
    extends: ["js/recommended"],
  },
  {
    files: ["**/*.{js,mjs,cjs,ts}"],
    languageOptions: { globals: globals.browser },
  },
  tseslint.configs.recommended,
  {
    files: ["**/*.json"],
    plugins: { json },
    language: "json/json",
    extends: ["json/recommended"],
  },
  {
    files: ["**/*.jsonc"],
    plugins: { json },
    language: "json/jsonc",
    extends: ["json/recommended"],
  },
  {
    files: ["**/*.css"],
    plugins: { css },
    language: "css/css",
    rules: {
      "css/no-invalid-at-rules": "off",
    },
  },
  // Configuration files - relaxed rules
  {
    files: ["*.config.{js,ts}", "eslint.config.js", "vite.config.ts", "vitest.config.ts"],
    languageOptions: {
      ecmaVersion: "latest",
      sourceType: "module",
    },
    rules: {
      "prettier/prettier": "error",
      "@typescript-eslint/no-unused-vars": "off",
      "no-console": "off",
    },
  },
  // Source files - strict rules
  {
    files: ["src/**/*.{ts,tsx,js,jsx}"],
    languageOptions: {
      // Use latest ECMAScript features (ES2024+)
      ecmaVersion: "latest",
      sourceType: "module",
      parserOptions: {
        // Enable all modern TypeScript features
        project: "./tsconfig.json",
        tsconfigRootDir: ".",
      },
    },
    plugins: {
      css,
      prettier,
    },
    rules: {
      // Code formatting
      "prettier/prettier": [
        "error",
        {
          quoteProps: "preserve",
        },
      ],

      // Modern JavaScript/TypeScript standards
      "@typescript-eslint/no-namespace": "error", // Prefer ES modules over namespaces
      "@typescript-eslint/prefer-namespace-keyword": "off", // Disable in favor of ES modules

      // Variable handling
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
      "@typescript-eslint/no-explicit-any": "off",

      // Modern JavaScript features
      "prefer-const": "error", // Use const by default
      "no-var": "error", // Prefer let/const over var
      "object-shorthand": "error", // Use shorthand object properties
      "prefer-template": "error", // Use template literals over string concatenation

      // Code quality
      "no-console": "warn", // Warn about console usage in production code
      "no-debugger": "error", // Prevent debugger statements
    },
  },
]);

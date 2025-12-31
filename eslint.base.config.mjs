import js from "@eslint/js";
import tseslint from "typescript-eslint";

/**
 * Base ESLint configuration for the monorepo
 * This config provides shared rules and settings for all workspaces
 */
export const baseConfig = [
  {
    ignores: [
      "**/node_modules/**",
      "**/dist/**",
      "**/.next/**",
      "**/out/**",
      "**/build/**",
      "**/.source/**",
      "**/next-env.d.ts",
      "**/*.mdx",
      "**/eslint.config.js",
      "**/eslint.config.mjs",
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
];

/**
 * TypeScript configuration for projects
 * Use this for apps and packages that use TypeScript
 */
export const typescriptConfig = {
  files: ["**/*.{ts,tsx}"],
  languageOptions: {
    parserOptions: {
      projectService: true,
      tsconfigRootDir: import.meta.dirname,
    },
  },
  rules: {
    "@typescript-eslint/no-explicit-any": "warn",
    "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
  },
};

/**
 * React configuration for React-based projects
 */
export const reactConfig = (reactHooks, reactRefresh) => ({
  files: ["**/*.{ts,tsx}"],
  plugins: {
    "react-hooks": reactHooks,
    "react-refresh": reactRefresh,
  },
  rules: {
    ...reactHooks.configs.recommended.rules,
    "react-refresh/only-export-components": [
      "warn",
      { allowConstantExport: true },
    ],
  },
});

import { baseConfig, typescriptConfig } from "./eslint.base.config.mjs";

/**
 * Root ESLint configuration for the monorepo
 * This config applies to the root workspace ONLY
 *
 * IMPORTANT: Workspace files (apps/*, packages/*) should NOT be linted by this config.
 * Each workspace has its own eslint.config.js that extends the base configuration.
 */
export default [
  ...baseConfig,
  {
    // Ignore all workspace directories - they have their own configs
    ignores: [
      "apps/**",
      "packages/**",
    ],
  },
  {
    ...typescriptConfig,
    languageOptions: {
      ...typescriptConfig.languageOptions,
      parserOptions: {
        ...typescriptConfig.languageOptions.parserOptions,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
];

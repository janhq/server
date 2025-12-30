import { baseConfig, typescriptConfig } from '../../eslint.base.config.mjs';

/**
 * ESLint configuration for apps/platform
 * This is a Next.js/Node.js platform service
 */
export default [
  ...baseConfig,
  {
    ...typescriptConfig,
    languageOptions: {
      ...typescriptConfig.languageOptions,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
];

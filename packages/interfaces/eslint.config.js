import globals from "globals";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import {
  baseConfig,
  typescriptConfig,
  reactConfig,
} from "../../eslint.base.config.mjs";

/**
 * ESLint configuration for packages/ui
 * This is a shared React component library
 */
export default [
  ...baseConfig,
  {
    ...typescriptConfig,
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    ...reactConfig(reactHooks, reactRefresh),
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
  },
];

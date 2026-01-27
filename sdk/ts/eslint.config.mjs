import globals from "globals";
import jsLint from "@eslint/js";
import tsLint from "typescript-eslint";
import pluginSimpleImportSort from "eslint-plugin-simple-import-sort";
import stylistic from '@stylistic/eslint-plugin';
import { globalIgnores } from 'eslint/config';
import importPlugin from 'eslint-plugin-import';

export default tsLint.config(
  globalIgnores(["./src/generated/"]),
  {
    plugins: {
      // key "simple-import-sort" is the plugin namespace
      "simple-import-sort": pluginSimpleImportSort,
      '@stylistic': stylistic,
    },
    rules: {
      "simple-import-sort/imports": [
        "error",
      ]
    }
  },
  {
    rules: {
      "@typescript-eslint/consistent-type-imports": [
        "error",
        {
          prefer: "type-imports",
          fixStyle: "separate-type-imports"
        }
      ],
      '@typescript-eslint/no-unused-vars': ['error', { varsIgnorePattern: '^_$' }],
    }
  },
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node
      }
    }
  },
  jsLint.configs.recommended,
  tsLint.configs.eslintRecommended,
  ...tsLint.configs.recommended,
  stylistic.configs.customize({
    indent: 2,
    quotes: "double",
    semi: true,
    braceStyle: "1tbs",
    commaDangle: "always-multiline",
    arrayBracketSpacing: true,
    quoteProps: "as-needed",
    arrowParens: "always",
    blockSpacing: true,
  }),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      importPlugin.flatConfigs.recommended,
      importPlugin.flatConfigs.typescript
    ],
    rules: {
      'import/no-unresolved': 'off',
      'import/extensions': ['error','ignorePackages'],
    }
  },
);

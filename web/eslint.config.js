// @ts-check
const tseslint = require('typescript-eslint');
const tseslintPlugin = require('@typescript-eslint/eslint-plugin');
const angular = require('angular-eslint');

module.exports = tseslint.config(
  {
    ignores: [
      'projects/**/*',
      '**/node_modules/**',
      '**/dist/**',
      '**/coverage/**',
      '**/out-tsc/**',
      'src/assets/monaco-editor/**',
    ],
  },
  {
    files: ['**/*.ts'],
    extends: [...angular.configs.tsRecommended],
    plugins: {
      '@typescript-eslint': tseslintPlugin,
    },
    languageOptions: {
      parserOptions: {
        project: ['tsconfig.eslint.json'],
        createDefaultProgram: true,
      },
    },
    processor: angular.processInlineTemplates,
    rules: {
      '@angular-eslint/component-selector': 'off',
      '@angular-eslint/directive-selector': 'off',
      '@angular-eslint/prefer-standalone': 'off',
      '@angular-eslint/prefer-inject': 'off',
      '@angular-eslint/prefer-on-push-component-change-detection': 'off',
      '@angular-eslint/no-output-on-prefix': 'off',
      '@angular-eslint/no-output-native': 'off',
      '@angular-eslint/no-empty-lifecycle-method': 'off',
      '@typescript-eslint/no-explicit-any': 'warn',
      '@typescript-eslint/no-unused-vars': 'warn',
      '@typescript-eslint/no-empty-function': 'off',
      '@typescript-eslint/no-empty-interface': 'warn',
      '@typescript-eslint/no-non-null-assertion': 'off',
      '@typescript-eslint/no-inferrable-types': 'off',
      '@typescript-eslint/triple-slash-reference': 'off',
      'no-empty': 'off',
    },
  },
  {
    files: ['**/*.html'],
    extends: [...angular.configs.templateRecommended],
    rules: {
      '@angular-eslint/template/eqeqeq': 'off',
      '@angular-eslint/template/prefer-control-flow': 'off',
    },
  }
);

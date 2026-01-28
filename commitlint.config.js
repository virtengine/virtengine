// VirtEngine Commitlint Configuration
// See: https://commitlint.js.org/
// Format: type(scope): description

module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // Enforce conventional commit types
    'type-enum': [
      2,
      'always',
      [
        'feat',     // New feature
        'fix',      // Bug fix
        'docs',     // Documentation only changes
        'style',    // Changes that do not affect the meaning of the code
        'refactor', // Code change that neither fixes a bug nor adds a feature
        'perf',     // Performance improvement
        'test',     // Adding missing tests or correcting existing tests
        'build',    // Changes that affect the build system or external dependencies
        'ci',       // Changes to CI configuration files and scripts
        'chore',    // Other changes that don't modify src or test files
        'revert',   // Reverts a previous commit
      ],
    ],
    // VirtEngine module scopes
    'scope-enum': [
      1, // Warning level - scope is optional
      'always',
      [
        // Blockchain modules (x/)
        'veid',
        'mfa',
        'encryption',
        'market',
        'escrow',
        'roles',
        'hpc',
        // Infrastructure
        'provider',
        'sdk',
        'cli',
        'app',
        // Development
        'deps',
        'ci',
        'api',
        'ml',
        'tests',
      ],
    ],
    // Subject must start with lowercase
    'subject-case': [2, 'always', 'lower-case'],
    // Subject must not be empty
    'subject-empty': [2, 'never'],
    // Subject must not end with period
    'subject-full-stop': [2, 'never', '.'],
    // Type must be lowercase
    'type-case': [2, 'always', 'lower-case'],
    // Type must not be empty
    'type-empty': [2, 'never'],
    // Header max length
    'header-max-length': [2, 'always', 100],
    // Body max line length
    'body-max-line-length': [1, 'always', 100],
  },
  helpUrl: 'https://www.conventionalcommits.org/en/v1.0.0/',
};

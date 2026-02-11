const MAP_ALIASES = {
  '^@test/(.*)$': '<rootDir>/test/$1',
};

const common = {
  extensionsToTreatAsEsm: ['.ts'],
  transform: {
    '^.+\\.ts$': ['ts-jest', { tsconfig: './tsconfig.spec.json', useESM: true }],
  },
  rootDir: '.',
  moduleNameMapper: {
    ...MAP_ALIASES,
    '^(\\.{1,2}/.*)\\.js$': '$1',
  },
  resolver: 'ts-jest-resolver',
  watchPathIgnorePatterns: ['<rootDir>/node_modules/.tmp'],
  testEnvironment: 'node',
  setupFilesAfterEnv: ['<rootDir>/test/jest.setup.ts'],
};

module.exports = {
  collectCoverageFrom: ['<rootDir>/src/**/*.{js,ts}', '!<rootDir>/src/**/*.spec.ts'],
  projects: [
    {
      displayName: 'unit',
      ...common,
      testMatch: ['<rootDir>/src/**/*.spec.ts'],
    },
    {
      displayName: 'functional',
      ...common,
      testMatch: ['<rootDir>/test/functional/**/*.spec.ts'],
    },
  ],
};

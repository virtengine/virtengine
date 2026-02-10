import type { Config } from "jest";

const MAP_ALIASES = {
  "^@test/(.*)$": "<rootDir>/test/$1",
};

const common = {
  transform: {
    "^.+\\.(t|j)s$": ["ts-jest", { tsconfig: "./tsconfig.spec.json" }],
  } as Config["transform"],
  rootDir: ".",
  moduleNameMapper: {
    ...MAP_ALIASES,
  },
  resolver: "ts-jest-resolver",
  watchPathIgnorePatterns: [
    "<rootDir>/node_modules/.tmp",
  ],
};

export default {
  collectCoverageFrom: [
    "<rootDir>/src/**/*.{js,ts}",
    "!<rootDir>/src/**/*.spec.ts",
  ],
  projects: [
    {
      displayName: "unit",
      ...common,
      testMatch: ["<rootDir>/src/**/*.spec.ts"],
    },
    {
      displayName: "functional",
      ...common,
      testMatch: ["<rootDir>/test/functional/**/*.spec.ts"],
    },
  ],
} satisfies Config;

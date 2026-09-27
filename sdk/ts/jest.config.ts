import type { Config } from "jest";

const MAP_ALIASES = {
  "^@test/(.*)$": "<rootDir>/test/$1",
};

const common = {
  transform: {
    // SECURITY (node-vuln-scan gate): @faker-js/faker>=10 is ESM-only
    // ("type": "module") while jest runs CJS, and a faker-scoped ts-jest
    // entry does not work — ts-jest transformer instances share one static
    // config-set cache keyed only on the project config, so the faker entry
    // silently reuses the project (NodeNext) config and emits ESM back.
    // This repo-owned transformer compiles just faker's `.js` to CJS with
    // the TypeScript compiler already in devDependencies (first match wins).
    "@faker-js[/\\\\]faker[/\\\\].+\\.js$": "<rootDir>/test/jest-faker-transform.cjs",
    "^.+\\.(t|j)s$": ["ts-jest", { tsconfig: "./tsconfig.spec.json" }],
  } as Config["transform"],
  rootDir: ".",
  moduleNameMapper: {
    ...MAP_ALIASES,
  },
  resolver: "ts-jest-resolver",
  // Faker is transpiled via the entry above, so it must not be excluded here;
  // everything else in node_modules stays excluded. (Character classes keep
  // this matching on both POSIX and Windows separators.)
  transformIgnorePatterns: ["<rootDir>[/\\\\]node_modules[/\\\\](?!@faker-js[/\\\\]faker[/\\\\])"],
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

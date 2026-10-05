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
    "^.+\\.(t|j)s$": ["ts-jest", { tsconfig: "./tsconfig.spec.cjs.json" }],
  } as Config["transform"],
  rootDir: ".",
  moduleNameMapper: {
    ...MAP_ALIASES,
  },
  resolver: "<rootDir>/test/jest-resolver.cjs",
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
      // t_631af860: the leading "**/" keeps this pattern ROOT-RELATIVE so jest never
      // interpolates an absolute <rootDir> prefix into it. That interpolation is what
      // breaks discovery on a Windows checkout path containing a glob-special segment
      // (e.g. `virtengine/.worktrees/<name>/`): jest-config escapes rootDir as a glob
      // and only THEN runs replacePathSepForGlob, whose lookahead `\\(?![$()+.?^{}])`
      // cannot fire on the backslash that now precedes a glob escape (`\.worktrees`).
      // The root's separators survive as literal backslashes, the absolute pattern ends
      // up MIXED-separator, and it matches no crawled file -- so `--listTests` reports
      // 0 of 40 specs while 40 spec files are present on disk.
      //
      // Measured on jest 30.5.2 at develop 006129710, from the same clean worktree:
      //   `{"rootDir":"."}`                              -> 43 discovered
      //   `{"rootDir":".","testMatch":["**/*.spec.ts"]}`  -> 43
      //   `{"rootDir":".","testMatch":["**/src/**/*.spec.ts"]}` -> 40  (unit suite)
      //   `{"rootDir":".","testMatch":["<rootDir>/src/**/*.spec.ts"]}` -> 0  (broken)
      //   `{"rootDir":".","testMatch":["src/**/*.spec.ts"]}`            -> 0  (broken)
      // A root-relative pattern has no interpolated separator to survive, so it
      // resolves identically on plain and dot-segment paths. scripts/ci/check-jest-spec-discovery.sh
      // asserts this from both path shapes so it cannot regress silently.
      testMatch: ["**/src/**/*.spec.ts"],
    },
    {
      displayName: "functional",
      ...common,
      // Root-relative for the same reason as the unit project above (t_631af860).
      testMatch: ["**/test/functional/**/*.spec.ts"],
    },
  ],
} satisfies Config;

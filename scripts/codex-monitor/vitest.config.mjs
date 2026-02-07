import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    globals: true,
    include: ["tests/**/*.test.mjs"],
    testTimeout: 5000,
  },
});

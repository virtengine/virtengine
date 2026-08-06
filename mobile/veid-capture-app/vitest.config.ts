const { defineConfig } = require("vitest/config");

module.exports = defineConfig({
  test: {
    environment: "node",
    include: ["src/test/**/*.test.ts"],
    passWithNoTests: false
  }
});

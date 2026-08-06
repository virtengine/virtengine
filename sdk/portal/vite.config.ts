import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import wasm from "vite-plugin-wasm";
import topLevelAwait from "vite-plugin-top-level-await";

export default defineConfig({
  plugins: [react(), wasm(), topLevelAwait()],
  server: {
    host: true,
    port: 5173,
  },
  preview: {
    host: true,
    port: 5173,
  },
  build: {
    chunkSizeWarningLimit: 9000,
    rollupOptions: {
      onwarn(warning, defaultHandler) {
        const suppressed = [
          "externalized for browser compatibility",
          "contains an annotation that Rollup cannot interpret",
          "Use of eval in",
        ];
        if (suppressed.some((message) => warning.message.includes(message))) {
          return;
        }
        defaultHandler(warning);
      },
    },
  },
});

import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
  },
  preview: {
    host: true,
    port: 5173,
  },
  build: {
    chunkSizeWarningLimit: 8000,
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

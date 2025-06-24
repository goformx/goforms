import { defineConfig } from "vitest/config";
import { pathAliases } from "./vite.config";

export default defineConfig({
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/js/test/setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      exclude: [
        "node_modules/",
        "dist/",
        "**/*.d.ts",
        "**/*.config.*",
        "**/test/**",
        "**/mocks/**",
      ],
    },
  },
  resolve: {
    alias: pathAliases,
  },
});

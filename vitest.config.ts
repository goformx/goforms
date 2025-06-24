import { defineConfig } from "vitest/config";
import { resolve } from "path";

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
    alias: {
      "@": resolve(__dirname, "src/js"),
      "@/core": resolve(__dirname, "src/js/core"),
      "@/features": resolve(__dirname, "src/js/features"),
      "@/shared": resolve(__dirname, "src/js/shared"),
      "@/pages": resolve(__dirname, "src/js/pages"),
    },
  },
});

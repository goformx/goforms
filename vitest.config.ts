import { defineConfig } from "vitest/config";
import { pathAliases } from "./vite.config";

export default defineConfig({
  test: {
    // Environment configuration
    environment: "jsdom",
    globals: true,

    // Mock management - reset mocks between tests for cleaner test isolation
    clearMocks: true,
    mockReset: false, // Keep implementations but clear history
    restoreMocks: false, // Don't restore original implementations
    unstubEnvs: true, // Reset environment variables between tests

    // Performance and timeout configuration
    testTimeout: 10000, // 10 seconds for individual tests
    hookTimeout: 10000, // 10 seconds for hooks (beforeEach, afterEach, etc.)

    // Test file patterns
    include: [
      "src/**/*.{test,spec}.{js,ts,tsx}",
      "src/**/__tests__/**/*.{js,ts,tsx}",
    ],
    exclude: [
      "**/node_modules/**",
      "**/dist/**",
      "**/coverage/**",
      "**/.{idea,git,cache,output,temp}/**",
      "**/e2e/**", // Exclude e2e tests from unit test runs
    ],

    // Watch mode configuration for better DX
    watch: true,

    // Sequence configuration for consistent test runs
    sequence: {
      concurrent: false, // Run tests sequentially for more predictable results
      shuffle: false, // Don't randomize test order
      hooks: "stack", // Run hooks in stack order (after hooks in reverse)
    },

    // Coverage configuration with meaningful thresholds
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html", "lcov"],

      // Include patterns - focus on source code
      include: ["src/**/*.{js,ts,tsx,vue}", "!src/**/*.d.ts"],

      // Comprehensive exclude patterns
      exclude: [
        "node_modules/",
        "dist/",
        "coverage/",
        "**/*.d.ts",
        "**/*.config.*",
        "**/test/**",
        "**/tests/**",
        "**/__tests__/**",
        "**/*.test.*",
        "**/*.spec.*",
        "**/mocks/**",
        "src/vite-plugin-ejs.ts", // Build tools
      ],

      // Coverage thresholds - start reasonable and improve over time
      thresholds: {
        // Global thresholds
        lines: 70,
        functions: 70,
        branches: 60,
        statements: 70,

        // Automatically update thresholds when coverage improves
        autoUpdate: true,

        // Check thresholds per file for granular control
        perFile: true,

        // Specific file patterns with higher requirements
        "src/lib/**/*.ts": {
          lines: 90,
          functions: 90,
          branches: 80,
          statements: 90,
        },

        "src/composables/**/*.ts": {
          lines: 85,
          functions: 85,
          branches: 75,
          statements: 85,
        },
      },

      // Coverage output configuration
      reportsDirectory: "./coverage",
      clean: true,
      cleanOnRerun: true,

      // Advanced coverage options
      all: true, // Include all files, even untested ones
      skipFull: false, // Show files with 100% coverage
      reportOnFailure: true, // Generate reports even if tests fail
    },

    // Environment variables for tests
    env: {
      NODE_ENV: "test",
      VITEST: "true",
    },

    // Advanced configuration
    isolate: true, // Run tests in isolated environments
    pool: "threads", // Use threads for better performance
    poolOptions: {
      threads: {
        singleThread: false, // Use multiple threads when possible
        isolate: true, // Isolate each test file
      },
    },

    // Debugging and development
    logHeapUsage: false, // Enable for memory leak debugging
    silent: false, // Show console outputs

    // Retry configuration for flaky tests
    retry: 0, // Number of retries for failing tests

    // Snapshot configuration
    snapshotFormat: {
      printBasicPrototype: false,
      escapeRegex: true,
    },

    // Reporter configuration
    reporters: ["default", "html"],
    outputFile: {
      html: "./coverage/test-results.html",
      json: "./coverage/test-results.json",
    },

    // Chai configuration for better test output
    chaiConfig: {
      includeStack: true, // Include stack traces in assertion errors
      showDiff: true, // Show diffs for failed assertions
      truncateThreshold: 120, // Longer threshold for complex objects
    },
  },

  // Define global test utilities at root level
  define: {
    __TEST__: true,
  },

  // Resolve configuration
  resolve: {
    alias: pathAliases,
  },

  // Vite-specific test configuration
  esbuild: {
    // Ensure source maps work properly in tests
    sourcemap: true,
  },
});


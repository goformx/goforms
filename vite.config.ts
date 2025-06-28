import { defineConfig } from "vite";
import type { Plugin, PluginOption } from "vite";
import { resolve } from "path";
import autoprefixer from "autoprefixer";
import postcssImport from "postcss-import";
import postcssNested from "postcss-nested";
import cssnano from "cssnano";
import ejsPlugin from "./src/vite-plugin-ejs";
import { watch } from "fs";
import { createHash } from "crypto";

// Shared path mapping configuration
const pathAliases = {
  "@": resolve(__dirname, "src/js"),
  "@/core": resolve(__dirname, "src/js/core"),
  "@/features": resolve(__dirname, "src/js/features"),
  "@/shared": resolve(__dirname, "src/js/shared"),
  "@/pages": resolve(__dirname, "src/js/pages"),
  "@goformx/formio": resolve(__dirname, "node_modules/@goformx/formio"),
};

// Enhanced custom plugin to watch templ-generated files with better performance
function templWatcherPlugin(): Plugin {
  let server: any;
  let fileCache = new Map<string, string>();

  return {
    name: "templ-watcher",
    enforce: "pre", // Run early in the plugin chain
    configureServer(srv) {
      server = srv;
      const internalDir = resolve(__dirname, "internal");

      // Optimize file watching - only watch specific patterns
      watch(
        internalDir,
        {
          recursive: true,
        },
        async (_eventType: string, filename: string | null) => {
          if (!filename || !filename.endsWith("_templ.go")) return;

          try {
            // Debounce rapid file changes
            const filePath = resolve(internalDir, filename);
            const fileContent = await import("fs/promises").then((fs) =>
              fs.readFile(filePath, "utf-8").catch(() => ""),
            );

            // Create a simple hash to detect actual content changes
            const contentHash = createHash("md5")
              .update(fileContent)
              .digest("hex");
            const cachedHash = fileCache.get(filePath);

            if (cachedHash !== contentHash) {
              fileCache.set(filePath, contentHash);
              console.log(`[Vite] Templ file updated: ${filename}`);

              // Only trigger HMR if content actually changed
              server.ws.send({
                type: "full-reload",
                path: "*",
              });
            }
          } catch (error) {
            console.warn(
              `[Vite] Failed to process templ file ${filename}:`,
              error,
            );
          }
        },
      );
    },

    // Clean up cache on server restart
    buildStart() {
      fileCache.clear();
    },
  };
}

// Performance monitoring plugin
function performancePlugin(): Plugin {
  return {
    name: "performance-monitor",
    transform(code, _id) {
      // Log slow transformations in development
      if (process.env["NODE_ENV"] === "development") {
        const start = performance.now();
        return {
          code,
          map: null,
          meta: {
            vite: {
              lang: "js",
              duration: performance.now() - start,
            },
          },
        };
      }
      // Return undefined for non-development mode
      return undefined;
    },
  };
}

export default defineConfig(({ mode, command }) => {
  const isDev = mode === "development";
  const isBuild = command === "build";

  console.log(
    `[Vite] Mode: ${mode}, Command: ${command}, Assets from: ${isDev ? "0.0.0.0:5173" : "/assets"}`,
  );

  return {
    root: ".",
    publicDir: "public",
    appType: "custom",
    base: "/",

    // Enhanced CSS configuration
    css: {
      devSourcemap: true,
      modules: {
        localsConvention: "camelCase",
        scopeBehaviour: "local",
        generateScopedName: isDev
          ? "[name]__[local]__[hash:base64:5]"
          : "[hash:base64:8]",
      },
      postcss: {
        plugins: [
          postcssImport({
            // Optimize CSS imports
            filter: (url: string) => !url.startsWith("http"),
          }),
          postcssNested(),
          autoprefixer({
            // Target modern browsers in development for faster processing
            overrideBrowserslist: isDev
              ? ["last 1 chrome version", "last 1 firefox version"]
              : undefined,
          } as any),
          ...(isDev
            ? []
            : [
                cssnano({
                  preset: [
                    "default",
                    {
                      // More aggressive optimization for production
                      normalizeWhitespace: true,
                      discardComments: { removeAll: true },
                      mergeRules: true,
                    },
                  ],
                }),
              ]),
        ],
      },
    },

    // Optimized server configuration
    server: {
      port: 5173,
      host: "0.0.0.0",
      strictPort: true,
      cors: {
        origin: ["http://localhost:8090", "http://127.0.0.1:8090"],
        credentials: true,
        methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        allowedHeaders: [
          "Content-Type",
          "Authorization",
          "X-Csrf-Token",
          "X-Requested-With",
        ],
      },
      hmr: {
        port: 5173,
        // Optimize HMR overlay for better UX
        overlay: true,
        clientPort: 5173,
      },
      fs: {
        strict: false,
        allow: [".."],
        // Deny access to sensitive files
        deny: [".env*", "*.key", "*.pem"],
      },
      watch: {
        // Use native file system events instead of polling for better performance
        usePolling: process.env["DOCKER"] === "true", // Only use polling in Docker
        interval: 2000, // Increase interval to reduce CPU usage
        ignored: [
          "!**/*_templ.go",
          "coverage/**",
          "**/node_modules/**",
          "**/.git/**",
          "**/dist/**",
        ],
        // Add depth limit to prevent deep directory scanning
        depth: 20,
      },
      // Optimize middleware for development
      middlewareMode: false,
      proxy: {
        // Add API proxying if needed
        "/api": {
          target: "http://localhost:8090",
          changeOrigin: true,
          secure: false,
        },
      },
    },

    // Enhanced build configuration
    build: {
      outDir: "dist",
      emptyOutDir: true,
      manifest: true,
      sourcemap: isDev ? "inline" : true, // Inline in dev for faster builds
      target: "es2020", // More modern target for better optimization
      minify: "terser",
      modulePreload: {
        polyfill: true,
        resolveDependencies: (_filename, deps) => {
          // Optimize module preloading by filtering unnecessary deps
          return deps.filter((dep) => !dep.includes("chunk-"));
        },
      },
      cssCodeSplit: true,
      reportCompressedSize: true, // Enable for production insights
      chunkSizeWarningLimit: 1000, // Increase threshold

      // Enhanced terser options
      terserOptions: {
        compress: {
          drop_console: !isDev,
          drop_debugger: !isDev,
          pure_funcs: isDev ? [] : ["console.log", "console.debug"],
          passes: 2, // Multiple passes for better compression
        },
        mangle: {
          safari10: true, // Better Safari compatibility
        },
        format: {
          comments: false, // Remove all comments
        },
      },

      // Optimized rollup options with manual chunking
      rollupOptions: {
        input: {
          main: resolve(__dirname, "src/js/pages/main.ts"),
          "main.css": resolve(__dirname, "src/css/main.css"),
          dashboard: resolve(__dirname, "src/js/pages/dashboard.ts"),
          "form-builder": resolve(__dirname, "src/js/pages/form-builder.ts"),
          "new-form": resolve(__dirname, "src/js/pages/new-form.ts"),
          "form-preview": resolve(__dirname, "src/js/pages/form-preview.ts"),
          "cta-form": resolve(__dirname, "src/js/pages/cta-form.ts"),
          demo: resolve(__dirname, "src/js/pages/demo.ts"),
          login: resolve(__dirname, "src/js/pages/login.ts"),
          signup: resolve(__dirname, "src/js/pages/signup.ts"),
        },

        // Manual chunk splitting for better caching
        output: {
          // Enhanced asset file naming
          assetFileNames: (assetInfo) => {
            if (!assetInfo.name) {
              return "assets/[name].[hash][extname]";
            }
            const info = assetInfo.name.split(".");
            const ext = info[info.length - 1];

            if (["woff", "woff2", "ttf", "eot"].includes(ext)) {
              return "fonts/[name].[hash][extname]";
            }
            if (ext === "css") {
              return "assets/css/[name].[hash].css";
            }
            if (["png", "jpg", "jpeg", "gif", "svg", "webp"].includes(ext)) {
              return "assets/images/[name].[hash][extname]";
            }
            return "assets/[name].[hash][extname]";
          },

          chunkFileNames: (chunkInfo) => {
            // Organize chunks by type
            if (chunkInfo.name.includes("vendor")) {
              return "assets/js/vendor/[name].[hash].js";
            }
            if (chunkInfo.name.includes("shared")) {
              return "assets/js/shared/[name].[hash].js";
            }
            return "assets/js/[name].[hash].js";
          },

          entryFileNames: "assets/js/[name].[hash].js",

          // Advanced manual chunking strategy
          manualChunks: {
            // Vendor libraries
            vendor: ["@formio/js"],
            goformx: ["@goformx/formio"],
            // Validation and utilities
            utils: ["zod", "dompurify"],
            // Form.io runtime dependencies
            formio: ["ace-builds", "@fortawesome/fontawesome-free"],
            // Core application code - only include if files exist
            core: ["./src/js/core/http-client", "./src/js/core/logger"],
          },
        },

        // External dependencies (if building a library)
        external: isBuild ? [] : [],

        // Optimize treeshaking
        treeshake: {
          moduleSideEffects: false,
          propertyReadSideEffects: false,
          tryCatchDeoptimization: false,
        },
      },
    },

    // Enhanced resolve configuration
    resolve: {
      alias: pathAliases,
      extensions: [
        ".mjs",
        ".js",
        ".ts",
        ".jsx",
        ".tsx",
        ".json",
        ".ejs",
        ".ejs.js",
      ],
      // Optimize dependency resolution
      dedupe: ["@formio/js", "@goformx/formio"],
    },

    // Enhanced dependency optimization
    optimizeDeps: {
      include: [
        "@formio/js",
        "@goformx/formio",
        "zod",
        "dompurify",
        // Form.io runtime dependencies
        "ace-builds",
        "@fortawesome/fontawesome-free",
      ],
      exclude: [
        // Exclude development-only packages
        "@vitejs/plugin-react",
        "vite",
      ],
      esbuildOptions: {
        target: "es2020",
        supported: {
          "top-level-await": true,
          "import-meta": true,
        },
        define: {
          global: "globalThis",
        },
      },
      // Force dependency re-bundling on lockfile changes
      force: false,
      // Hold deps until static imports are crawled
      holdUntilCrawlEnd: true,
    },

    // Enhanced plugin configuration with optimized order
    plugins: [
      // Development-only plugins
      ...(isDev ? [performancePlugin()] : []),

      // Core plugins
      ejsPlugin(),

      // Custom plugins
      templWatcherPlugin(),

      // Production-only plugins
      ...(isBuild
        ? [
            // Add any production-specific plugins here
          ]
        : []),
    ] as PluginOption[],

    // Asset handling
    assetsInclude: ["**/*.ejs", "**/*.ejs.js"],

    // Enhanced worker configuration
    worker: {
      format: "es",
      plugins: () => [],
    },

    // Environment variables
    define: {
      __DEV__: isDev,
      __PROD__: !isDev,
      __VERSION__: JSON.stringify(process.env["npm_package_version"]),
    },

    // Preview server configuration
    preview: {
      port: 4173,
      host: "0.0.0.0",
      strictPort: true,
      cors: true,
    },

    // ESBuild configuration for better performance
    esbuild: {
      // Optimize for development speed
      target: isDev ? "es2022" : "es2020",
      // Remove console logs in production
      drop: isBuild ? ["console", "debugger"] : [],
      // Optimize JSX handling
      jsxDev: isDev,
      // Legal comments handling
      legalComments: isBuild ? "none" : "inline",
    },

    // JSON handling
    json: {
      namedExports: true,
      stringify: false,
    },

    // Experimental features
    experimental: {
      // Enable if you need build optimizations
      hmrPartialAccept: true,
    },

    // Cache directory for better performance
    cacheDir: "node_modules/.vite",

    // Clear screen on restart
    clearScreen: true,

    // Log level
    logLevel: isDev ? "info" : "warn",
  };
});

// Export path aliases for use in other config files
export { pathAliases };

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { resolve } from "path";
import tailwindcss from "@tailwindcss/postcss";
import autoprefixer from "autoprefixer";

// Shared path mapping configuration
const pathAliases = {
  "@": resolve(__dirname, "src"),
  "@/components": resolve(__dirname, "src/components"),
  "@/pages": resolve(__dirname, "src/pages"),
  "@/composables": resolve(__dirname, "src/composables"),
  "@/lib": resolve(__dirname, "src/lib"),
  "@/types": resolve(__dirname, "src/types"),
  // Preserve existing aliases for backward compatibility
  "@/core": resolve(__dirname, "src/lib/core"),
  "@/features": resolve(__dirname, "src/lib/features"),
  "@/shared": resolve(__dirname, "src/lib/shared"),
  "@goformx/formio": resolve(__dirname, "node_modules/@goformx/formio"),
};

export default defineConfig(({ mode, command }) => {
  const isDev = mode === "development";
  const isBuild = command === "build";

  console.log(
    `[Vite] Mode: ${mode}, Command: ${command}, Assets from: ${isDev ? "0.0.0.0:5173" : "/assets"}`,
  );

  return {
    root: ".",
    publicDir: "public",
    base: "/",

    plugins: [
      vue(),
    ],

    // CSS configuration with Tailwind
    css: {
      devSourcemap: true,
      postcss: {
        plugins: [
          tailwindcss(),
          autoprefixer(),
        ],
      },
    },

    // Server configuration
    server: {
      port: 5173,
      host: isDev ? "0.0.0.0" : "localhost",
      strictPort: true,
      cors: {
        origin: [
          "http://localhost:8090",
          "http://127.0.0.1:8090",
          "http://localhost:5173",
          "http://127.0.0.1:5173",
        ],
        credentials: true,
        methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        allowedHeaders: [
          "Content-Type",
          "Authorization",
          "X-Csrf-Token",
          "X-Requested-With",
          "X-Inertia",
          "X-Inertia-Version",
        ],
      },
      hmr: {
        port: 5173,
        overlay: true,
        clientPort: 5173,
      },
      fs: {
        strict: isDev ? false : true,
        allow: isDev ? [".."] : [],
        deny: [".env*", "*.key", "*.pem", "*.crt", "*.p12", "*.pfx"],
      },
      watch: {
        usePolling: process.env["DOCKER"] === "true",
        interval: 2000,
        ignored: [
          "**/node_modules/**",
          "**/.git/**",
          "**/dist/**",
        ],
        depth: 20,
      },
      proxy: {
        "/api": {
          target: "http://localhost:8090",
          changeOrigin: true,
          secure: false,
        },
      },
    },

    // Build configuration
    build: {
      outDir: "dist",
      emptyOutDir: true,
      manifest: true,
      sourcemap: isDev ? "inline" : true,
      target: "es2020",
      minify: "terser",
      cssCodeSplit: true,
      reportCompressedSize: true,
      chunkSizeWarningLimit: 1000,

      terserOptions: {
        compress: {
          drop_console: !isDev,
          drop_debugger: !isDev,
          pure_funcs: isDev ? [] : ["console.log", "console.debug"],
          passes: 2,
        },
        mangle: {
          safari10: true,
        },
        format: {
          comments: false,
        },
      },

      rollupOptions: {
        input: {
          main: resolve(__dirname, "index.html"),
        },
        output: {
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
            if (chunkInfo.name.includes("vendor")) {
              return "assets/js/vendor/[name].[hash].js";
            }
            if (chunkInfo.name.includes("shared")) {
              return "assets/js/shared/[name].[hash].js";
            }
            return "assets/js/[name].[hash].js";
          },
          entryFileNames: "assets/js/[name].[hash].js",
          manualChunks: {
            "vendor-vue": ["vue", "@inertiajs/vue3"],
            "vendor-formio": ["@formio/js"],
            "vendor-goformx": ["@goformx/formio"],
            "vendor-utils": ["zod", "clsx", "tailwind-merge", "class-variance-authority"],
          },
        },
        treeshake: {
          moduleSideEffects: false,
          propertyReadSideEffects: false,
          tryCatchDeoptimization: false,
        },
      },
    },

    // Resolve configuration
    resolve: {
      alias: pathAliases,
      extensions: [".mjs", ".js", ".ts", ".jsx", ".tsx", ".json", ".vue"],
      dedupe: ["vue", "@formio/js", "@goformx/formio"],
    },

    // Dependency optimization
    optimizeDeps: {
      include: [
        "vue",
        "@inertiajs/vue3",
        "@formio/js",
        "@goformx/formio",
        "zod",
        "clsx",
        "tailwind-merge",
        "class-variance-authority",
        "radix-vue",
        "lucide-vue-next",
      ],
      exclude: ["@vitejs/plugin-vue"],
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
      force: false,
      holdUntilCrawlEnd: true,
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
      host: isDev ? "0.0.0.0" : "localhost",
      strictPort: true,
    },

    // ESBuild configuration
    esbuild: {
      target: isDev ? "es2022" : "es2020",
      drop: isBuild ? ["console", "debugger"] : [],
      jsxDev: isDev,
      legalComments: isBuild ? "none" : "inline",
    },

    // JSON handling
    json: {
      namedExports: true,
      stringify: false,
    },

    // Experimental features
    experimental: {
      hmrPartialAccept: true,
    },

    cacheDir: "node_modules/.vite",
    clearScreen: true,
    logLevel: isDev ? "info" : "warn",
  };
});

export { pathAliases };

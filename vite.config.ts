import { defineConfig } from "vite";
import { resolve } from "path";
import autoprefixer from "autoprefixer";
import postcssImport from "postcss-import";
import postcssNested from "postcss-nested";
import cssnano from "cssnano";
import ejsPlugin from "./src/vite-plugin-ejs";

export default defineConfig({
  root: ".",
  publicDir: "public",
  appType: "custom",
  base: process.env.NODE_ENV === "production" ? "/assets/" : "/",
  css: {
    devSourcemap: true,
    modules: {
      localsConvention: "camelCase",
    },
    postcss: {
      plugins: [
        autoprefixer(),
        postcssImport(),
        postcssNested(),
        cssnano({
          preset: "default",
        }),
      ],
    },
  },
  server: {
    port: 3000,
    strictPort: true,
    cors: {
      origin: "http://localhost:8090", // Your Go server URL
      methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
      credentials: true,
    },
    proxy: {
      "/api": {
        target: "http://localhost:8090",
        changeOrigin: true,
        secure: false,
      },
    },
    hmr: {
      protocol: "ws",
      host: "localhost",
      port: 3000,
      clientPort: 3000,
    },
    host: true,
    middlewareMode: false,
    fs: {
      strict: false,
      allow: [".."],
    },
    watch: {
      usePolling: true,
      interval: 1000,
    },
    origin: "http://localhost:3000", // Explicitly set the Vite dev server origin
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
    manifest: true, // Generate manifest.json
    sourcemap: true,
    target: "esnext",
    minify: "terser",
    modulePreload: {
      polyfill: true,
    },
    cssCodeSplit: true,
    reportCompressedSize: false,
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
    rollupOptions: {
      input: {
        main: resolve(__dirname, "src/js/main.ts"),
        "main.css": resolve(__dirname, "src/css/main.css"),
        dashboard: resolve(__dirname, "src/js/dashboard.ts"),
        "form-builder": resolve(__dirname, "src/js/form-builder.ts"),
        login: resolve(__dirname, "src/js/login.ts"),
        signup: resolve(__dirname, "src/js/signup.ts"),
        "cta-form": resolve(__dirname, "src/js/cta-form.ts"),
        demo: resolve(__dirname, "src/js/demo.ts"),
      },
      output: {
        assetFileNames: (assetInfo) => {
          if (!assetInfo.name) {
            return "assets/[name][hash][extname]";
          }
          const info = assetInfo.name.split(".");
          const ext = info[info.length - 1];
          // Skip font files from being processed by Vite
          if (
            ext === "woff" ||
            ext === "woff2" ||
            ext === "ttf" ||
            ext === "eot"
          ) {
            // Place all font files in the fonts directory
            return "fonts/[name][extname]";
          }
          if (ext === "css") {
            return "assets/css/[name].[hash].css";
          }
          return "assets/[name].[hash][extname]";
        },
        chunkFileNames: "assets/js/[name].[hash].js",
        entryFileNames: "assets/js/[name].[hash].js",
      },
    },
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
      "@goformx/formio": resolve(__dirname, "node_modules/@goformx/formio"),
    },
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
  },
  optimizeDeps: {
    force: true,
    include: ["@formio/js", "@goformx/formio"],
    esbuildOptions: {
      target: "esnext",
      supported: {
        "top-level-await": true,
      },
    },
  },
  preview: {
    port: 8090,
    strictPort: true,
  },
  plugins: [ejsPlugin()],
  assetsInclude: ["**/*.ejs", "**/*.ejs.js"],
});

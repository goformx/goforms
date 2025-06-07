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
  base: "/",
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
  build: {
    outDir: "dist",
    emptyOutDir: true,
    manifest: true,
    sourcemap: true,
    target: "esnext",
    minify: "terser",
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true,
      },
    },
    rollupOptions: {
      input: {
        styles: resolve(__dirname, "src/css/main.css"),
        app: resolve(__dirname, "src/js/main.ts"),
        validation: resolve(__dirname, "src/js/validation.ts"),
        signup: resolve(__dirname, "src/js/signup.ts"),
        login: resolve(__dirname, "src/js/login.ts"),
        formBuilder: resolve(__dirname, "src/js/form-builder.ts"),
      },
      output: {
        entryFileNames: "assets/js/[name].[hash].js",
        chunkFileNames: "assets/js/[name].[hash].js",
        assetFileNames: (assetInfo) => {
          const name = assetInfo.name || "";
          if (name.endsWith(".css")) {
            return "assets/css/[name].[hash][extname]";
          }
          return "assets/[name].[hash][extname]";
        },
      },
    },
  },
  server: {
    port: 3000,
    strictPort: true,
    proxy: {
      "/api": {
        target: "http://localhost:8090",
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, "/api/v1"),
      },
    },
    hmr: {
      port: 3000,
    },
    host: true,
    middlewareMode: false,
    fs: {
      strict: false,
      allow: [".."],
    },
    watch: {
      usePolling: false,
      interval: 1000,
    },
  },
  resolve: {
    alias: {
      "@": resolve(__dirname, "src"),
      "@goformx/formio": resolve(__dirname, "node_modules/@goformx/formio"),
      "goforms-template": resolve(
        __dirname,
        "../goforms-template/lib/mjs/index.js",
      ),
      "goforms-template/templates": resolve(
        __dirname,
        "../goforms-template/lib/mjs/templates",
      ),
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
  // Configure how Vite handles different file types
  assetsInclude: ["**/*.ejs", "**/*.ejs.js"],
});

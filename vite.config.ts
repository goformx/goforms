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
  base: "/assets/",
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
    outDir: "dist/assets",
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
        main: "src/js/main.ts",
        dashboard: "src/js/dashboard.ts",
        "form-builder": "src/js/form-builder.ts",
        login: "src/js/login.ts",
        signup: "src/js/signup.ts",
        "cta-form": "src/js/cta-form.ts",
      },
      output: {
        assetFileNames: (assetInfo) => {
          if (assetInfo.name?.endsWith(".css")) {
            return "css/[name][hash][extname]";
          }
          return "assets/[name][hash][extname]";
        },
        chunkFileNames: "js/[name][hash].js",
        entryFileNames: "js/[name][hash].js",
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

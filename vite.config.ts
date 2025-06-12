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
    outDir: "dist",
    emptyOutDir: true,
    manifest: true,
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
        main: "src/js/main.ts",
        dashboard: "src/js/dashboard.ts",
        "form-builder": "src/js/form-builder.ts",
        login: "src/js/login.ts",
        signup: "src/js/signup.ts",
        "cta-form": "src/js/cta-form.ts",
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
            return "assets/css/[name].[hash][extname]";
          }
          return "assets/[name].[hash][extname]";
        },
        chunkFileNames: "assets/js/[name].[hash].js",
        entryFileNames: "assets/js/[name].[hash].js",
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
  assetsInclude: ["**/*.ejs", "**/*.ejs.js"],
});

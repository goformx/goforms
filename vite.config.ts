import { defineConfig } from "vite";
import { resolve } from "path";
import autoprefixer from "autoprefixer";
import postcssImport from "postcss-import";
import postcssNested from "postcss-nested";
import cssnano from "cssnano";
import ejsPlugin from "./src/vite-plugin-ejs";
import { watch } from "fs";

// Shared path mapping configuration
const pathAliases = {
  "@": resolve(__dirname, "src/js"),
  "@/core": resolve(__dirname, "src/js/core"),
  "@/features": resolve(__dirname, "src/js/features"),
  "@/shared": resolve(__dirname, "src/js/shared"),
  "@/pages": resolve(__dirname, "src/js/pages"),
  "@goformx/formio": resolve(__dirname, "node_modules/@goformx/formio"),
};

// Custom plugin to watch templ-generated files
function templWatcherPlugin() {
  return {
    name: "templ-watcher",
    configureServer(server: any) {
      // Watch for changes in *_templ.go files
      const internalDir = resolve(__dirname, "internal");
      
      watch(internalDir, { recursive: true }, (_eventType: string, filename: string | null) => {
        if (filename && filename.endsWith("_templ.go")) {
          console.log(`[Vite] Templ file changed: ${filename}`);
          // Trigger a full page reload when templ files change
          server.ws.send({
            type: "full-reload",
            path: "*",
          });
        }
      });
    },
  };
}

export default defineConfig(({ mode }) => {
  const isDev = mode === "development";
  
  console.log(`[Vite] Mode: ${mode}, serving assets from 0.0.0.0:3000`);

  return {
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
          ...(isDev ? [] : [cssnano({ preset: "default" })]),
        ],
      },
    },
    
    server: {
      port: 3000,
      host: "0.0.0.0", // Changed from "localhost" for Docker
      strictPort: true,
      cors: true,
      hmr: {
        protocol: "ws",
        host: "0.0.0.0", // Changed from "localhost" for Docker
        port: 3000,
        clientPort: 3000,
      },
      middlewareMode: false,
      fs: {
        strict: false,
        allow: [".."],
      },
      watch: {
        usePolling: true,
        interval: 1000,
        ignored: ["!**/*_templ.go", "coverage/**"],
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
        output: {
          assetFileNames: (assetInfo) => {
            if (!assetInfo.name) {
              return "assets/[name][hash][extname]";
            }
            const info = assetInfo.name.split(".");
            const ext = info[info.length - 1];
            
            if (["woff", "woff2", "ttf", "eot"].includes(ext)) {
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
      alias: pathAliases,
      extensions: [".mjs", ".js", ".ts", ".jsx", ".tsx", ".json", ".ejs", ".ejs.js"],
    },
    
    optimizeDeps: {
      include: ["@formio/js", "@goformx/formio"],
      esbuildOptions: {
        target: "esnext",
        supported: {
          "top-level-await": true,
        },
      },
    },
        
    plugins: [ejsPlugin(), templWatcherPlugin()],
    assetsInclude: ["**/*.ejs", "**/*.ejs.js"],
  };
});

// Export path aliases for use in other config files
export { pathAliases };

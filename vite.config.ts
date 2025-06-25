import { defineConfig, loadEnv } from "vite";
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
      const watchTemplFiles = () => {
        const internalDir = resolve(__dirname, "internal");
        
        // Watch the internal directory recursively
        const watchDir = (dir: string) => {
          watch(dir, { recursive: true }, (_eventType: string, filename: string | null) => {
            if (filename && filename.endsWith("_templ.go")) {
              console.log(`[Vite] Templ file changed: ${filename}`);
              // Trigger a full page reload when templ files change
              server.ws.send({
                type: "full-reload",
                path: "*",
              });
            }
          });
        };

        watchDir(internalDir);
      };

      // Start watching after server is ready
      server.middlewares.use((req: any, _res: any, next: any) => {
        if (req.url === "/") {
          // Initialize watcher on first request
          setTimeout(watchTemplFiles, 1000);
        }
        next();
      });
    },
  };
}

export default defineConfig(({ mode }) => {
  // Load environment variables
  const env = loadEnv(mode, process.cwd(), "");
  
  // Get Vite configuration from environment variables
  const viteHost = env.GOFORMS_VITE_DEV_HOST || "localhost";
  const vitePort = parseInt(env.GOFORMS_VITE_DEV_PORT || "3000", 10);
  const appHost = env.GOFORMS_APP_HOST || "localhost";
  const appPort = parseInt(env.GOFORMS_APP_PORT || "8090", 10);
  const appScheme = env.GOFORMS_APP_SCHEME || "http";
  
  // Build the app URL for CORS and proxy configuration
  // Use GOFORMS_APP_URL if provided, otherwise construct from individual parts
  const appUrl = env.GOFORMS_APP_URL || `${appScheme}://${appHost}:${appPort}`;
  
  console.log(`[Vite] Configuration:`, {
    viteHost,
    vitePort,
    appUrl,
    mode,
  });

  return {
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
      port: vitePort,
      strictPort: true,
      cors: {
        origin: appUrl, // Use environment-based app URL
        methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        credentials: true,
      },
      proxy: {
        "/api": {
          target: appUrl, // Use environment-based app URL
          changeOrigin: true,
          secure: false,
        },
      },
      hmr: {
        protocol: "ws",
        host: viteHost,
        port: vitePort,
        clientPort: vitePort,
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
        // Watch for templ-generated files
        ignored: ["!**/*_templ.go"],
      },
      origin: `http://${viteHost}:${vitePort}`, // Use environment-based Vite URL
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
          main: resolve(__dirname, "src/js/pages/main.ts"),
          "main.css": resolve(__dirname, "src/css/main.css"),
          dashboard: resolve(__dirname, "src/js/pages/dashboard.ts"),
          "form-builder": resolve(__dirname, "src/js/pages/form-builder.ts"),
          login: resolve(__dirname, "src/js/features/auth/login.ts"),
          signup: resolve(__dirname, "src/js/features/auth/signup.ts"),
          "cta-form": resolve(__dirname, "src/js/pages/cta-form.ts"),
          demo: resolve(__dirname, "src/js/pages/demo.ts"),
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
    plugins: [ejsPlugin(), templWatcherPlugin()],
    assetsInclude: ["**/*.ejs", "**/*.ejs.js"],
  };
});

// Export path aliases for use in other config files
export { pathAliases };

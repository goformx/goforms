import { defineConfig } from 'vite';
import { resolve } from 'path';
import autoprefixer from 'autoprefixer';
import postcssImport from 'postcss-import';
import postcssNested from 'postcss-nested';
import cssnano from 'cssnano';

export default defineConfig({
  root: '.',
  publicDir: 'public',
  appType: 'custom',
  base: '/',
  css: {
    devSourcemap: true,
    modules: {
      localsConvention: 'camelCase'
    },
    postcss: {
      plugins: [
        autoprefixer(),
        postcssImport(),
        postcssNested(),
        cssnano({
          preset: 'default'
        })
      ]
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    manifest: true,
    sourcemap: true,
    target: 'esnext',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    },
    rollupOptions: {
      input: {
        styles: resolve(__dirname, 'src/css/main.css'),
        app: resolve(__dirname, 'src/js/main.ts'),
        validation: resolve(__dirname, 'src/js/validation.ts'),
        signup: resolve(__dirname, 'src/js/signup.ts'),
        login: resolve(__dirname, 'src/js/login.ts'),
        formBuilder: resolve(__dirname, 'src/js/form-builder.ts')
      },
      output: {
        entryFileNames: 'js/[name].[hash].js',
        chunkFileNames: 'js/[name].[hash].js',
        assetFileNames: (assetInfo) => {
          const name = assetInfo.name || '';
          if (name.endsWith('.css')) {
            return 'css/[name].[hash][extname]';
          }
          return 'assets/[name].[hash][extname]';
        }
      }
    }
  },
  server: {
    port: 3000,
    strictPort: true,
    host: true,
    middlewareMode: false,
    fs: {
      strict: true,
      allow: ['src', 'node_modules']
    },
    watch: {
      usePolling: false,
      interval: 1000
    },
    hmr: {
      protocol: 'ws',
      host: 'localhost',
      port: 3000,
      clientPort: 3000,
      timeout: 5000,
      overlay: true
    },
    proxy: {
      // Proxy all API requests to the Go server
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true,
        secure: false,
        ws: true
      }
    }
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    },
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json']
  },
  optimizeDeps: {
    force: true,
    esbuildOptions: {
      target: 'esnext',
      supported: {
        'top-level-await': true
      }
    }
  },
  preview: {
    port: 3000,
    strictPort: true,
    host: true
  }
}); 
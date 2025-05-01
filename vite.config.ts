import { defineConfig } from 'vite';
import { resolve } from 'path';
import autoprefixer from 'autoprefixer';
import postcssImport from 'postcss-import';
import postcssNested from 'postcss-nested';
import cssnano from 'cssnano';

export default defineConfig({
  root: 'static',
  publicDir: 'public',
  css: {
    devSourcemap: true,
    modules: {
      localsConvention: 'camelCase'
    },
    postcss: {
      plugins: [
        autoprefixer,
        postcssImport,
        postcssNested,
        cssnano({
          preset: 'default'
        })
      ]
    }
  },
  build: {
    outDir: '../static/dist',
    emptyOutDir: true,
    manifest: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'static/js/main.ts'),
        validation: resolve(__dirname, 'static/js/validation.ts'),
        signup: resolve(__dirname, 'static/js/signup.ts'),
        login: resolve(__dirname, 'static/js/login.ts'),
        styles: resolve(__dirname, 'static/css/main.css')
      },
      output: {
        entryFileNames: 'js/[name].[hash].js',
        chunkFileNames: 'js/[name].[hash].js',
        assetFileNames: (assetInfo) => {
          const name = assetInfo.names?.[0] || '';
          if (name.endsWith('.css')) {
            return 'css/[name].[hash][extname]';
          }
          return 'assets/[name].[hash][extname]';
        }
      }
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true
      }
    },
    hmr: {
      protocol: 'ws',
      host: 'localhost',
      port: 3000
    },
    watch: {
      usePolling: true,
      ignored: ['**/node_modules/**', '**/dist/**']
    }
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'static')
    }
  },
  esbuild: {
    format: 'esm'
  },
  optimizeDeps: {
    esbuildOptions: {
      format: 'esm'
    }
  }
}); 
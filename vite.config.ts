import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  root: 'static',
  publicDir: 'public',
  build: {
    outDir: '../static/dist',
    emptyOutDir: true,
    manifest: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'static/js/main.ts'),
        validation: resolve(__dirname, 'static/js/validation.ts'),
        signup: resolve(__dirname, 'static/js/signup.ts'),
        login: resolve(__dirname, 'static/js/login.ts')
      },
      output: {
        entryFileNames: 'js/[name].[hash].js',
        chunkFileNames: 'js/[name].[hash].js',
        assetFileNames: 'assets/[name].[hash].[ext]'
      }
    }
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8090',
        changeOrigin: true
      }
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
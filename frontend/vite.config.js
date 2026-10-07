import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'
import { fileURLToPath, URL } from 'node:url'

const backendTarget = process.env.VITE_BACKEND_PROXY_URL || 'http://localhost:8080'

export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  resolve: {
    alias: {
      $lib: fileURLToPath(new URL('./src/lib', import.meta.url)),
    },
  },
  server: {
    watch: {
      usePolling: true,
      interval: 500,
    },
    proxy: {
      '/api': {
        target: backendTarget,
        changeOrigin: true,
      },
      '/healthz': {
        target: backendTarget,
        changeOrigin: true,
      },
      '/uploads': {
        target: backendTarget,
        changeOrigin: true,
      },
    },
  },
})


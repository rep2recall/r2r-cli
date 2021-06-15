import path from 'path'

import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import mpa from 'vite-plugin-mpa'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), mpa()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      external: /^\/vendor\//,
    },
  },
})

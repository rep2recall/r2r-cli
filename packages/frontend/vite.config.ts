import { spawnSync } from 'child_process'
import path from 'path'

import mpa from '@patarapolw/vite-plugin-mpa'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    mpa(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      external: /^\/vendor\//,
    },
    outDir: `../../dist/${spawnSync('go', ['env', 'GOOS']).stdout.toString().trimEnd()}/public`,
    emptyOutDir: true
  },
  server: {
    proxy: {
      '/api': 'http://localhost:25459',
      '/proxy': 'http://localhost:25459',
      '/server': 'http://localhost:25459',
    },
  },
})

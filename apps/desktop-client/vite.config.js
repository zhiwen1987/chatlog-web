import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Desktop client 前端构建（Tauri 通过 devUrl / frontendDist 接入）
export default defineConfig({
  plugins: [vue()],
  clearScreen: false,
  server: {
    port: 1420,
    strictPort: true
  }
})
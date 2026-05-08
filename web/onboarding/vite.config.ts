import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api/users': {
        target: 'http://localhost:3003',
        rewrite: path => path.replace(/^\/api/, ''),
      },
      '/api': {
        target: 'http://localhost:3000',
        rewrite: path => path.replace(/^\/api/, ''),
      },
      '/captcha': {
        target: 'http://localhost:3003',
      },
    },
  },
})

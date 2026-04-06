import { resolve } from 'path'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import eslint from 'vite-plugin-eslint2'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// https://vitejs.dev/config/
export default defineConfig({
  base: './',
  plugins: [
    vue(),
    eslint({
      cache: false,
      fix: false
    }),
    Components({
      resolvers: [ElementPlusResolver()]
    })
  ],
  resolve: {
    alias: {
      '@': resolve('./src')
    }
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "${resolve('./src/assets/css/_variables.scss').replace(/\\/g, '/')}" as *;`
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vue: ['vue'],
          'element-plus': ['element-plus', '@element-plus/icons-vue']
        }
      }
    },
    chunkSizeWarningLimit: 500
  }
})

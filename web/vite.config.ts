import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import UnoCSS from 'unocss/vite'
import viteCompression from 'vite-plugin-compression';
import { VitePWA } from 'vite-plugin-pwa'

import { welcomePlugin } from './src/plugins/welcome-plugin'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
    UnoCSS(),
    viteCompression({
      deleteOriginFile: false,
    }),
    VitePWA({
      registerType: 'autoUpdate',
      injectRegister: 'auto',
      workbox: {
        // 💡 恢复 .html 预缓存，作为离线时的最后一道防线
        globPatterns: ['**/*.{js,css,html,ico,png,svg,webp,woff2}'],
        globIgnores: ['**/app.webmanifest', '**/rss*', '**/sitemap*', '**/robots.txt'],
        // 增加缓存容量限制，防止大资源无法缓存 (5MB)
        maximumFileSizeToCacheInBytes: 5 * 1024 * 1024,

        // 💡 优化后的 PWA 多媒体与 S3 专项策略
        runtimeCaching: [
          // 1. 首页 HTML 导航
          {
            urlPattern: ({ request }) => request.mode === 'navigate',
            handler: 'NetworkFirst',
            options: {
              cacheName: 'html-cache',
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          // 2. 视觉媒体缓存 (Images/Avatars): 涵盖本地与外部 S3/Fediverse 内容
          {
            urlPattern: ({ url }) =>
              /\.(?:png|jpg|jpeg|svg|gif|webp|avif)$/i.test(url.pathname) ||
              url.pathname.startsWith('/api/icon') ||
              url.pathname.includes('/images/'),
            handler: 'CacheFirst',
            options: {
              cacheName: 'api-media-cache',
              expiration: {
                maxEntries: 300, // 增加容量以涵盖更多图片
                maxAgeSeconds: 60 * 60 * 24 * 30, // 缓存 30 天
              },
              cacheableResponse: { statuses: [0, 200] }, // 支持 Opaque 响应
            },
          },
          // 4. 核心启动与元数据 (NetworkFirst): 保证 initStores 加载的数据【即时更新】
          {
            urlPattern: ({ url }) =>
              url.pathname.includes('/settings') || // 系统、S3、评论、存储设置
              url.pathname.includes('/info') ||     // Agent 信息
              url.pathname.includes('/user') ||     // 登录态
              url.pathname.includes('/tags') ||     // 标签列表 (UI 筛选核心)
              url.pathname.includes('/status') ||   // 准备状态
              url.pathname.includes('/hello'),      // 基础属性
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-core-config-cache',
              expiration: { maxEntries: 60, maxAgeSeconds: 60 * 60 * 24 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          // 5. 辅助功能与身份清单 (StaleWhileRevalidate): 保证 App 身份识别秒开
          {
            urlPattern: ({ url }) =>
              url.pathname.includes('/connect') ||
              url.pathname.includes('/passkeys') ||
              url.pathname.includes('/getmusic') ||
              url.pathname.endsWith('.webmanifest'),
            handler: 'StaleWhileRevalidate',
            options: {
              cacheName: 'api-aux-config-cache',
              expiration: { maxEntries: 100, maxAgeSeconds: 60 * 60 * 24 * 7 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          // 6. 动态流内容读取 (NetworkFirst): 首页动态、收件箱、待办、AI近况、热力图
          {
            urlPattern: ({ url }) =>
              url.pathname.includes('/echo/page') ||
              url.pathname.includes('/timeline') ||
              url.pathname.includes('/inbox') ||
              url.pathname.includes('/heatmap') ||
              url.pathname.includes('/todo') ||
              url.pathname.includes('/agent/recent'),
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-content-cache',
              expiration: { maxEntries: 100, maxAgeSeconds: 60 * 60 * 24 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
        ],

        // 排除掉所有后端专属的纯数据/功能路由，防止 Vue 路由误拦截
        navigateFallbackDenylist: [
          /^\/rss/,
          /^\/api/,
          /^\/oauth/,
          /^\/ws/,
          /^\/swagger/,
          /^\/healthz/,
          /^\/robots\.txt/,
          /^\/sitemap.*/,
          /^\/\.well-known\//,
          /^\/users\//,
          /^\/objects\//,
          /^\/images\//, // 💡 原始资源文件直通后端
          /^\/videos\//,
          /^\/audios\//,
          /^\/avatar\//,
          /^\/webhook\//,
          /app\.webmanifest$/,
        ],
      },
      // 由于 Ech0 使用后端动态注入 PWA 属性 (Manifest/Icons)，我们在此仅处理 Service Worker 逻辑
      manifest: false,
    }),

    welcomePlugin() // 欢迎横幅插件
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    // 当使用embed时则调整构建输出到后端的template/dist目录
    outDir: '../template/dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        // 代码分割：将重型库打包到单独的 chunk 中，利用浏览器缓存
        manualChunks: {
          // Markdown 编辑器核心
          'md-editor': ['md-editor-v3'],
          // 代码高亮
          'highlight': ['highlight.js'],
          // 数学公式
          'katex': ['katex'],
          // 图表
          'mermaid': ['mermaid'],
          // 代码格式化
          'prettier': ['prettier'],
          // 图表库
          'echarts': ['echarts', 'vue-echarts'],
        },
      },
    },
  }
})

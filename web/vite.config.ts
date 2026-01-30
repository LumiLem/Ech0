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
        // 💡 允许 index.html 回归预缓存清单
        // 但我们要强制禁止 Workbox 自动生成 index.html 的 NavigationRoute (它会锁死动态 HTML)
        // 💡 保证 index.html 在预缓存中，但通过 navigateFallback: null 禁用自动拦截
        globPatterns: ['**/*.{css,js,html,svg,png,ico,webp,woff2}'],
        navigateFallback: null,
        globIgnores: ['**/app.webmanifest', '**/rss*', '**/sitemap*', '**/robots.txt'],
        // 增加缓存容量限制，防止大资源无法缓存 (5MB)
        maximumFileSizeToCacheInBytes: 5 * 1024 * 1024,

        // 💡 优化后的 PWA 多媒体与 S3 专项策略
        runtimeCaching: [
          // 1. 首页与动态页面导航 (SEO + 离线兜底)
          {
            urlPattern: ({ request }) => request.mode === 'navigate',
            handler: 'NetworkFirst',
            options: {
              cacheName: 'html-cache',
              cacheableResponse: { statuses: [0, 200] },
              // 💡 核心秘诀：离线兜底插件
              plugins: [
                {
                  // 当网络请求失败且本地也没有该特定 URL 的缓存时触发
                  handlerDidError: async () => {
                    const ctx = (globalThis as any);
                    if (!ctx.caches) return;
                    // 💡 关键：使用 ignoreSearch 忽略预缓存的版本号，并尝试多个可能的 Key
                    const options = { ignoreSearch: true };
                    return (await ctx.caches.match('index.html', options)) ||
                      (await ctx.caches.match('/index.html', options)) ||
                      (await ctx.caches.match('/', options));
                  }
                }
              ]
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
              expiration: { maxEntries: 60, maxAgeSeconds: 86400 },
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
              expiration: { maxEntries: 100, maxAgeSeconds: 604800 },
              cacheableResponse: { statuses: [0, 200] },
            },
          },
          // 6. 首页列表拆解器 (Scheme A): 当用户刷首页时，自动把每一条动态的详情 API 预填进缓存
          {
            urlPattern: ({ url }) => url.pathname.includes('/echo/page'),
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-content-cache',
              expiration: { maxEntries: 200, maxAgeSeconds: 86400 },
              cacheableResponse: { statuses: [0, 200] },
              plugins: [
                {
                  fetchDidSucceed: async ({ response }) => {
                    const ctx = (globalThis as any);
                    if (!ctx.caches || !ctx.Response) return response;
                    const clone = response.clone();
                    // 在后台执行拆解注入，不阻塞当前列表返回
                    (async () => {
                      try {
                        const json = (await clone.json()) as any;
                        if (json && json.code === 1 && json.data?.items) {
                          const cache = await ctx.caches.open('api-content-cache');
                          for (const item of json.data.items) {
                            // 构造详情页 API 的 URL Key
                            const detailUrl = new URL(`/api/echo/${item.id}`, ctx.location.origin).href;
                            // 构造符合后端定义的标准响应结构
                            const detailRes = new ctx.Response(JSON.stringify({
                              code: 1,
                              msg: "OK",
                              data: item
                            }), {
                              headers: { 'Content-Type': 'application/json' }
                            });
                            await cache.put(detailUrl, detailRes);
                          }
                        }
                      } catch (e) { }
                    })();
                    return response;
                  }
                }
              ]
            },
          },
          // 7. 动态详情、收件箱等其他读请求
          {
            urlPattern: ({ url }) =>
              /^\/api\/echo\//.test(url.pathname) ||
              url.pathname.includes('/timeline') ||
              url.pathname.includes('/inbox') ||
              url.pathname.includes('/heatmap') ||
              url.pathname.includes('/todo') ||
              url.pathname.includes('/agent/recent'),
            handler: 'NetworkFirst',
            options: {
              cacheName: 'api-content-cache',
              expiration: { maxEntries: 200, maxAgeSeconds: 86400 },
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

<script setup lang="ts">
import { RouterView, useRouter, useRoute } from 'vue-router'
import { computed, onMounted, ref } from 'vue'
import { watch } from 'vue'
import { useHead } from '@unhead/vue'
import { useSettingStore, useEditorStore, useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { Toaster } from 'vue-sonner'
import { getApiUrl } from './service/request/shared'
import 'vue-sonner/style.css'
import BaseDialog from './components/common/BaseDialog.vue'

import { useBaseDialog } from '@/composables/useBaseDialog'

const { register, title, description, handleConfirm } = useBaseDialog()
const dialogRef = ref()

// 路由切换动画
const router = useRouter()
const route = useRoute()
const transitionName = ref('fade')

// 监听路由变化，根据导航方向选择动画
router.afterEach((to, from) => {
  // Panel 子页面之间切换不使用动画
  const toName = to.name as string
  const fromName = from.name as string
  if (toName?.startsWith('panel-') && fromName?.startsWith('panel-')) {
    transitionName.value = 'none'
    return
  }

  // 定义路由层级（用于判断前进/后退）
  const routeDepth: Record<string, number> = {
    home: 0,
    echo: 1,
    panel: 1,
    auth: 1,
    connect: 1,
    hub: 1,
    widget: 1,
    'not-found': 2,
  }

  const toDepth = routeDepth[toName] ?? 1
  const fromDepth = routeDepth[fromName] ?? 1

  if (toDepth > fromDepth) {
    transitionName.value = 'slide-left'
  } else if (toDepth < fromDepth) {
    transitionName.value = 'slide-right'
  } else {
    transitionName.value = 'fade'
  }
})

const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

const API_URL = getApiUrl()
import { ensureAbsoluteUrl } from '@/utils/other'

const editorStore = useEditorStore()
const userStore = useUserStore()
const { isLogin } = storeToRefs(userStore)

// 使用 unhead 管理全局 Meta
useHead({
  title: computed(() => SystemSetting.value.site_title || 'Ech0'),
  meta: [
    {
      name: 'description',
      content: computed(() => SystemSetting.value.site_description)
    },
    {
      name: 'keywords',
      content: computed(() => SystemSetting.value.site_keywords)
    },
    {
      name: 'author',
      content: computed(() => SystemSetting.value.server_name || 'Ech0 Team')
    },
    // OpenGraph 基础配置
    {
      property: 'og:title',
      content: computed(() => SystemSetting.value.site_title)
    },
    {
      property: 'og:description',
      content: computed(() => SystemSetting.value.site_description)
    },
    {
      property: 'og:image',
      content: computed(() => {
        const logo = SystemSetting.value.server_logo
        return ensureAbsoluteUrl(logo || '/Ech0.png', SystemSetting.value.server_url)
      })
    },
    {
      property: 'og:url',
      content: computed(() => window.location.href)
    },
    {
      property: 'og:site_name',
      content: computed(() => SystemSetting.value.server_name || SystemSetting.value.site_title)
    }
  ],
  link: [
    {
      id: 'favicon',
      rel: 'icon',
        href: computed(() => {
          const logo = SystemSetting.value.server_logo
          if (!logo?.trim()) return '/favicon.ico'
          if (logo.startsWith('http')) return logo
          // 💡 提取文件名作为版本号，强制触发浏览器 Favicon 更新
          const v = logo.split('/').pop() || ''
          return `/api/icon?s=32&fmt=ico&v=${v}`
        })
    },
    {
      rel: 'apple-touch-icon',
        href: computed(() => {
          const logo = SystemSetting.value.server_logo
          if (!logo?.trim()) return '/apple-touch-icon.png'
          if (logo.startsWith('http')) return logo
          // 同步应用版本号
          const v = logo.split('/').pop() || ''
          return `/api/icon?s=180&v=${v}`
        })
    }
  ]
})

const injectCustomContent = () => {
  // 注入自定义 CSS
  if (SystemSetting.value.custom_css && SystemSetting.value.custom_css.length > 0) {
    const styleTag = document.createElement('style')
    styleTag.textContent = SystemSetting.value.custom_css
    document.head.appendChild(styleTag)
  }

  // 注入自定义 JS
  if (SystemSetting.value.custom_js && SystemSetting.value.custom_js.length > 0) {
    const scriptTag = document.createElement('script')
    scriptTag.textContent = SystemSetting.value.custom_js
    document.body.appendChild(scriptTag)
  }
}

onMounted(() => {
  // 注入自定义CSS 和 JS
  watch(
    () => SystemSetting.value.custom_css || SystemSetting.value.custom_js,
    (newSetting) => {
      if (newSetting) {
        injectCustomContent()
      }
    },
    { immediate: true },
  )

  // 初始注入
  register(dialogRef.value) // 全局注册弹窗对话框

  // 处理 PWA 分享目标 (Web Share Target)
  // 如果 URL 包含 share=true，说明是从系统分享跳转过来的
  router.isReady().then(() => {
    if (route.query.share === 'true' && isLogin.value) {
      const params = {
        title: route.query.share_title as string,
        text: route.query.share_text as string,
        url: route.query.share_url as string,
      }

      if (params.title || params.text || params.url) {
        editorStore.handleIncomingShare(params)

        // 清理 URL 参数，避免刷新时重复填入
        const newQuery = { ...route.query }
        delete newQuery.share
        delete newQuery.share_title
        delete newQuery.share_text
        delete newQuery.share_url

        router.replace({ query: newQuery })
      }
    }
  })
})
</script>

<template>
  <!-- 路由视图 - 带切换动画 -->
  <RouterView v-slot="{ Component, route }">
    <Transition :name="transitionName" mode="out-in">
      <component :is="Component" :key="route.path" />
    </Transition>
  </RouterView>
  <!-- 通知组件 -->
  <Toaster theme="light" position="top-right" :expand="false" richColors />
  <!-- 全局弹窗对话框 -->
  <BaseDialog ref="dialogRef" :title="title" :description="description" @confirm="handleConfirm" />
</template>

<style scoped>
/* 路由切换动画 - 淡入淡出 + 轻微滑动 */
.fade-enter-active,
.fade-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* 滑动动画 - 用于前进后退 */
.slide-left-enter-active,
.slide-left-leave-active,
.slide-right-enter-active,
.slide-right-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.slide-left-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.slide-left-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

.slide-right-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.slide-right-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>

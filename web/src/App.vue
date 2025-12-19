<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted, ref } from 'vue'
import { watch } from 'vue'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { Toaster } from 'vue-sonner'
import { getApiUrl } from './service/request/shared'
import 'vue-sonner/style.css'
import BaseDialog from './components/common/BaseDialog.vue'

import { useBaseDialog } from '@/composables/useBaseDialog'

const { register, title, description, handleConfirm } = useBaseDialog()
const dialogRef = ref()

const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

const DEFAULT_FAVICON = '/favicon.ico'
const API_URL = getApiUrl()

const updateFavicon = (logo?: string) => {
  const head = document.head
  if (!head) return

  const href = logo?.trim() ? API_URL + logo : DEFAULT_FAVICON
  const iconLinks = head.querySelectorAll<HTMLLinkElement>('link[rel*="icon"]')

  if (iconLinks.length > 0) {
    iconLinks.forEach((link) => {
      link.href = href
    })
    return
  }

  const newFavicon = document.createElement('link')
  newFavicon.rel = 'icon'
  newFavicon.href = href
  head.appendChild(newFavicon)
}

watch(
  () => SystemSetting.value.site_title,
  (title) => {
    if (title) document.title = title
  },
  { immediate: true },
)

watch(
  () => SystemSetting.value.server_logo,
  (logo) => {
    updateFavicon(logo)
  },
  { immediate: true },
)

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
})
</script>

<template>
  <!-- 路由视图 -->
  <RouterView />
  <!-- 通知组件 -->
  <Toaster theme="light" position="top-right" :expand="false" richColors />
  <!-- 全局弹窗对话框 -->
  <BaseDialog ref="dialogRef" :title="title" :description="description" @confirm="handleConfirm" />
</template>

<style scoped></style>

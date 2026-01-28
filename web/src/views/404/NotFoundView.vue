<script setup lang="ts">
import NotFoundPage from './modules/NotFoundPage.vue'
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import { useSettingStore } from '@/stores'

const settingStore = useSettingStore()

useHead({
  title: computed(() => `页面未找到 - ${settingStore.SystemSetting.site_title}`)
})
</script>

<template>
  <main>
    <NotFoundPage />
  </main>
</template>

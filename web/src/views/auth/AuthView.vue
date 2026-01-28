<script setup lang="ts">
import AuthPage from './modules/AuthPage.vue'
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import { useSettingStore } from '@/stores'

const settingStore = useSettingStore()

useHead({
  title: computed(() => `登录 - ${settingStore.SystemSetting.site_title}`)
})
</script>

<template>
  <div class="w-full">
    <AuthPage />
  </div>
</template>

<style></style>

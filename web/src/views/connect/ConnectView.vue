<script setup lang="ts">
import ConnectPage from './modules/ConnectPage.vue'
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import { useSettingStore } from '@/stores'

const settingStore = useSettingStore()

useHead({
  title: computed(() => `我的连接 - ${settingStore.SystemSetting.site_title}`)
})
</script>

<template>
  <div class="w-full">
    <ConnectPage />
  </div>
</template>

<style></style>

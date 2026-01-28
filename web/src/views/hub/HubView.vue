<script setup lang="ts">
import HubPage from './modules/HubPage.vue'
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import { useSettingStore } from '@/stores'

const settingStore = useSettingStore()

useHead({
  title: computed(() => `Hub - ${settingStore.SystemSetting.site_title}`)
})
</script>

<template>
  <div class="w-full px-2">
    <HubPage />
  </div>
</template>

<style></style>

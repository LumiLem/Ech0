<script setup lang="ts">
import WidgetPage from './modules/WidgetPage.vue'
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import { useSettingStore } from '@/stores'

const settingStore = useSettingStore()

useHead({
  title: computed(() => `Widgets - ${settingStore.SystemSetting.site_title}`)
})
</script>

<template>
  <div class="w-full px-2">
    <WidgetPage />
  </div>
</template>

<style></style>

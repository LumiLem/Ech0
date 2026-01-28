<script setup lang="ts">
import FediversePage from './modules/FediversePage.vue'
import { computed } from 'vue'
import { useHead } from '@unhead/vue'
import { useSettingStore } from '@/stores'

const settingStore = useSettingStore()

useHead({
  title: computed(() => `联邦宇宙 - ${settingStore.SystemSetting.site_title}`)
})
</script>

<template>
  <div class="w-full">
    <FediversePage />
  </div>
</template>

<style></style>

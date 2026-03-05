<template>
  <div class="w-full px-2">
    <!-- 存储设置 -->
    <TheStorageSetting class="mb-3" />

    <!-- 图片处理设置 -->
    <TheImageProcessSetting class="mb-3" />
  </div>
</template>

<script setup lang="ts">
import TheStorageSetting from './TheSetting/TheStorageSetting.vue'
import TheImageProcessSetting from './TheSetting/TheImageProcessSetting.vue'
</script>

<style scoped></style>

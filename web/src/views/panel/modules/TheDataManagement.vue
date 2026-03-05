<template>
  <div class="w-full px-2">
    <!-- 数据管理 -->
    <div class="flex flex-col md:flex-row gap-3 mb-3">
      <TheBackupSetting class="flex-1" />
      <TheOriginalCompatibility class="flex-1" />
    </div>

    <!-- 备份计划设置 -->
    <TheBackupScheduleSetting />
  </div>
</template>

<script setup lang="ts">
import TheBackupSetting from './TheSetting/TheBackupSetting.vue'
import TheOriginalCompatibility from './TheSetting/TheOriginalCompatibility.vue'
import TheBackupScheduleSetting from './TheSetting/TheBackupScheduleSetting.vue'
</script>

<style scoped></style>

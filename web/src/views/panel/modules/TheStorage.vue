<template>
  <div class="w-full px-2">
    <!-- 存储设置 -->
    <TheStorageSetting class="mb-3" />

    <!-- 数据管理 -->
    <div class="flex flex-col md:flex-row gap-3">
      <TheBackupSetting class="flex-1 mb-3" />
      <TheOriginalCompatibility class="flex-1 mb-3" />
    </div>

    <!-- 备份计划设置 -->
    <TheBackupScheduleSetting />
  </div>
</template>

<script setup lang="ts">
import TheStorageSetting from './TheSetting/TheStorageSetting.vue'
import TheBackupSetting from './TheSetting/TheBackupSetting.vue'
import TheOriginalCompatibility from './TheSetting/TheOriginalCompatibility.vue'
import TheBackupScheduleSetting from './TheSetting/TheBackupScheduleSetting.vue'
</script>

<style scoped></style>

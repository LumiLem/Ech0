<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--text-color-600)] font-bold text-lg">原版数据兼容</h1>
      </div>

      <div class="text-xs text-[var(--text-color-next-400)] mb-4 leading-relaxed">
        当前版本与原版的数据库结构有所不同。如果您需要切换回原版运行，请先手动同步数据。
      </div>

      <div class="flex flex-col gap-4 font-semibold text-sm">
        <!-- 同步到原版 -->
        <div class="flex flex-start items-center gap-2">
          <p class="text-[var(--text-color-next-500)] shrink-0 w-18">重建数据:</p>
          <BaseButton
            :icon="RestoreBackup"
            @click="handleSyncLegacy"
            class="rounded-lg text-[var(--text-color-600)]!"
            title="手动同步索引"
          />
          <span class="text-[10px] text-[var(--text-color-next-400)] font-normal leading-tight"
            >将数据同步回原版表 (不支持视频/实况照片)</span
          >
        </div>

        <!-- 清理原版数据 -->
        <div class="flex flex-start items-center gap-2">
          <p class="text-[var(--text-color-next-500)] shrink-0 w-18">清理数据:</p>
          <BaseButton
            :icon="Trashbin"
            @click="handleCleanLegacy"
            class="rounded-lg text-[var(--text-color-600)]!"
            title="清理原版表"
          />
          <span class="text-[10px] text-[var(--text-color-next-400)] font-normal leading-tight"
            >删除原版表以释放空间 (切记备份)</span
          >
        </div>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import RestoreBackup from '@/components/icons/restorebackup.vue'
import Trashbin from '@/components/icons/trashbin.vue'
import { fetchSyncLegacy, fetchCleanLegacy } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'

const { openConfirm } = useBaseDialog()

/**
 * 同步数据到原版表索引
 */
const handleSyncLegacy = async () => {
  try {
    await theToast.promise(fetchSyncLegacy(), {
      loading: '正在努力同步索引中...',
      success: (res: any) => res.msg || '同步成功！您可以随时切换回原版了',
      error: '同步失败，请检查后端日志',
    })
  } catch (err) {
    console.error('[DataSync] 同步异常', err)
  }
}

/**
 * 清理原版表
 */
const handleCleanLegacy = async () => {
  openConfirm({
    title: '确定要删除原版数据表吗？',
    description: '清理后，如果您想回退到原版，需要重新执行同步操作。',
    onConfirm: async () => {
      try {
        await theToast.promise(fetchCleanLegacy(), {
          loading: '正在清理中...',
          success: (res: any) => res.msg || '清理成功，数据库现在更干净了',
          error: '清理失败，请重试',
        })
      } catch (err) {
        console.error('[DataClean] 清理异常', err)
      }
    },
  })
}
</script>

<style scoped></style>

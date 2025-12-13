<template>
  <div class="mx-auto px-2 sm:px-4 md:px-6 my-4 sm:my-5 md:my-6">
    <!-- Echos -->
    <div v-if="echoStore.filteredEchoList">
      <div v-for="echo in echoStore.filteredEchoList" :key="echo.id" class="will-change-transform">
        <TheEchoCard
          :echo="echo"
          @refresh="handleRefresh"
          @update-like-count="handleUpdateLikeCount"
        />
      </div>
    </div>
    <!-- 加载更多 -->
    <div
      v-if="echoStore.filteredHasMore && !echoStore.isLoading"
      class="mb-4 mt-1 -ml-2 flex items-center justify-between"
    >
      <BaseButton
        @click="handleLoadMore"
        class="rounded-full bg-[var(--timeline-load-more-bg-color)] !active:bg-[var(--bg-color-100)] mr-2"
      >
        <span class="text-[var(--text-color-600)] text-md text-center px-2 py-1">继续装填</span>
      </BaseButton>
      <TheBackTop class="xl:hidden w-8 h-8 p-1" />
    </div>
    <!-- 没有更多 -->
    <div v-if="!echoStore.filteredHasMore && !echoStore.isLoading" class="mx-auto my-5 text-center">
      <span class="text-xl text-[var(--text-color-400)]">没有啦！🎉</span>
    </div>
    <!-- 加载中 -->
    <div v-if="echoStore.isLoading" class="mx-auto my-5 text-center">
      <span class="text-xl text-[var(--text-color-400)]">加载中...</span>
    </div>
    <!-- 备案号 -->
    <div class="text-center">
      <a href="https://beian.miit.gov.cn/" target="_blank">
        <span class="text-[var(--text-color-400)] text-sm">
          {{ SystemSetting.ICP_number }}
        </span>
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEchoCard from '@/components/advanced/TheEchoCard.vue'
import { onMounted, watch } from 'vue'
import { useEchoStore } from '@/stores/echo'
import { useSettingStore } from '@/stores/setting'
import BaseButton from '@/components/common/BaseButton.vue'
import { storeToRefs } from 'pinia'
import TheBackTop from '@/components/advanced/TheBackTop.vue'

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)
const { filteredTag, filteredDate, filteredYearMonth } = storeToRefs(echoStore)

const handleLoadMore = async () => {
  echoStore.filteredCurrent = echoStore.filteredCurrent + 1
  await echoStore.getEchosByPageForFilter()
}

// 刷新数据
const handleRefresh = () => {
  echoStore.refreshEchosForFilter()
}

// 刷新点赞数据
const handleUpdateLikeCount = (echoId: number) => {
  echoStore.updateLikeCount(echoId, 1)
}

// 监听筛选参数变化，重新加载数据
watch([filteredTag, filteredDate, filteredYearMonth], () => {
  // 当筛选参数变化时，重新加载数据
  if (filteredTag.value || filteredDate.value || filteredYearMonth.value) {
    echoStore.getEchosByPageForFilter()
  }
})

onMounted(async () => {
  // 获取数据
  echoStore.getEchosByPageForFilter()
})
</script>

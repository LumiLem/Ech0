<template>
  <div class="mx-auto px-2 sm:px-5 my-4">
    <div v-for="item in items" :key="item.id" class="mb-3">
      <TheInboxCard :inbox="item" />
    </div>

    <!-- 加载更多 -->
    <div v-if="hasMore && !loading" class="my-4 ml-1 flex items-center justify-start">
      <BaseButton
        @click="loadMore"
        class="rounded-full bg-[var(--timeline-load-more-bg-color)] !active:bg-[var(--timeline-load-more-active-bg-color)] mr-2"
      >
        <span
          class="text-[var(--timeline-load-more-text-color)] text-md font-serif text-center px-2 py-1"
          >加载更多</span
        >
      </BaseButton>
    </div>
    <!-- 没有更多 -->
    <div v-if="!hasMore && !loading" class="mx-auto my-5 text-center">
      <p class="text-xl text-[var(--text-color-400)] flex items-center justify-center">
        没有啦！<Flowers />
      </p>
    </div>
    <!-- 加载中 -->
    <div v-if="loading" class="mx-auto my-5 text-center">
      <span class="text-xl text-[var(--text-color-400)]">加载中...</span>
    </div>
  </div>
</template>
<script lang="ts" setup>
import { onMounted, onUnmounted } from 'vue'
import { storeToRefs } from 'pinia'
import { useInboxStore } from '@/stores'
import { fetchMarkInboxRead } from '@/service/api'
import TheInboxCard from '@/components/advanced/TheInboxCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Flowers from '@/components/icons/flowers.vue'

const inboxStore = useInboxStore()
const { items, hasMore, loading } = storeToRefs(inboxStore)
const { loadMore } = inboxStore

let timer: ReturnType<typeof setInterval>

onMounted(async () => {
  // 用户停留超过 1 秒则更新消息为已读
  timer = setInterval(() => {
    if (items.value.length > 0) {
      items.value.forEach((item) => {
        if (!item.read) {
          fetchMarkInboxRead(item.id).then(() => {
            item.read = true
          })
        }
      })
    }
  }, 1500)
})

onUnmounted(() => {
  clearInterval(timer)
})
</script>
<style scoped></style>

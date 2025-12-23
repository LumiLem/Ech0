<template>
  <div>
    <h2 class="text-[var(--text-color-500)] font-bold mb-1">收件箱模式</h2>

    <!-- 当前共有 -->
    <div class="text-[var(--text-color-400)] my-2 font-serif">
      <p>当前消息共有 {{ total }} 条</p>
    </div>

    <!-- 未读消息共有 -->
    <div class="text-[var(--text-color-400)] my-2 font-serif">
      <p>未读消息共有 {{ unreadItems.length }} 条</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useInboxStore } from '@/stores'

const inboxStore = useInboxStore()
const { unreadItems, total } = storeToRefs(inboxStore)
</script>

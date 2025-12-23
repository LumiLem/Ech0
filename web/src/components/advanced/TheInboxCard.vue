<template>
  <div
    class="widget flex flex-col gap-2 p-4 rounded-md ring-1 ring-[var(--ring-color)] ring-inset mx-auto shadow-sm hover:shadow-md"
  >
    <!-- Type -->
    <div class="flex items-center justify-between mb-1 font-serif">
      <div class="flex justify-start gap-2 items-center h-auto font-bold text-2xl">
        <Notification class="w-4 h-4" />
      </div>
      <div
        class="uppercase text-[var(--text-color-next-300)] text-xs font-mono bg-[var(--bg-color-next-100)] px-2 py-0.5 rounded-full"
      >
        {{ props.inbox.type }}
      </div>
    </div>

    <!-- Content -->
    <div>
      <p class="text-[var(--text-color-next-500)] text-sm whitespace-pre-wrap">
        {{ props.inbox.content }}
      </p>
    </div>

    <!-- Time & Source -->
    <div
      class="flex justify-between items-center font-serif mt-2 pt-2 border-t border-dashed border-[var(--border-color-400)]"
    >
      <div class="text-[var(--text-color-next-300)] text-xs font-mono">
        {{ formatDate(props.inbox.created_at) }}
      </div>
      <div class="text-xs">
        <span class="text-[var(--text-color-next-300)] uppercase">{{ props.inbox.source }}</span>
      </div>
    </div>
  </div>
</template>
<script lang="ts" setup>
import Notification from '../icons/ notification.vue'
import { formatDate } from '@/utils/other'
const props = defineProps<{
  inbox: App.Api.Inbox.Inbox
}>()
</script>
<style scoped></style>

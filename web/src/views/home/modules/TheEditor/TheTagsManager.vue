<template>
  <div class="py-4">
    <h2 class="text-[var(--text-color-500)] font-bold mb-3">标签管理</h2>
    <div class="flex flex-wrap gap-2">
      <div
        v-for="tag in tagList"
        :key="tag.id"
        class="flex items-center gap-1 border rounded-sm border-gray-300 border-dashed py-0.5 px-1 mb-1"
        style="white-space: nowrap"
      >
        <div
          @click="handleDeleteTag(tag.id)"
          class="hover:cursor-pointer text-[var(--text-color-400)] flex items-center justify-start gap-2"
        >
          <div>#</div>
          {{ tag.name }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useEchoStore } from '@/stores'
import { fetchDeleteTagById } from '@/service/api'
import { storeToRefs } from 'pinia'
import { useBaseDialog } from '@/composables/useBaseDialog'

const echoStore = useEchoStore()
const { tagList } = storeToRefs(echoStore)

const { openConfirm } = useBaseDialog()

const handleDeleteTag = (tagId: number) => {
  openConfirm({
    title: '确认删除该标签吗？',
    description: '删除后，所有使用该标签的内容将不再关联此标签',
    onConfirm: () => {
      fetchDeleteTagById(tagId).then((res) => {
        if (res.code === 1) {
          echoStore.getTags()
        }
      })
    },
  })
}
</script>

<style scoped></style>

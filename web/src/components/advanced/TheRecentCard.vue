<template>
  <div v-if="recent" class="px-9 md:px-11">
    <div
      class="widget rounded-md shadow-sm hover:shadow-md ring-1 ring-[var(--ring-color)] ring-inset p-4"
    >
      <h2 class="text-[var(--widget-title-color)] font-bold text-lg mb-1 flex items-center">
        <RecentIcon class="mr-2" /> 近况总结(AI)：
      </h2>

      <div v-if="!loading" class="text-[var(--text-color-next-500)] text-sm">
        {{ recent || '作者最近很神秘～' }}
      </div>
      <div v-else>
        <div class="text-[var(--text-color-next-500)] text-sm">生成中...</div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { fetchGetRecent } from '@/service/api'
import { onMounted, ref } from 'vue'
import RecentIcon from '../icons/recent.vue'

const recent = ref<string>('')
const loading = ref<boolean>(true)

onMounted(() => {
  fetchGetRecent()
    .then((res) => {
      if (res.code === 1) {
        recent.value = res.data
      }
    })
    .finally(() => {
      loading.value = false
    })
})
</script>
<style scoped></style>

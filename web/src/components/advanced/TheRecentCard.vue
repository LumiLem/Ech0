<template>
  <div v-if="AgentSetting.enable" class="px-9 md:px-11">
    <div
      class="widget rounded-md shadow-sm hover:shadow-md ring-1 ring-[var(--ring-color)] ring-inset p-4"
    >
      <h2 class="text-[var(--widget-title-color)] font-bold text-lg mb-1 flex items-center">
        <RecentIcon class="mr-2" /> 近况总结(AI)：
      </h2>

      <div v-if="!loading" class="text-[var(--text-color-next-500)] text-sm p-1">
        <TheMdPreview :content="recent" />
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
import TheMdPreview from './TheMdPreview.vue'
import { useSettingStore } from '@/stores/setting'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { AgentSetting } = storeToRefs(settingStore)

const recent = ref<string>('作者最近很神秘～')
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
<style scoped>
.md-editor-dark,
.md-editor-modal-container[data-theme='dark'] {
  --md-bk-color: #212121 !important;
}

:deep(#preview-only-preview) p {
  color: var(--text-color-next-500) !important;
  font-size: var(--text-sm-fontSize) !important;
  line-height: var(--un-leading, var(--text-sm-lineHeight)) !important;
}
</style>

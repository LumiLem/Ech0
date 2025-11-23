<template>
  <PanelCard>
    <!-- 评论设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--text-color-600)] font-bold text-lg">评论设置</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button v-if="commentEditMode" @click="handleUpdateCommentSetting" title="编辑">
            <Saveupdate class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
          <button @click="commentEditMode = !commentEditMode" title="编辑">
            <Edit
              v-if="!commentEditMode"
              class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6"
            />
            <Close v-else class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>

      <!-- 开启评论 -->
      <div class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] h-10">
        <h2 class="font-semibold w-24 shrink-0">启用评论:</h2>
        <BaseSwitch v-model="CommentSetting.enable_comment" :disabled="!commentEditMode" />
      </div>

      <!-- 评论服务 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-24 shrink-0">评论服务:</h2>
        <BaseSelect
          v-model="CommentSetting.provider"
          :options="commentServiceOptions"
          :disabled="!commentEditMode"
          class="w-34 h-8"
        />
      </div>

      <!-- 评论 API -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-24 shrink-0">评论 API:</h2>
        <span
          v-if="!commentEditMode"
          class="truncate max-w-40 inline-block align-middle"
          :title="CommentSetting.comment_api"
          style="vertical-align: middle"
        >
          {{ CommentSetting.comment_api.length === 0 ? '暂无' : CommentSetting.comment_api }}
        </span>
        <BaseInput
          v-else
          v-model="CommentSetting.comment_api"
          type="text"
          placeholder="评论 API地址,带http(s)"
          class="w-full py-1!"
        />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted } from 'vue'
import { fetchUpdateCommentSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores/setting'
import { storeToRefs } from 'pinia'
import { CommentProvider } from '@/enums/enums'

const settingStore = useSettingStore()
const { getCommentSetting } = settingStore
const { CommentSetting } = storeToRefs(settingStore)

const commentEditMode = ref<boolean>(false)

const commentServiceOptions = ref<{ label: string; value: CommentProvider }[]>([
  { label: 'Twikoo', value: CommentProvider.TWIKOO },
  // { label: 'Artalk', value: CommentProvider.ARTALK },
  // { label: 'Waline', value: CommentProvider.WALINE },
  // { label: 'Giscus', value: CommentProvider.GISCUS },
])

const handleUpdateCommentSetting = async () => {
  await fetchUpdateCommentSettings(settingStore.CommentSetting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      commentEditMode.value = false
      // 重新获取评论设置
      getCommentSetting()
    })
}

onMounted(() => {
  getCommentSetting()
})
</script>

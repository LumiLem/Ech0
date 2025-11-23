<template>
  <PanelCard>
    <!-- 联邦宇宙设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--text-color-600)] font-bold text-lg">联邦宇宙</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button v-if="fediverseEditMode" @click="handleUpdateFediverseSetting" title="保存">
            <Saveupdate class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
          <button @click="fediverseEditMode = !fediverseEditMode" title="编辑">
            <Edit
              v-if="!fediverseEditMode"
              class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6"
            />
            <Close v-else class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>

      <!-- 启用联邦宇宙 -->
      <div class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] h-10">
        <h2 class="font-semibold shrink-0 mr-1">启用主动推送:</h2>
        <BaseSwitch v-model="FediverseSetting.enable" :disabled="!fediverseEditMode" />
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted } from 'vue'
import { fetchUpdateFediverseSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores/setting'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { getFediverseSetting } = settingStore
const { FediverseSetting } = storeToRefs(settingStore)

const fediverseEditMode = ref<boolean>(false)

const handleUpdateFediverseSetting = async () => {
  const res = await fetchUpdateFediverseSettings(settingStore.FediverseSetting)

  if (res.code === 1) {
    theToast.success(res.msg)
  }

  fediverseEditMode.value = false
  getFediverseSetting()
}

onMounted(() => {
  getFediverseSetting()
})
</script>

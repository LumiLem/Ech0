<template>
  <PanelCard>
    <!-- Agent 设置 -->
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--text-color-600)] font-bold text-lg">Agent 设置</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button v-if="agentEditMode" @click="handleUpdateAgentSetting" title="保存">
            <Saveupdate class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
          <button @click="agentEditMode = !agentEditMode" title="编辑">
            <Edit
              v-if="!agentEditMode"
              class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6"
            />
            <Close v-else class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>

      <!-- 启用 Agent -->
      <div class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] h-10">
        <h2 class="font-semibold w-24 shrink-0">启用 Agent:</h2>
        <BaseSwitch v-model="AgentSetting.enable" :disabled="!agentEditMode" />
      </div>

      <!-- LLM 提供商 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-24 shrink-0">提供商:</h2>
        <BaseSelect
          v-model="AgentSetting.provider"
          :options="agentProviderOptions"
          :disabled="!agentEditMode"
          class="w-40 h-8"
        />
      </div>

      <!-- 模型名称 -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-24 shrink-0">模型名称:</h2>
        <span
          v-if="!agentEditMode"
          class="truncate max-w-60 inline-block align-middle"
          :title="AgentSetting.model"
        >
          {{ AgentSetting.model || '暂无' }}
        </span>
        <div v-else>
          <BaseInput
            v-model="AgentSetting.model"
            type="text"
            placeholder="输入模型名称"
            class="w-full py-1!"
          />
        </div>
      </div>

      <!-- API Key -->
      <div
        class="flex flex-row items-center justify-start text-[var(--text-color-next-500)] gap-2 h-10"
      >
        <h2 class="font-semibold w-24 shrink-0">API Key:</h2>
        <span v-if="!agentEditMode" class="truncate max-w-60 inline-block align-middle">
          {{ AgentSetting.api_key ? '********' : '暂无' }}
        </span>
        <BaseInput
          v-else
          v-model="AgentSetting.api_key"
          type="password"
          placeholder="输入 API Key"
          class="w-full py-1!"
        />
      </div>

      <!-- 自定义 Base URL -->
      <div class="flex justify-start text-[var(--text-color-next-500)] gap-2 mt-2">
        <h2 class="font-semibold w-24 shrink-0">Base URL:</h2>
        <span v-if="!agentEditMode" class="truncate max-w-60 inline-block align-middle">
          {{ AgentSetting.base_url.length == 0 ? '暂无' : AgentSetting.base_url }}
        </span>
        <BaseInput
          v-if="agentEditMode"
          v-model="AgentSetting.base_url"
          placeholder="输入自定义 Base URL"
          class="w-full py-1!"
        />
      </div>

      <!-- Prompt -->
      <div class="flex justify-start text-[var(--text-color-next-500)] gap-2 mt-2">
        <h2 class="font-semibold w-24 shrink-0">Prompt:</h2>
        <span v-if="!agentEditMode" class="truncate max-w-60 inline-block align-middle">
          {{ AgentSetting.prompt.length == 0 ? '暂无' : '' }}
        </span>
        <BaseTextArea
          v-if="agentEditMode"
          v-model="AgentSetting.prompt"
          placeholder="输入自定义 Prompt"
          class="w-full"
          :rows="4"
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
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, onMounted } from 'vue'
import { fetchUpdateAgentSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { AgentProvider } from '@/enums/enums'

const settingStore = useSettingStore()
const { getAgentSetting } = settingStore
const { AgentSetting } = storeToRefs(settingStore)

const agentEditMode = ref<boolean>(false)

const agentProviderOptions = ref<{ label: string; value: AgentProvider }[]>([
  { label: 'OpenAI', value: AgentProvider.OPENAI },
  { label: 'DeepSeek', value: AgentProvider.DEEPSEEK },
  { label: 'Anthropic', value: AgentProvider.ANTHROPIC },
  { label: 'Gemini', value: AgentProvider.GEMINI },
  { label: 'Qwen', value: AgentProvider.QWEN },
  { label: 'Ollama', value: AgentProvider.OLLAMA },
  { label: '自定义', value: AgentProvider.CUSTOM },
])

const handleUpdateAgentSetting = async () => {
  await fetchUpdateAgentSettings(settingStore.AgentSetting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      agentEditMode.value = false
      // 重新获取 Agent 设置
      getAgentSetting()
    })
}

onMounted(() => {
  getAgentSetting()
})
</script>

<style scoped></style>

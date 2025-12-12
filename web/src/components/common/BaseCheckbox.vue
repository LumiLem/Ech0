<template>
  <div class="base-checkbox">
    <label
      :for="id"
      class="inline-flex items-center cursor-pointer select-none"
      :class="[
        disabled ? 'cursor-not-allowed opacity-70' : 'hover:opacity-80',
        customClass
      ]"
    >
      <div class="relative">
        <input
          :id="id"
          type="checkbox"
          :checked="currentValue"
          :disabled="disabled"
          class="sr-only"
          @change="onChange"
        />
        <div
          class="checkbox-box"
          :class="{
            'checkbox-checked': currentValue,
            'checkbox-disabled': disabled
          }"

        >
          <!-- 勾选图标 -->
          <svg
            v-if="currentValue"
            class="checkbox-icon"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fill-rule="evenodd"
              d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
              clip-rule="evenodd"
            />
          </svg>
        </div>
      </div>
      
      <!-- 标签文本 -->
      <span
        v-if="$slots.default || label"
        class="checkbox-label"
        :class="{ 'text-[var(--text-color-400)]': disabled }"
      >
        <slot>{{ label }}</slot>
      </span>
    </label>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  modelValue?: boolean  // 可选，支持非受控模式
  id?: string
  label?: string
  disabled?: boolean
  class?: string
  defaultChecked?: boolean  // 非受控模式的默认值
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'change', event: Event): void
}>()

const customClass = props.class

// 内部状态，用于非受控模式
const internalValue = ref(props.defaultChecked || false)

// 当前值：如果传入了 modelValue 则使用它（受控模式），否则使用内部状态（非受控模式）
const currentValue = computed(() => {
  return props.modelValue !== undefined ? props.modelValue : internalValue.value
})

// 监听 modelValue 变化，同步到内部状态
watch(() => props.modelValue, (newValue) => {
  if (newValue !== undefined) {
    internalValue.value = newValue
  }
}, { immediate: true })

function onChange(event: Event) {
  if (!props.disabled) {
    const target = event.target as HTMLInputElement
    const checked = target.checked
    
    // 如果是非受控模式，更新内部状态
    if (props.modelValue === undefined) {
      internalValue.value = checked
    }
    
    // 总是发出事件
    emit('update:modelValue', checked)
    emit('change', event)
  }
}
</script>

<style scoped>
.base-checkbox {
  display: inline-block;
}

.checkbox-box {
  width: 1rem;
  height: 1rem;
  border: 1px solid var(--input-border-color);
  border-radius: 0.25rem;
  background-color: var(--input-bg-color);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease-in-out;
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
}

.checkbox-box:hover {
  border-color: #f97316;
}

.checkbox-checked {
  background-color: #f97316;
  border-color: #f97316;
  color: white;
}

.checkbox-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.checkbox-disabled:hover {
  border-color: var(--input-border-color);
}

.checkbox-icon {
  width: 0.75rem;
  height: 0.75rem;
}

.checkbox-label {
  margin-left: 0.5rem;
  font-size: 0.875rem;
  color: var(--text-color-700);
  line-height: 1.25rem;
}

/* 隐藏原生复选框但保持可访问性 */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
<template>
  <div class="base-combobox">
    <!-- Label -->
    <label
      v-if="label"
      :for="id"
      class="block text-sm font-medium text-[var(--text-color-700)] mb-1"
    >
      {{ label }}
    </label>

    <!-- Combobox Wrapper -->
    <Combobox
      :modelValue="internalValue"
      :by="by"
      :multiple="multiple"
      :nullable="true"
      @update:model-value="onSelect"
    >
      <div class="relative">
        <!-- Input -->
        <div
          class="flex items-center px-0.5 py-0.5 rounded-md bg-[var(--input-bg-color)] transition duration-150 ease-in-out"
          @focusout="onBlurOutside"
          @focusin="onFocusInput"
          @mousedown="onFocusInput"
        >
          <ComboboxInput
            :displayValue="displayValue"
            :placeholder="placeholder"
            @focus="onFocusInput"
            @click="onFocusInput"
            @input="onInputChange"
            :class="['outline-none text-md', inputClass]"
          />

          <!-- 可选的 suffix slot -->
          <slot name="suffix">
            <ComboboxButton class="ml-1 text-[var(--text-color-400)]">
              <!-- <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24">
                <path
                  fill="#888888"
                  d="m12 19.15l3.875-3.875q.3-.3.7-.3t.7.3t.3.713t-.3.712l-3.85 3.875q-.575.575-1.425.575t-1.425-.575L6.7 16.7q-.3-.3-.288-.712t.313-.713t.713-.3t.712.3zm0-14.3L8.15 8.7q-.3.3-.7.288t-.7-.288q-.3-.3-.312-.712t.287-.713l3.85-3.85Q11.15 2.85 12 2.85t1.425.575l3.85 3.85q.3.3.288.713t-.313.712q-.3.275-.7.288t-.7-.288z"
                />
              </svg> -->
            </ComboboxButton>
          </slot>
        </div>

        <!-- Dropdown -->
        <Transition
          enter="transition ease-out duration-100"
          enter-from="opacity-0 translate-y-1"
          enter-to="opacity-100 translate-y-0"
          leave="transition ease-in duration-75"
          leave-from="opacity-100 translate-y-0"
          leave-to="opacity-0 translate-y-1"
        >
          <ComboboxOptions
            static
            v-if="dropdownOpen && (filteredOptions.length > 0 || allowCreate)"
            class="w-auto! absolute z-10 mt-2 max-h-64 overflow-y-scroll rounded-md bg-[var(--combobox-bg-color)] py-1 text-sm shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none"
          >
            <!-- Existing Options -->
            <div
              @click="toggleManaging"
              class="px-2 py-1 w-36 text-[var(--text-color-400)] cursor-pointer flex items-center justify-start gap-2 text-lg font-bold mx-auto hover:text-amber-500"
            >
              <tagsetting class="w-4 h-4" />
              标签管理
              <jump class="w-4 h-4" />
            </div>
            <ComboboxOption
              v-for="item in filteredOptions"
              :key="getOptionLabel(item) || String(item)"
              :value="item"
              @mousedown="isUserClicking = true"
              class="w-full! max-w-32! truncate text-[var(--text-color-300)] hover:text-[var(--text-color-500)] text-lg cursor-pointer select-none px-4 py-1 whitespace-nowrap text-ellipsis"
            >
              <slot name="option" :option="item"> # {{ getOptionLabel(item) }} </slot>
            </ComboboxOption>
          </ComboboxOptions>
        </Transition>
      </div>
    </Combobox>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  Combobox,
  ComboboxInput,
  ComboboxOptions,
  ComboboxOption,
  ComboboxButton,
} from '@headlessui/vue'
import tagsetting from '../icons/tagsetting.vue'
import jump from '../icons/jump.vue'
import { useEditorStore } from '@/stores'
import { Mode } from '@/enums/enums'

type ClassValue = string | string[] | Record<string, boolean | number | string>

const props = defineProps<{
  /** 绑定到外部的值，支持单选或多选 */
  modelValue: string | object | null | (string | object)[]
  /** 可供选择的选项列表 */
  options: string[] | object[]
  /** 输入框上方显示的标签文本 */
  label?: string
  /** 关联 label 与输入框的 id */
  id?: string
  /** 输入框提示文本 */
  placeholder?: string
  /** 自定义选项比对逻辑或字段名 */
  by?: string | ((a: object | string, b: object | string) => boolean)
  /** 显示选项时使用的字段名 */
  labelField?: string
  /** 是否允许创建新选项 */
  allowCreate?: boolean
  /** 是否启用多选模式 */
  multiple?: boolean
  /** 输入框额外的样式类 */
  inputClass?: ClassValue
}>()

const emit = defineEmits(['update:modelValue', 'create'])

const query = ref('')
const dropdownOpen = ref(false)
const internalValue = ref(props.modelValue)
const labelField = props.labelField || 'name'
const allowCreate = props.allowCreate ?? false
const multiple = props.multiple ?? false
const isManaging = ref<boolean>(false)
const isUserClicking = ref(false) // 标记用户是否正在主动点击选项
const editorStore = useEditorStore()

const toggleManaging = () => {
  isManaging.value = !isManaging.value
  console.log(isManaging.value)
  editorStore.setMode(Mode.TagManage)
}

watch(
  () => props.modelValue,
  (val) => {
    internalValue.value = val
  },
)
watch(internalValue, (val) => {
  emit('update:modelValue', val)
})

const onSelect = (val: object | string) => {
  const selectedLabel = getOptionLabel(val)

  // 只有当用户主动点击选项时才接受选择并关闭下拉框
  if (isUserClicking.value) {
    internalValue.value = val
    query.value = selectedLabel
    dropdownOpen.value = false
    isUserClicking.value = false
    return
  }

  // 自动选择（非用户点击）：只有精确匹配时才接受
  if (query.value && selectedLabel.toLowerCase() === query.value.toLowerCase()) {
    internalValue.value = val
    query.value = selectedLabel
    // 不关闭下拉框
    return
  }

  // 不精确匹配，忽略这次选择，不做任何改变
}

const onInputChange = (e: Event) => {
  const value = (e.target as HTMLInputElement).value.trim()
  query.value = value

  // 输入框被清空时 -> 清空绑定值
  if (value === '') {
    internalValue.value = multiple ? [] : null
    emit('update:modelValue', internalValue.value)
    dropdownOpen.value = true
    return
  }

  // 如果输入内容刚好匹配某个现有选项 -> 自动选择该项
  const matched = props.options.find(
    (option) => getOptionLabel(option).toLowerCase() === value.toLowerCase(),
  )
  if (matched) {
    internalValue.value = matched
    emit('update:modelValue', matched)
  } else {
    // 否则表示用户正在输入新的标签
    internalValue.value = value
    emit('create', value) // 可选：通知外部准备创建
    emit('update:modelValue', internalValue.value)
  }

  // 无论如何都保持下拉框打开
  dropdownOpen.value = true
}

const onFocusInput = () => {
  dropdownOpen.value = true
}

const onBlurOutside = (e: FocusEvent) => {
  // 确保焦点确实离开整个 Combobox（不是内部选项）
  const currentTarget = e.currentTarget as HTMLElement
  if (!currentTarget.contains(e.relatedTarget as Node)) {
    dropdownOpen.value = false
    if (query.value.trim() === '') {
      internalValue.value = multiple ? [] : null
      emit('update:modelValue', internalValue.value)
    }
  }
}

const getOptionLabel = (option: unknown): string => {
  if (option == null) return ''
  if (typeof option === 'object' && !Array.isArray(option)) {
    const record = option as Record<string, unknown>
    const candidate = record[labelField]
    if (candidate != null) return String(candidate)
  }
  return String(option ?? '')
}

const normalizedQuery = computed(() => query.value.trim().toLowerCase())

const filteredOptions = computed(() => {
  if (!normalizedQuery.value) return props.options
  const lowerQuery = normalizedQuery.value
  return props.options.filter((option) => getOptionLabel(option).toLowerCase().includes(lowerQuery))
})

const displayValue = (item: unknown) => {
  if (Array.isArray(item)) return item.map((i) => getOptionLabel(i)).join(', ')
  return getOptionLabel(item)
}
</script>

<style scoped>
.base-combobox {
  display: flex;
  flex-direction: column;
}
</style>

<template>
  <div class="base-select" ref="selectRef">
    <!-- Label -->
    <label
      v-if="label"
      :for="id"
      class="block text-sm font-medium text-[var(--select-label-text-color)] mb-1"
    >
      {{ label }}
    </label>

    <!-- Select Button -->
    <div class="relative inline-block">
      <button
        :id="id"
        type="button"
        :disabled="disabled"
        :class="[
          'inline-flex items-center justify-between px-3 py-2 rounded-lg border border-[var(--select-border-color)] focus:outline-none focus:ring-2 focus:ring-orange-300 focus:border-orange-300 transition duration-150 ease-in-out shadow-xs sm:text-sm text-left',
          disabled
            ? 'bg-[var(--select-disabled-bg-color)] cursor-not-allowed opacity-70'
            : 'bg-[var(--select-bg-color)] hover:border-orange-400 cursor-pointer',
          customClass,
        ]"
        @click="onToggle"
        @keydown.space.prevent="onToggle"
        @keydown.enter.prevent="onToggle"
        @keydown.up.prevent="onNavigate(-1)"
        @keydown.down.prevent="onNavigate(1)"
        @keydown.escape="onClose"
      >
        <!-- Selected Value Display -->
        <span
          :class="[
            'truncate',
            !selectedOption && placeholder
              ? 'text-[var(--text-color-500)]'
              : 'text-[var(--text-color-600)]',
          ]"
        >
          {{ displayValue }}
        </span>

        <!-- Dropdown Arrow -->
        <svg
          :class="[
            'w-8 text-yellow-200 transition-transform duration-200',
            isOpen ? 'rotate-180' : '',
          ]"
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
        >
          <!-- Icon from Material Symbols by Google - https://github.com/google/material-design-icons/blob/master/LICENSE -->
          <path fill="#888888" d="m12 15.4l-6-6L7.4 8l4.6 4.6L16.6 8L18 9.4z" />
        </svg>
      </button>

      <!-- Dropdown Menu -->
      <Transition
        enter-active-class="transition ease-out duration-100"
        enter-from-class="transform opacity-0 scale-95"
        enter-to-class="transform opacity-100 scale-100"
        leave-active-class="transition ease-in duration-75"
        leave-from-class="transform opacity-100 scale-100"
        leave-to-class="transform opacity-0 scale-95"
      >
        <div
          v-show="isOpen"
          class="absolute z-5000 mt-1 min-w-full bg-[var(--select-bg-color)] shadow-lg max-h-70 rounded-lg border border-[var(--select-border-color)] overflow-auto focus:outline-none"
        >
          <div
            v-for="(option, index) in normalizedOptions"
            :key="String(getOptionValue(option) ?? index)"
            :class="[
              'cursor-pointer select-none relative px-3 py-2 text-sm',
              index === highlightedIndex
                ? 'bg-[var(--select-label-hover-bg-color)] text-orange-900'
                : 'text-[var(--text-color-900)] hover:bg-[var(--select-label-clicked-bg-color)]',
              isSelected(option) ? 'font-medium' : 'font-normal',
            ]"
            @click="onSelect(option)"
            @mouseenter="highlightedIndex = index"
          >
            <div class="flex items-center justify-between">
              <span class="truncate text-[var(--text-color-500)] font-bold">{{
                getOptionLabel(option)
              }}</span>
              <!-- Check Icon for Selected -->
              <svg
                v-if="isSelected(option)"
                class="text-orange-500"
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
              >
                <!-- Icon from Typicons by Stephen Hutchings - https://creativecommons.org/licenses/by-sa/4.0/ -->
                <path
                  fill="#888888"
                  d="M16.972 6.251a2 2 0 0 0-2.72.777l-3.713 6.682l-2.125-2.125a2 2 0 1 0-2.828 2.828l4 4c.378.379.888.587 1.414.587l.277-.02a2 2 0 0 0 1.471-1.009l5-9a2 2 0 0 0-.776-2.72"
                />
              </svg>
            </div>
          </div>

          <!-- Empty State -->
          <div
            v-if="normalizedOptions.length === 0"
            class="px-3 py-2 text-sm text-[var(--text-color-500)] text-center"
          >
            {{ emptyText }}
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'

// 定义值的类型
type SelectValue = string | number | boolean | null | undefined

// 定义选项接口
interface SelectOption {
  label: string
  value: SelectValue
  disabled?: boolean
}

// 定义通用对象类型用于自定义键名
interface CustomKeyOption {
  [key: string]: unknown
}

// 定义选项类型联合
type OptionType = SelectOption | string | number | CustomKeyOption

const props = defineProps<{
  modelValue: SelectValue
  options: OptionType[]
  id?: string
  label?: string
  placeholder?: string
  disabled?: boolean
  class?: string
  emptyText?: string
  labelKey?: string
  valueKey?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: SelectValue): void
  (e: 'change', value: SelectValue): void
  (e: 'open'): void
  (e: 'close'): void
}>()

// Refs
const selectRef = ref<HTMLElement>()
const isOpen = ref(false)
const highlightedIndex = ref(-1)

// Computed
const customClass = props.class
const emptyText = props.emptyText || '暂无选项'

const normalizedOptions = computed((): SelectOption[] => {
  return props.options.map((option): SelectOption => {
    if (typeof option === 'string' || typeof option === 'number') {
      return { label: String(option), value: option }
    }

    // 如果是对象类型，检查是否已经是 SelectOption 格式
    if (typeof option === 'object' && option !== null) {
      const objOption = option as Record<string, unknown>

      // 如果有自定义键名
      if (props.labelKey && props.valueKey) {
        return {
          label: String(objOption[props.labelKey] || ''),
          value: objOption[props.valueKey] as SelectValue,
          disabled: objOption.disabled as boolean | undefined,
        }
      }

      // 如果是标准的 SelectOption 格式
      if ('label' in objOption && 'value' in objOption) {
        return option as SelectOption
      }
    }

    // 兜底情况
    return { label: String(option), value: option as unknown as SelectValue }
  })
})

const selectedOption = computed(() => {
  return normalizedOptions.value.find((option) => option.value === props.modelValue)
})

const displayValue = computed(() => {
  return selectedOption.value?.label || props.placeholder || '请选择'
})

// Methods
function getOptionLabel(option: SelectOption): string {
  return option.label
}

function getOptionValue(option: SelectOption): SelectValue {
  return option.value
}

function isSelected(option: SelectOption): boolean {
  return getOptionValue(option) === props.modelValue
}

function onToggle(): void {
  if (props.disabled) return

  if (isOpen.value) {
    onClose()
  } else {
    onOpen()
  }
}

function onOpen(): void {
  isOpen.value = true
  highlightedIndex.value = normalizedOptions.value.findIndex((option) => isSelected(option))
  emit('open')
}

function onClose(): void {
  isOpen.value = false
  highlightedIndex.value = -1
  emit('close')
}

function onSelect(option: SelectOption): void {
  if (option.disabled) return

  const value = getOptionValue(option)
  emit('update:modelValue', value)
  emit('change', value)
  onClose()
}

function onNavigate(direction: number): void {
  if (!isOpen.value) {
    onOpen()
    return
  }

  const optionsCount = normalizedOptions.value.length
  if (optionsCount === 0) return

  let newIndex = highlightedIndex.value + direction

  if (newIndex < 0) {
    newIndex = optionsCount - 1
  } else if (newIndex >= optionsCount) {
    newIndex = 0
  }

  highlightedIndex.value = newIndex
}

// Handle clicks outside to close dropdown
function handleClickOutside(event: Event): void {
  if (selectRef.value && event.target instanceof Node && !selectRef.value.contains(event.target)) {
    onClose()
  }
}

// Lifecycle
onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

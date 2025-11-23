<template>
  <div class="base-input w-full">
    <!-- Label -->
    <label
      v-if="label"
      :for="id"
      class="block text-sm font-medium text-[var(--text-color-700)] mb-1"
    >
      {{ label }}
    </label>

    <!-- Input Wrapper -->
    <div class="flex items-center">
      <slot name="prefix" />
      <input
        :id="id"
        :type="type"
        :value="modelValue"
        :placeholder="placeholder"
        :disabled="disabled"
        :readonly="readonly"
        :class="[
          'block px-3 py-2 rounded-lg border border-[var(--input-border-color)] text-[var(--placeholder-color)] bg-[var(--input-bg-color)]! focus:outline-none focus:ring-2 focus:ring-orange-300 focus:border-orange-300 transition duration-150 ease-in-out shadow-xs sm:text-sm  w-full',
          disabled
            ? 'cursor-not-allowed opacity-70 text-[var(--text-color-400)]'
            : 'hover:border-orange-400 focus:text-[var(--text-color-700)]',
          customClass,
        ]"
        @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
        v-bind="$attrs"
      />
      <slot name="suffix" />
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: string | number | null | undefined
  id?: string
  label?: string
  placeholder?: string
  type?: string // 默认值为 'text'
  disabled?: boolean
  readonly?: boolean
  class?: string
}>()

const customClass = props.class
const type = props.type || 'text'
</script>

<style scoped>
.base-input {
  display: flex;
  flex-direction: column;
}

input {
  outline: none;
}
</style>

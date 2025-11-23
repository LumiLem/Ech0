<!-- ConfirmDialog.vue -->
<template>
  <TransitionRoot :show="isOpen" as="template">
    <Dialog @close="close" class="relative z-5000">
      <!-- 背景遮罩 -->
      <TransitionChild
        enter="duration-300 ease-out"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="duration-200 ease-in"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-black/30" aria-hidden="true" />
      </TransitionChild>

      <!-- 对话框面板 -->
      <div class="fixed inset-0 flex items-center justify-center p-4">
        <TransitionChild
          enter="duration-300 ease-out"
          enter-from="opacity-0 scale-95"
          enter-to="opacity-100 scale-100"
          leave="duration-200 ease-in"
          leave-from="opacity-100 scale-100"
          leave-to="opacity-0 scale-95"
        >
          <DialogPanel
            class="w-full max-w-sm rounded-lg bg-[var(--dialog-bg-color)] p-6 shadow-sm ring-1 ring-inset ring-[var(--ring-color)]"
          >
            <DialogTitle class="text-base font-semibold text-[var(--dialog-title-color)]">
              {{ title }}
            </DialogTitle>
            <DialogDescription class="mt-2 text-sm text-[var(--dialog-text-color)] leading-relaxed">
              {{ description }}
            </DialogDescription>

            <div class="mt-6 flex justify-end gap-3">
              <button
                @click="cancel"
                class="cursor-pointer px-3 py-2 rounded-lg bg-[var(--dialog-cancel-btn-bg-color)] shadow-xs ring-1 ring-inset ring-[var(--ring-color)] text-[var(--dialog-btn-text-color)] hover:text-orange-400"
              >
                取消
              </button>
              <button
                @click="confirm"
                class="cursor-pointer px-3 py-2 rounded-lg bg-[var(--dialog-confirm-btn-bg-color)] text-white shadow-xs hover:bg-orange-500"
              >
                确认
              </button>
            </div>
          </DialogPanel>
        </TransitionChild>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import {
  Dialog,
  DialogPanel,
  DialogTitle,
  DialogDescription,
  TransitionChild,
  TransitionRoot,
} from '@headlessui/vue'

defineProps({
  title: String,
  description: String,
})

const emit = defineEmits(['confirm', 'cancel'])

const isOpen = ref(false)

function open() {
  isOpen.value = true
}

function close() {
  isOpen.value = false
}

function confirm() {
  emit('confirm')
  close()
}

function cancel() {
  emit('cancel')
  close()
}

defineExpose({ open })
</script>

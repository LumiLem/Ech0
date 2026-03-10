<template>
  <TransitionRoot :show="isOpen" as="template">
    <Dialog @close="close" class="relative z-5000">
      <!-- Background overlay -->
      <TransitionChild
        enter="duration-300 ease-out"
        enter-from="opacity-0"
        enter-to="opacity-100"
        leave="duration-200 ease-in"
        leave-from="opacity-100"
        leave-to="opacity-0"
      >
        <div class="fixed inset-0 bg-black/40 backdrop-blur-sm" aria-hidden="true" />
      </TransitionChild>

      <!-- Panel -->
      <div class="fixed inset-0 flex items-center justify-center p-4">
        <TransitionChild
          enter="duration-300 ease-out"
          enter-from="opacity-0 scale-95 translate-y-4"
          enter-to="opacity-100 scale-100 translate-y-0"
          leave="duration-200 ease-in"
          leave-from="opacity-100 scale-100"
          leave-to="opacity-0 scale-95 translate-y-4"
        >
          <DialogPanel
            class="w-full max-w-md rounded-2xl bg-white/95 dark:bg-[var(--dialog-bg-color)]/95 backdrop-blur-md p-6 shadow-2xl ring-1 ring-inset ring-[var(--ring-color)] flex flex-col gap-6 relative overflow-hidden"
          >
            <!-- Decorative gradient background -->
            <div class="absolute inset-0 bg-gradient-to-br from-orange-100/40 via-purple-50/20 to-blue-50/40 dark:from-orange-900/10 dark:via-purple-900/5 dark:to-blue-900/10 pointer-events-none z-0"></div>

            <div class="relative z-10 flex items-center justify-between">
              <div class="w-8"></div> <!-- Spacer for center alignment -->
              <DialogTitle class="text-xl font-semibold text-gray-800 dark:text-gray-200 flex items-center gap-2">
                AI 写作
              </DialogTitle>
              <button @click="close" class="w-8 h-8 flex items-center justify-center rounded-full bg-gray-100 dark:bg-gray-800 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 transition">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
              </button>
            </div>
            
            <!-- Default Home View -->
            <div v-if="!isPolishing && !polishedResult" class="relative z-10 flex flex-col gap-6">
              
              <!-- Generate Input Area -->
              <div class="relative flex items-center bg-white/80 dark:bg-gray-800/80 backdrop-blur rounded-2xl p-2 shadow-sm border border-gray-100 dark:border-gray-700">
                <div class="pl-2 pr-1 text-orange-400">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/><path d="M20 3v4"/><path d="M22 5h-4"/><path d="M4 17v2"/><path d="M5 18H3"/></svg>
                </div>
                <input 
                  v-model="customPrompt" 
                  @keyup.enter="startAIAction('generate')"
                  type="text" 
                  placeholder="输入需求，AI帮您创作" 
                  class="w-full bg-transparent border-none outline-none text-sm px-2 py-1 text-gray-700 dark:text-gray-200 placeholder-gray-400"
                />
                <button v-if="customPrompt" @click="startAIAction('generate')" class="bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 rounded-xl px-3 py-1.5 text-xs font-medium cursor-pointer hover:opacity-90 transition whitespace-nowrap">
                  发送
                </button>
              </div>

              <!-- Main Action Cards -->
              <div class="grid grid-cols-3 gap-3">
                <button @click="startAIAction('summarize')" class="flex flex-col items-center justify-center p-4 bg-white/80 dark:bg-gray-800/80 backdrop-blur rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition cursor-pointer group">
                  <div class="mb-3 text-gray-700 dark:text-gray-300 group-hover:scale-110 transition-transform">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-7 h-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="6" x2="20" y2="6"/><line x1="4" y1="12" x2="14" y2="12"/><line x1="4" y1="18" x2="18" y2="18"/><path d="m19 11 1.5 1.5-1.5 1.5"/><path d="m19 11-1.5 1.5 1.5 1.5"/></svg>
                  </div>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">摘要</span>
                </button>
                
                <button @click="startAIAction('correct')" class="flex flex-col items-center justify-center p-4 bg-white/80 dark:bg-gray-800/80 backdrop-blur rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition cursor-pointer group">
                  <div class="mb-3 text-gray-700 dark:text-gray-300 group-hover:scale-110 transition-transform relative">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-7 h-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>
                    <span class="absolute top-1/2 left-1/2 -translate-x-[60%] -translate-y-1/2 text-[10px] font-bold">?</span>
                  </div>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">纠错</span>
                </button>
                
                <button @click="startAIAction('expand')" class="flex flex-col items-center justify-center p-4 bg-white/80 dark:bg-gray-800/80 backdrop-blur rounded-2xl shadow-sm border border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition cursor-pointer group">
                  <div class="mb-3 text-gray-700 dark:text-gray-300 group-hover:scale-110 transition-transform">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-7 h-7" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/><path d="m11 13-2 2 2 2"/><path d="m15 17 2-2-2-2"/></svg>
                  </div>
                  <span class="text-sm font-medium text-gray-800 dark:text-gray-200">扩写</span>
                </button>
              </div>

              <!-- Text Polish Sections -->
              <div>
                <h3 class="text-sm font-medium text-gray-400 dark:text-gray-500 mb-3 ml-1">文本润色</h3>
                <div class="grid grid-cols-4 gap-2">
                  <button @click="startAIAction('polish', 'general')" class="flex flex-col items-center gap-2 group cursor-pointer">
                    <div class="w-14 h-14 rounded-full bg-white/80 dark:bg-gray-800/80 shadow-sm border border-gray-100 dark:border-gray-700 flex items-center justify-center group-hover:shadow-md transition">
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-gray-700 dark:text-gray-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>
                    </div>
                    <span class="text-xs text-gray-600 dark:text-gray-400">通用的</span>
                  </button>

                  <button @click="startAIAction('polish', 'professional')" class="flex flex-col items-center gap-2 group cursor-pointer">
                    <div class="w-14 h-14 rounded-full bg-white/80 dark:bg-gray-800/80 shadow-sm border border-gray-100 dark:border-gray-700 flex items-center justify-center group-hover:shadow-md transition">
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-gray-700 dark:text-gray-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="6" cy="15" r="4"/><circle cx="18" cy="15" r="4"/><path d="M14 15a2 2 0 0 0-2-2 2 2 0 0 0-2 2"/><path d="M2.5 13 5 7c.7-1.3 1.4-2 3-2"/><path d="M21.5 13 19 7c-.7-1.3-1.5-2-3-2"/></svg>
                    </div>
                    <span class="text-xs text-gray-600 dark:text-gray-400">专业的</span>
                  </button>

                  <button @click="startAIAction('polish', 'concise')" class="flex flex-col items-center gap-2 group cursor-pointer">
                    <div class="w-14 h-14 rounded-full bg-white/80 dark:bg-gray-800/80 shadow-sm border border-gray-100 dark:border-gray-700 flex items-center justify-center group-hover:shadow-md transition">
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-gray-700 dark:text-gray-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><path d="m14 10.5 1.5 1.5-1.5 1.5"/><path d="m14 10.5-1.5 1.5 1.5 1.5"/></svg>
                    </div>
                    <span class="text-xs text-gray-600 dark:text-gray-400">简洁的</span>
                  </button>

                  <button @click="startAIAction('polish', 'friendly')" class="flex flex-col items-center gap-2 group cursor-pointer">
                    <div class="w-14 h-14 rounded-full bg-white/80 dark:bg-gray-800/80 shadow-sm border border-gray-100 dark:border-gray-700 flex items-center justify-center group-hover:shadow-md transition">
                      <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6 text-gray-700 dark:text-gray-300" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M8 14s1.5 2 4 2 4-2 4-2"/><line x1="9" y1="9" x2="9.01" y2="9"/><line x1="15" y1="9" x2="15.01" y2="9"/></svg>
                    </div>
                    <span class="text-xs text-gray-600 dark:text-gray-400">友好的</span>
                  </button>
                </div>
              </div>

            </div>

            <!-- Loading State -->
            <div v-else-if="isPolishing" class="py-12 flex flex-col items-center justify-center gap-6 relative z-10 min-h-[300px]">
              <div class="relative w-16 h-16">
                <div class="absolute inset-0 border-4 border-orange-100 border-t-orange-500 rounded-full animate-spin"></div>
                <div class="absolute inset-4 bg-orange-50 rounded-full animate-pulse flex items-center justify-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-orange-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.937 15.5A2 2 0 0 0 8.5 14.063l-6.135-1.582a.5.5 0 0 1 0-.962L8.5 9.936A2 2 0 0 0 9.937 8.5l1.582-6.135a.5.5 0 0 1 .963 0L14.063 8.5A2 2 0 0 0 15.5 9.937l6.135 1.581a.5.5 0 0 1 0 .964L15.5 14.063a2 2 0 0 0-1.437 1.437l-1.582 6.135a.5.5 0 0 1-.963 0z"/><path d="M20 3v4"/><path d="M22 5h-4"/><path d="M4 17v2"/><path d="M5 18H3"/></svg>
                </div>
              </div>
              <p class="text-gray-600 dark:text-gray-400 text-sm font-medium animate-pulse">AI 正在为您写作，请稍候...</p>
            </div>

            <!-- Result State -->
            <div v-else-if="polishedResult" class="flex flex-col gap-4 relative z-10">
              <div v-if="polishSummary && polishSummary !== '已执行指定创作/修改操作'" class="text-sm bg-blue-50/80 dark:bg-blue-900/20 px-4 py-3 rounded-xl border border-blue-100 dark:border-blue-800/50 text-blue-700 dark:text-blue-300 flex items-start gap-2 backdrop-blur">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 shrink-0 mt-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/></svg>
                <div>
                  <span class="font-bold mr-1">摘要：</span>{{ polishSummary }}
                </div>
              </div>
              
              <div class="relative">
                <div class="absolute right-2 top-2 z-20">
                  <button @click="copyToClipboard" class="p-1.5 bg-gray-100 dark:bg-gray-800 rounded-md text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 shadow-sm cursor-pointer" title="复制结果">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2" ry="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg>
                  </button>
                </div>
                <textarea 
                  v-model="polishedResult" 
                  rows="10" 
                  class="w-full rounded-2xl border border-gray-200 dark:border-gray-700 bg-white/60 dark:bg-gray-900/60 p-4 text-sm text-[var(--text-color)] focus:border-orange-500 focus:ring-1 focus:ring-orange-500 outline-none resize-none backdrop-blur shadow-inner leading-relaxed"
                ></textarea>
              </div>

              <div class="mt-2 flex flex-wrap justify-end gap-3">
                <button @click="resetPolish" class="cursor-pointer px-5 py-2.5 rounded-xl text-sm font-medium bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 transition">重新写作</button>
                <button @click="applyReplace" class="cursor-pointer px-5 py-2.5 rounded-xl text-sm font-medium bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 shadow-md hover:shadow-lg transition">替换原文本</button>
              </div>
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
import { fetchAIWrite, type AIWriteRequest } from '@/service/api/agent'
import { theToast } from '@/utils/toast'

const isOpen = ref(false)
const isPolishing = ref(false)
const polishedResult = ref('')
const polishSummary = ref('')
const originalContent = ref('')
const customPrompt = ref('')

const emit = defineEmits(['apply'])

function open(content: string) {
  originalContent.value = content || ''
  polishedResult.value = ''
  polishSummary.value = ''
  customPrompt.value = ''
  isPolishing.value = false
  isOpen.value = true
}

function close() {
  isOpen.value = false
}

function resetPolish() {
  polishedResult.value = ''
  polishSummary.value = ''
  isPolishing.value = false
}

async function startAIAction(action: 'generate' | 'summarize' | 'correct' | 'expand' | 'polish', prompt: string = '') {
  if (action === 'generate' && !customPrompt.value.trim()) {
    theToast.warning('请先输入生成需求')
    return
  }

  if (action !== 'generate' && (!originalContent.value || originalContent.value.trim() === '')) {
    theToast.warning('编辑器当前没有内容，无法执行此操作')
    return
  }

  isPolishing.value = true
  
  const actualPrompt = action === 'generate' ? customPrompt.value : prompt

  try {
    const res = await fetchAIWrite({
      original_content: originalContent.value,
      action: action,
      prompt: actualPrompt
    })
    
    if (res.code === 1 && res.data) {
      polishedResult.value = res.data.content
      polishSummary.value = res.data.summary
    } else {
      theToast.error(res.msg || '操作失败')
      close()
    }
  } catch (error) {
    theToast.error('网络请求出错')
    close()
  } finally {
    isPolishing.value = false
  }
}

async function copyToClipboard() {
  try {
    await navigator.clipboard.writeText(polishedResult.value)
    theToast.success('已复制到剪贴板')
  } catch (e) {
    theToast.error('复制失败')
  }
}

function applyReplace() {
  emit('apply', polishedResult.value)
  close()
}

export interface TheAIPolishModalType {
  open: (content: string) => void
}

defineExpose({ open })
</script>

<template>
  <MdEditor
    class="h-auto sm:min-h-[8rem] md:min-h-[13rem]"
    v-model="content"
    :id="initEditor.id"
    :theme="initEditor.theme"
    :language="initEditor.language"
    :show-code-row-number="initEditor.showCodeRowNumber"
    :preview-theme="initEditor.previewTheme"
    :code-theme="initEditor.codeTheme"
    :code-style-reverse="initEditor.codeStyleReverse"
    :no-img-zoom-in="initEditor.noImgZoomIn"
    :code-foldable="initEditor.codeFoldable"
    :auto-fold-threshold="initEditor.autoFoldThreshold"
    :toolbars="initEditor.toolbars"
    :no-prettier="initEditor.noPrettier"
    :tab-width="initEditor.tabWidth"
    :placeholder="initEditor.placeholder"
    :preview="initEditor.preview"
    :scroll-auto="initEditor.scrollAuto"
    :show-toolbar-name="initEditor.showToolbarName"
    :footers="initEditor.footers"
    :no-upload-img="initEditor.noUploadImg"
  >
    <template #defToolbars>
      <div class="flex items-center h-full ml-3 pointer-events-none select-none">
        <Transition name="status-fade" mode="out-in">
          <div
            v-if="lastSavedTime"
            :key="isSaving ? 'saving' : 'saved'"
            class="flex items-center gap-1.5"
          >
            <div class="flex items-center justify-center">
              <svg
                v-if="isSaving"
                class="animate-spin w-3 h-3 text-[var(--text-color-400)] opacity-70"
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  class="opacity-20"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-60"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              <svg
                v-else
                class="w-3.5 h-3.5 text-[var(--text-color-300)] opacity-60"
                xmlns="http://www.w3.org/2000/svg"
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
            <span
              class="text-[11px] font-normal text-[var(--text-color-400)] opacity-60 whitespace-nowrap tracking-tight"
            >
              {{ isSaving ? '保存中...' : `上次保存: ${lastSavedTime}` }}
            </span>
          </div>
        </Transition>
      </div>
    </template>
  </MdEditor>
</template>

<script setup lang="ts">
import { reactive, computed } from 'vue'
import { MdEditor, config } from 'md-editor-v3'
import type { ToolbarNames } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { useEditorStore, useThemeStore } from '@/stores'
import { storeToRefs } from 'pinia'

const editorStore = useEditorStore()
const { isSaving, lastSavedTime } = storeToRefs(editorStore)
const themeStore = useThemeStore()

const content = computed<string>({
  get: () => editorStore.echoToAdd.content,
  set: (val: string) => {
    // 💡 用户开始编辑时，解除分享锁定，允许正常保存草稿
    editorStore.confirmShareContent()
    editorStore.echoToAdd.content = val
  },
})

const theme = computed(() => (themeStore.theme === 'light' ? 'light' : 'dark'))

const initEditor = reactive({
  class: 'theMdEditor',
  theme: theme,
  language: 'my-lang',
  id: 'theMdEditor',
  showCodeRowNumber: false,
  previewTheme: 'github',
  codeTheme: 'atom',
  codeStyleReverse: false,
  noImgZoomIn: false,
  codeFoldable: true,
  autoFoldThreshold: 15,
  preview: false,
  toolbars: [
    'bold',
    'italic',
    'strikeThrough',
    'quote',
    '-',
    0,
    '=',
    'previewOnly',
    'pageFullscreen',
  ] as ToolbarNames[],
  noPrettier: false,
  tabWidth: 2,
  placeholder: '一吐为快~',
  scrollAuto: true,
  showToolbarName: false,
  footers: [],
  noUploadImg: true,
})

config({
  editorConfig: {
    languageUserDefined: {
      'my-lang': {
        toolbarTips: {
          bold: '加粗',
          underline: '下划线',
          italic: '斜体',
          strikeThrough: '删除线',
          title: '标题',
          sub: '下标',
          sup: '上标',
          quote: '引用',
          unorderedList: '无序列表',
          orderedList: '有序列表',
          task: '任务列表',
          codeRow: '行内代码',
          code: '块级代码',
          link: '链接',
          image: '图片',
          table: '表格',
          mermaid: 'mermaid图',
          katex: 'katex公式',
          revoke: '后退',
          next: '前进',
          save: '保存',
          prettier: '美化',
          pageFullscreen: '浏览器全屏',
          fullscreen: '屏幕全屏',
          preview: '预览',
          htmlPreview: 'html代码预览',
          catalog: '目录',
          github: '源码地址',
        },
        titleItem: {
          h1: '一级标题',
          h2: '二级标题',
          h3: '三级标题',
          h4: '四级标题',
          h5: '五级标题',
          h6: '六级标题',
        },
        imgTitleItem: {
          link: '添加链接',
          upload: '上传图片',
          clip2upload: '裁剪上传',
        },
        linkModalTips: {
          linkTitle: '添加链接',
          imageTitle: '添加图片',
          descLabel: '链接描述：',
          descLabelPlaceHolder: '请输入描述...',
          urlLabel: '链接地址：',
          urlLabelPlaceHolder: '请输入链接...',
          buttonOK: '确定',
        },
        clipModalTips: {
          title: '裁剪图片上传',
          buttonUpload: '上传',
        },
        copyCode: {
          text: '复制代码',
          successTips: '已复制！',
          failTips: '复制失败！',
        },
        mermaid: {
          flow: '流程图',
          sequence: '时序图',
          gantt: '甘特图',
          class: '类图',
          state: '状态图',
          pie: '饼图',
          relationship: '关系图',
          journey: '旅程图',
        },
        katex: {
          inline: '行内公式',
          block: '块级公式',
        },
        footer: {
          markdownTotal: '字数',
          scrollAuto: '同步滚动',
        },
      },
    },
  },
})
</script>

<style scoped lang="css">
#theMdEditor {
  height: 10rem;
}

:deep(.md-editor-custom-scrollbar__track) {
  overflow: auto !important; /* 保持可滚动 */
  scrollbar-width: none !important; /* Firefox */
  -ms-overflow-style: none !important; /* IE/Edge */
  display: none !important;
}

:deep(.md-editor-custom-scrollbar__track::-webkit-scrollbar) {
  display: none !important; /* Chrome/Safari */
}

:deep(.md-editor-toolbar-item) {
  color: #c5c5c5;
}

:deep(.md-editor-toolbar-wrapper .md-editor-toolbar-active) {
  color: #ff9f46d4;
  background-color: #f6d6bf59;
}

:deep(ul li) {
  list-style-type: disc;
}
:deep(ul li li) {
  list-style-type: circle;
}
:deep(ul li li li) {
  list-style-type: square;
}
:deep(ol li) {
  list-style-type: decimal;
}

/* 状态切换动画 */
.status-fade-enter-active,
.status-fade-leave-active {
  transition: all 0.3s ease;
}

.status-fade-enter-from,
.status-fade-leave-to {
  opacity: 0;
  transform: translateY(2px);
}
</style>

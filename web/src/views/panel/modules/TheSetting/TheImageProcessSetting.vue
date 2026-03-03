<template>
  <div>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--text-color-600)] font-bold text-lg">图片处理</h1>
        <div class="flex flex-row items-center justify-end gap-2 w-14">
          <button v-if="editMode" @click="handleSave" title="保存">
            <Saveupdate class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
          <button @click="editMode = !editMode" title="编辑">
            <Edit
              v-if="!editMode"
              class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6"
            />
            <Close v-else class="w-5 h-5 text-[var(--text-color-400)] hover:w-6 hover:h-6" />
          </button>
        </div>
      </div>

      <div class="flex flex-col w-full">
        <!-- 本地图片处理 -->
        <div class="flex flex-row flex-wrap items-center justify-start text-[var(--text-color-next-500)] gap-x-2 gap-y-1 min-h-10">
          <h2 class="font-semibold w-30 shrink-0">本地图片:</h2>
          <div class="flex items-center gap-2">
            <BaseSelect
              v-model="localSetting.local_process"
              :options="localProcessOptions"
              :disabled="!editMode"
              class="w-fit h-8"
              @change="handleLocalProcessChange"
            />
            <template v-if="localSetting.local_process === 'local'">
              <button @click="showLocalDocModal = true" title="查看本地图片处理参数文档" class="text-xs text-[var(--text-color-400)] hover:text-[var(--primary-color)] transition-colors flex items-center gap-1 cursor-pointer whitespace-nowrap">
                <Info class="w-3.5 h-3.5" /> <span class="hover:underline">官方指引</span>
              </button>
            </template>
            <a v-else-if="localDocLink" :href="localDocLink" target="_blank" title="查看官方图片处理参数文档" class="text-xs text-[var(--text-color-400)] hover:text-[var(--primary-color)] transition-colors flex items-center gap-1 whitespace-nowrap">
              <Info class="w-3.5 h-3.5" /> <span class="hover:underline">官方指引</span>
            </a>
          </div>
        </div>

        <template v-if="localSetting.local_process">
          <div class="flex flex-row items-start justify-start text-[var(--text-color-next-500)] gap-2 min-h-10 py-1.5">
            <h2 class="font-semibold w-30 shrink-0 pl-4 mt-1.5">缩略图参数:</h2>
            <div class="flex flex-col flex-1 min-w-0 justify-center">
              <div class="flex items-center w-full min-w-0">
                <span
                  v-if="!editMode"
                  class="truncate flex-1 inline-block align-middle leading-8"
                  :title="localSetting.local_thumb_param"
                >
                  {{ localSetting.local_thumb_param || '无 (输出原图)' }}
                </span>
                <template v-else>
                  <div class="flex items-center flex-1 min-w-0">
                    <span v-if="localPrefix" class="mr-1 text-sm text-[var(--text-color-500)] font-mono shrink-0 select-none">{{ localPrefix }}</span>
                    <BaseInput v-model="localThumbModel" placeholder="如 w=800&q=80 (不含前缀)" class="flex-1 py-1! min-w-0" />
                  </div>
                  <button @click="localSetting.local_thumb_param = ''" class="shrink-0 p-1 ml-2 text-[var(--text-color-400)] hover:text-[var(--primary-color)] transition-colors cursor-pointer" title="清空参数 (输出原图)">
                    <ImageIcon class="w-5 h-5" />
                  </button>
                </template>
              </div>
              <span class="text-[11px] text-[var(--text-color-400)] mt-0.5 truncate select-none leading-none" :title="localThumbHint">{{ localThumbHint }}</span>
            </div>
          </div>
          <div class="flex flex-row items-start justify-start text-[var(--text-color-next-500)] gap-2 min-h-10 py-1.5">
            <h2 class="font-semibold w-30 shrink-0 pl-4 mt-1.5">大图参数:</h2>
            <div class="flex flex-col flex-1 min-w-0 justify-center">
              <div class="flex items-center w-full min-w-0">
                <span
                  v-if="!editMode"
                  class="truncate flex-1 inline-block align-middle leading-8"
                  :title="localSetting.local_full_param"
                >
                  {{ localSetting.local_full_param || '无 (输出原图)' }}
                </span>
                <template v-else>
                  <div class="flex items-center flex-1 min-w-0">
                    <span v-if="localPrefix" class="mr-1 text-sm text-[var(--text-color-500)] font-mono shrink-0 select-none">{{ localPrefix }}</span>
                    <BaseInput v-model="localFullModel" placeholder="原图可留空，或填参 (不含前缀)" class="flex-1 py-1! min-w-0" />
                  </div>
                  <button @click="localSetting.local_full_param = ''" class="shrink-0 p-1 ml-2 text-[var(--text-color-400)] hover:text-[var(--primary-color)] transition-colors cursor-pointer" title="清空参数 (输出原图)">
                    <ImageIcon class="w-5 h-5" />
                  </button>
                </template>
              </div>
              <span class="text-[11px] text-[var(--text-color-400)] mt-0.5 truncate select-none leading-none" :title="localFullHint">{{ localFullHint }}</span>
            </div>
          </div>
        </template>

        <!-- S3 图片处理 -->
        <div class="flex flex-row flex-wrap items-center justify-start text-[var(--text-color-next-500)] gap-x-2 gap-y-1 min-h-10">
          <h2 class="font-semibold w-30 shrink-0">S3 图片:</h2>
          <template v-if="s3Available">
            <div class="flex items-center gap-2">
              <BaseSelect
                v-model="localSetting.s3_process"
                :options="s3ProcessOptions"
                :disabled="!editMode"
                class="w-fit h-8"
                @change="handleS3ProcessChange"
              />
              <a v-if="localSetting.s3_process && s3DocLink" :href="s3DocLink" target="_blank" title="查看官方图片处理参数文档" class="text-xs text-[var(--text-color-400)] hover:text-[var(--primary-color)] transition-colors flex items-center gap-1 whitespace-nowrap">
                <Info class="w-3.5 h-3.5" /> <span class="hover:underline">官方指引</span>
              </a>
            </div>
          </template>
          <span v-else class="text-[var(--text-color-400)] text-xs">
            需先在存储设置中启用S3
          </span>
        </div>

        <template v-if="s3Available && localSetting.s3_process">
          <div class="flex flex-row items-start justify-start text-[var(--text-color-next-500)] gap-2 min-h-10 py-1.5">
            <h2 class="font-semibold w-30 shrink-0 pl-4 mt-1.5">缩略图参数:</h2>
            <div class="flex flex-col flex-1 min-w-0 justify-center">
              <div class="flex items-center w-full min-w-0">
                <span
                  v-if="!editMode"
                  class="truncate flex-1 inline-block align-middle leading-8"
                  :title="localSetting.s3_thumb_param"
                >
                  {{ localSetting.s3_thumb_param || '无 (输出原图)' }}
                </span>
                <template v-else>
                  <div class="flex items-center flex-1 min-w-0">
                    <span v-if="s3Prefix" class="mr-1 text-sm text-[var(--text-color-500)] font-mono shrink-0 select-none">{{ s3Prefix }}</span>
                    <BaseInput v-model="s3ThumbModel" placeholder="如 thumbnail/800x (不含前缀)" class="flex-1 py-1! min-w-0" />
                  </div>
                  <button @click="localSetting.s3_thumb_param = ''" class="shrink-0 p-1 ml-2 text-[var(--text-color-400)] hover:text-[var(--primary-color)] transition-colors cursor-pointer" title="清空参数 (输出原图)">
                    <ImageIcon class="w-5 h-5" />
                  </button>
                </template>
              </div>
              <span class="text-[11px] text-[var(--text-color-400)] mt-0.5 truncate select-none leading-none" :title="s3ThumbHint">{{ s3ThumbHint }}</span>
            </div>
          </div>
          <div class="flex flex-row items-start justify-start text-[var(--text-color-next-500)] gap-2 min-h-10 py-1.5">
            <h2 class="font-semibold w-30 shrink-0 pl-4 mt-1.5">大图参数:</h2>
            <div class="flex flex-col flex-1 min-w-0 justify-center">
              <div class="flex items-center w-full min-w-0">
                <span
                  v-if="!editMode"
                  class="truncate flex-1 inline-block align-middle leading-8"
                  :title="localSetting.s3_full_param"
                >
                  {{ localSetting.s3_full_param || '无 (输出原图)' }}
                </span>
                <template v-else>
                  <div class="flex items-center flex-1 min-w-0">
                    <span v-if="s3Prefix" class="mr-1 text-sm text-[var(--text-color-500)] font-mono shrink-0 select-none">{{ s3Prefix }}</span>
                    <BaseInput v-model="s3FullModel" placeholder="原图可留空，或填自定义参数 (不含前缀)" class="flex-1 py-1! min-w-0" />
                  </div>
                  <button @click="localSetting.s3_full_param = ''" class="shrink-0 p-1 ml-2 text-[var(--text-color-400)] hover:text-[var(--primary-color)] transition-colors cursor-pointer" title="清空参数 (输出原图)">
                    <ImageIcon class="w-5 h-5" />
                  </button>
                </template>
              </div>
              <span class="text-[11px] text-[var(--text-color-400)] mt-0.5 truncate select-none leading-none" :title="s3FullHint">{{ s3FullHint }}</span>
            </div>
          </div>
        </template>
      </div>
    </div>
  </PanelCard>

  <!-- 本地处理官方指引弹窗 -->
  <BaseModal v-model="showLocalDocModal">
    <div class="flex flex-col">
      <div class="flex items-center justify-between mb-4 pb-3 border-b border-gray-100">
        <h3 class="text-lg font-bold text-[var(--text-color-700)]">内置本地图片处理指引</h3>
        <button @click="showLocalDocModal = false" class="text-[var(--text-color-400)] hover:text-red-500 transition-colors">
          <Close class="w-5 h-5" />
        </button>
      </div>
      <div class="text-sm text-[var(--text-color-600)] space-y-4 max-h-[70vh] overflow-y-auto pr-2 custom-scrollbar">
        <p>系统内置了基于 <strong>Golang Native (x/image)</strong> 的高性能轻量级图片处理引擎。无需第三方云服务商即可实现基础的图片格式转换与缩放。参数可直接拼接于 URL 末端。</p>
        
        <div>
          <h4 class="font-bold text-[var(--text-color-700)] mb-1">使用示例</h4>
          <code class="block bg-[var(--bg-color)] px-3 py-2 rounded-md text-orange-600 font-mono text-xs overflow-x-auto whitespace-nowrap border border-gray-100/50 shadow-inner">
            ?w=800&h=600&q=85&fmt=webp&mode=lfit
          </code>
        </div>
        
        <div>
          <h4 class="font-bold text-[var(--text-color-700)] mb-2">核心参数</h4>
          <ul class="list-disc list-inside space-y-1 ml-1 marker:text-orange-400">
            <li><code class="text-orange-600 bg-orange-50 px-1 rounded mx-0.5">w</code> : 目标宽度 (正整数, px)</li>
            <li><code class="text-orange-600 bg-orange-50 px-1 rounded mx-0.5">h</code> : 目标高度 (正整数, px)</li>
            <li><code class="text-orange-600 bg-orange-50 px-1 rounded mx-0.5">q</code> : 输出画质 (1-100)，默认 75。仅作用于 JPEG/WEBP 这种有损压缩格式。</li>
            <li>
              <code class="text-orange-600 bg-orange-50 px-1 rounded mx-0.5">fmt</code> : 强制输出格式。支持 <code class="font-mono text-xs">jpg, png, webp, gif</code> 等常见格式转换。留空保持原格式，强烈建议指定为 <code class="font-mono text-xs">webp</code> 极大缩减带宽！
            </li>
          </ul>
        </div>
        
        <div>
          <h4 class="font-bold text-[var(--text-color-700)] mb-2">缩放模式 (mode)</h4>
          <div class="space-y-3 pl-2">
            <div class="bg-[var(--bg-color)] p-2.5 rounded border border-gray-100/50 shadow-xs">
              <span class="font-mono font-bold text-orange-600 text-sm">lfit</span><span class="text-xs text-[var(--text-color-400)] ml-2 border border-gray-200 rounded px-1">默认推荐</span>
              <p class="mt-1 text-xs"><strong>等比缩小不拉伸：</strong>图片会自动等比缩小，直到长和宽都能完全塞进你指定的 w/h 矩形框内。如果原图比 w/h 小，系统将<strong>不会</strong>粗暴放大原图（避免模糊）。</p>
            </div>
            <div class="bg-[var(--bg-color)] p-2.5 rounded border border-gray-100/50 shadow-xs">
              <span class="font-mono font-bold text-orange-600 text-sm">mfit</span>
              <p class="mt-1 text-xs"><strong>等比短边缩放：</strong>也是等比缩放并保持原图完整，但与 lfit 不同，它是以<strong>短边</strong>能匹配到目标尺寸为准。注：为保证画质，系统目前针对它做了小图保护，<strong>即使原图过小也不会强行拉伸放大</strong>。</p>
            </div>
            <div class="bg-[var(--bg-color)] p-2.5 rounded border border-gray-100/50 shadow-xs">
              <span class="font-mono font-bold text-orange-600 text-sm">fill</span>
              <p class="mt-1 text-xs"><strong>居中裁剪填充：</strong>系统先将图片等比例放大或缩小，直到填满整个 w/h 矩形，然后把超出边界的部分(通常是上下或左右多余的区域)<strong>居中裁剪</strong>掉。最适合生成尺寸严格一致的列表九宫格缩略图，杜绝留白。</p>
            </div>
          </div>
        </div>
      </div>
      <div class="mt-6 text-right">
        <button @click="showLocalDocModal = false" class="px-5 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg text-sm font-medium transition-colors cursor-pointer">
          关闭了解
        </button>
      </div>
    </div>
  </BaseModal>
  </div>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import Edit from '@/components/icons/edit.vue'
import Close from '@/components/icons/close.vue'
import Info from '@/components/icons/info.vue'
import ImageIcon from '@/components/icons/image.vue'
import Saveupdate from '@/components/icons/saveupdate.vue'
import { ref, computed, onMounted } from 'vue'
import { S3Provider } from '@/enums/enums'
import { fetchUpdateImageProcessSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { getImageProcessSetting } = settingStore
const { S3Setting, ImageProcessSetting } = storeToRefs(settingStore)

const editMode = ref(false)
const showLocalDocModal = ref(false)

// 本地编辑副本
const localSetting = ref<App.Api.Setting.ImageProcessSetting>({
  local_process: '',
  local_thumb_param: '',
  local_full_param: '',
  s3_process: '',
  s3_thumb_param: '',
  s3_full_param: '',
})

// 同步 store 到本地副本
const syncFromStore = () => {
  localSetting.value = { ...ImageProcessSetting.value }
}

// 本地图片处理选项
const localProcessOptions = ref([
  { label: '禁用', value: '' },
  { label: '内置本地处理', value: 'local' },
  { label: '阿里云 ESA', value: 'aliyun_esa' },
  { label: '腾讯云 EO', value: 'tencent_eo' },
  { label: '自定义', value: 'custom' },
])

const handleLocalProcessChange = (val: string | number | boolean | null | undefined) => {
  if (val === 'local') {
    localSetting.value.local_thumb_param = '?w=800&q=75&mode=lfit&fmt=webp'
    localSetting.value.local_full_param = '?fmt=webp'
  } else if (val === 'aliyun_esa') {
    localSetting.value.local_thumb_param = '?image_process=resize,w_800,limit_1/quality,q_75/format,webp'
    localSetting.value.local_full_param = '?image_process=format,webp'
  } else if (val === 'tencent_eo') {
    localSetting.value.local_thumb_param = '?eo-img.resize=w/800&eo-img.format=webp'
    localSetting.value.local_full_param = '?eo-img.format=webp'
  }
}

const localDocLink = computed(() => {
  if (localSetting.value.local_process === 'aliyun_esa') return 'https://help.aliyun.com/zh/edge-security-acceleration/esa/user-guide/image-optimization#5eaee9f90f8o2'
  if (localSetting.value.local_process === 'tencent_eo') return 'https://cloud.tencent.com/document/product/1552/84731#87fb27be-cf15-43b6-8117-6e724526a340'
  return ''
})

const s3DocLink = computed(() => {
  if (localSetting.value.s3_process === S3Provider.ALIYUN) return 'https://help.aliyun.com/zh/oss/user-guide/overview-17/'
  if (localSetting.value.s3_process === S3Provider.TENCENT) return 'https://cloud.tencent.com/document/product/436/42215'
  return ''
})

// 提取前缀
const localPrefix = computed(() => {
  if (localSetting.value.local_process === 'aliyun_esa') return '?image_process='
  if (localSetting.value.local_process === 'tencent_eo') return '?eo-img.'
  if (localSetting.value.local_process === 'local') return '?'
  return ''
})

const s3Prefix = computed(() => {
  if (localSetting.value.s3_process === S3Provider.ALIYUN) return '?x-oss-process=image/'
  if (localSetting.value.s3_process === S3Provider.TENCENT) return '?imageMogr2/'
  return ''
})

// 参数双向绑定简写工厂
const createParamModel = (key: 'local_thumb_param' | 'local_full_param' | 's3_thumb_param' | 's3_full_param', prefixRef: any) => computed({
  get() {
    const p = prefixRef.value
    const val = localSetting.value[key] || ''
    return (p && val.startsWith(p)) ? val.slice(p.length) : val
  },
  set(val: string) {
    localSetting.value[key] = val ? prefixRef.value + val : ''
  }
})

const localThumbModel = createParamModel('local_thumb_param', localPrefix)
const localFullModel = createParamModel('local_full_param', localPrefix)
const s3ThumbModel = createParamModel('s3_thumb_param', s3Prefix)
const s3FullModel = createParamModel('s3_full_param', s3Prefix)

// 动态参数实时解析器 (基于阿里云 OSS/ESA, 腾讯云 COS/EO 官方文档全量覆盖)
const analyzeParam = (param: string) => {
  if (!param) return '直接输出原图'
  const parts: string[] = []

  // 1. 宽高解析 (含比例缩放 px, % 等)
  const scaleMatch = param.match(/(?:thumbnail\/!?|resize,p_)(\d{1,3})p/i)
  if (scaleMatch?.[1]) {
    parts.push(`缩放 ${scaleMatch[1]}%`)
  } else {
    // 基础宽高 (px)
    const wMatch = param.match(/(?:[?&,.]w=|w_|w\/|thumbnail\/(?!x)(?:!)?)(\d+)/i)
    if (wMatch?.[1]) parts.push(`宽 ${wMatch[1]}`)
    const hMatch = param.match(/(?:[?&,.]h=|h_|h\/|thumbnail\/(?:!|\d+)?x)(\d+)/i)
    if (hMatch?.[1]) parts.push(`高 ${hMatch[1]}`)
  }

  // 2. 缩放与裁剪模式
  if (/(?:lfit|m_lfit|type\/lfit|limit)/i.test(param)) parts.push('等比缩小')
  else if (/(?:mfit|m_mfit|type\/mfit)/i.test(param)) parts.push('等比放大')
  else if (/(?:fill|m_fill|type\/fill|crop|circle|indexcrop)/i.test(param)) parts.push('裁剪填充')
  else if (/(?:pad|m_pad|type\/pad)/i.test(param)) parts.push('等比留白')
  else if (/(?:fixed|m_fixed|force|type\/fixed|!\/?$)/i.test(param)) parts.push('强制拉伸')
  else if (/>\/?$/.test(param)) parts.push('仅缩小不放大')
  else if (/<\/?$/.test(param)) parts.push('仅放大不缩小')
  else if (/r\/?$/.test(param)) parts.push('自动按短边裁剪')
  else if (parts.some(p => p.startsWith('宽') || p.startsWith('高'))) parts.push('常规缩放')

  // 3. 画质解析
  const qMatch = param.match(/(?:[?&,.]q=|q_|q\/|quality[=,/](?:Q_|q_|!?)?|rquality\/)(\d+)/i)
  if (qMatch?.[1]) {
    const isAbs = /(?:Q_|!)/.test(param) ? '绝对' : (param.includes('rquality') ? '相对' : '')
    parts.push(`${isAbs}画质 ${qMatch[1]}%`)
  }

  // 4. 格式解析 (全面扩充)
  const fmtMatch = param.match(/(?:fmt[=_/]|format[=,/])(webp|jpg|jpeg|png|gif|avif|heic|heif|bmp|tpg|tiff)/i)
  if (fmtMatch?.[1]) {
    parts.push(`格式 ${fmtMatch[1].toUpperCase()}`)
  } else if (/webp/i.test(param)) {
    parts.push('转 WebP')
  }

  // 5. 特殊功能与图像增强
  const special: string[] = []
  if (/interlace[=,/]1|渐进/i.test(param)) special.push('交错渐进显示')
  if (/ignore-error[=,/]1/i.test(param)) special.push('防报错')
  if (/strip/i.test(param)) special.push('去除 EXIF 数据')
  if (/sharpen/i.test(param)) special.push('锐化')
  if (/blur|gaussian/i.test(param)) special.push('高斯模糊')
  if (/watermark/i.test(param)) special.push('水印')
  if (/auto-orient|orient[=,/]1/i.test(param)) special.push('自适应旋转')
  if (/rotate/i.test(param)) {
    const rMatch = param.match(/rotate[=,/](\d+)/i)
    special.push(rMatch?.[1] ? `旋转 ${rMatch[1]}°` : '旋转')
  }
  if (/bright/i.test(param)) special.push('亮度调整')
  if (/contrast/i.test(param)) special.push('对比度调整')

  if (special.length > 0) {
    parts.push(special.join('/'))
  }

  return parts.length > 0 ? parts.join('，') : '自定义预设或原图'
}

const localThumbHint = computed(() => analyzeParam(localSetting.value.local_thumb_param))
const localFullHint = computed(() => analyzeParam(localSetting.value.local_full_param))
const s3ThumbHint = computed(() => analyzeParam(localSetting.value.s3_thumb_param))
const s3FullHint = computed(() => analyzeParam(localSetting.value.s3_full_param))

// S3 是否可以启用图片处理
const s3Available = computed(() => {
  return S3Setting.value.enable
})

// S3 图片处理选项（动态基于选择的 S3 provider）
const s3ProcessOptions = computed(() => {
  const p = S3Setting.value.provider
  if (p === S3Provider.ALIYUN) {
    return [
      { label: '禁用', value: '' },
      { label: '阿里云 OSS', value: S3Provider.ALIYUN },
      { label: '自定义', value: 'custom' },
    ]
  } else if (p === S3Provider.TENCENT) {
    return [
      { label: '禁用', value: '' },
      { label: '腾讯云 COS', value: S3Provider.TENCENT },
      { label: '自定义', value: 'custom' },
    ]
  }
  return [
    { label: '禁用', value: '' },
    { label: '自定义', value: 'custom' }
  ]
})

const handleS3ProcessChange = (val: string | number | boolean | null | undefined) => {
  if (val === S3Provider.ALIYUN) {
    localSetting.value.s3_thumb_param = '?x-oss-process=image/resize,m_lfit,w_800,limit_1/quality,q_75/interlace,1/format,webp'
    localSetting.value.s3_full_param = '?x-oss-process=image/format,webp'
  } else if (val === S3Provider.TENCENT) {
    localSetting.value.s3_thumb_param = '?imageMogr2/thumbnail/800x>/quality/75/ignore-error/1/interlace/1/format/webp'
    localSetting.value.s3_full_param = '?imageMogr2/format/webp'
  }
}

const handleSave = async () => {
  const res = await fetchUpdateImageProcessSettings(localSetting.value)
  if (res.code === 1) {
    theToast.success(res.msg)
  }
  editMode.value = false
  getImageProcessSetting()
}

onMounted(() => {
  getImageProcessSetting().then(() => syncFromStore())
})
</script>

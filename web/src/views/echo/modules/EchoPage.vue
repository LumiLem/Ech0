<template>
  <div class="px-3 pb-4 py-2 mt-4 sm:mt-6 mb-10 mx-auto flex justify-center items-center">
    <div class="w-full sm:max-w-lg mx-auto">
      <div class="mx-auto max-w-sm">
        <!-- 返回上一页 -->
        <BaseButton
          @click="goBack"
          class="text-[var(--text-color-600)] rounded-md shadow-none! border-none! ring-0! bg-transparent! group"
          title="返回首页"
        >
          <Arrow
            class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
          />
        </BaseButton>
      </div>

      <div v-if="echo" class="w-full sm:mt-1 mx-auto">
        <TheEchoDetail
          :echo="echo"
          @update-like-count="handleUpdateLikeCount"
          @print-echo="handlePrintEcho"
        />
        <TheComment class="my-2" />
      </div>
      <div v-else class="w-full sm:mt-1 text-[var(--text-color-300)]">
        <p class="text-center">正在加载 Echo 详情...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useHead } from '@unhead/vue'
import { fetchGetEchoById } from '@/service/api'
import TheEchoDetail from '@/components/advanced/TheEchoDetail.vue'
import TheComment from '@/components/advanced/TheComment.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
import { useEchoStore, useSettingStore, useInboxStore, useZoneStore } from '@/stores'
import { getApiUrl } from '@/service/request/shared'
import { storeToRefs } from 'pinia'

const router = useRouter()
const route = useRoute()
const echoId = route.params.echoId as string

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const inboxStore = useInboxStore()
const { inboxMode } = storeToRefs(inboxStore)
const zoneStore = useZoneStore()
import { ensureAbsoluteUrl } from '@/utils/other'

const isLoading = ref(true)
const echo = ref<App.Api.Ech0.Echo | null>(null)

// 动态 Meta 管理
useHead({
  title: computed(() => {
    if (!echo.value) return settingStore.SystemSetting.site_title
    const d = new Date(echo.value.created_at)
    const dateStr = `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
    return `${echo.value.username}发表于${dateStr}的动态 - ${settingStore.SystemSetting.site_title}`
  }),
  meta: [
    {
      name: 'description',
      content: computed(() => {
        if (!echo.value) return settingStore.SystemSetting.site_description
        let desc = echo.value.content?.replace(/<[^>]*>/g, '').replace(/[*#_~`]/g, '').slice(0, 200)
        if (!desc) {
          const mediaCount = echo.value.media?.length || 0
          desc = mediaCount > 0 
            ? `${echo.value.username}分享了${mediaCount}个媒体文件`
            : `这是来自${echo.value.username}的一条动态，点击查看详情。`
        }
        return desc
      })
    },
    {
      name: 'author',
      content: computed(() => echo.value?.username || settingStore.SystemSetting.server_name)
    },
    {
      property: 'og:title',
      content: computed(() => {
        if (!echo.value) return settingStore.SystemSetting.site_title
        const d = new Date(echo.value.created_at)
        const dateStr = `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`
        return `${echo.value.username}发表于${dateStr}的动态 - ${settingStore.SystemSetting.site_title}`
      })
    },
    {
      property: 'og:image',
      content: computed(() => {
        if (echo.value?.media && echo.value.media.length > 0) {
          // 💡 只寻找图片类型，视频/音频无法被识别为预览图
          const firstImage = echo.value.media.find(m => m.media_type === 'image')
          if (firstImage?.media_url) {
            return ensureAbsoluteUrl(firstImage.media_url, settingStore.SystemSetting.server_url)
          }
        }
        // 如果动态没图（或只有视频），回退到站点 Logo（绝对路径）
        const logo = settingStore.SystemSetting.server_logo
        return ensureAbsoluteUrl(logo || '/Ech0.png', settingStore.SystemSetting.server_url)
      })
    },
    {
      property: 'og:url',
      content: computed(() => window.location.href)
    }
  ],
  script: [
    {
      id: 'ldjson-schema',
      key: 'ldjson-schema',
      type: 'application/ld+json',
      innerHTML: computed(() => {
        if (!echo.value) return ''
        const date = new Date(echo.value.created_at).toISOString()
        const logo = settingStore.SystemSetting.server_logo
        const siteLogo = ensureAbsoluteUrl(logo || '/Ech0.png', settingStore.SystemSetting.server_url)
            
        // 提前安全处理 Media URL (严格限制为图片类型)
        const previewImageUrl = echo.value.media?.find(m => m.media_type === 'image')?.media_url 
        
        const echoImage = previewImageUrl 
          ? ensureAbsoluteUrl(previewImageUrl, settingStore.SystemSetting.server_url)
          : siteLogo

        const data = {
          "@context": "https://schema.org",
          "@type": "BlogPosting",
          "headline": `${echo.value.username}发表于${date.split('T')[0]}的动态`,
          "description": echo.value.content?.slice(0, 200),
          "image": echoImage,
          "datePublished": date,
          "author": {
            "@type": "Person",
            "name": echo.value.username
          },
          "publisher": {
            "@type": "Organization",
            "name": settingStore.SystemSetting.server_name || settingStore.SystemSetting.site_title,
            "logo": {
              "@type": "ImageObject",
              "url": siteLogo
            }
          }
        }
        return JSON.stringify(data)
      })
    }
  ]
})

// 从 echoIndexMap 获取对应的 EchoList索引
const getEchoFromStore = (): App.Api.Ech0.Echo | null => {
  const idx = echoStore.echoIndexMap.get(Number(echoId))
  if (idx !== undefined) {
    return echoStore.echoList[idx] ?? null
  }
  return null
}

// 刷新点赞数据
const handleUpdateLikeCount = () => {
  if (echo.value) {
    // 更新 Echo 的点赞数量
    echo.value.fav_count += 1
  }
}

const handlePrintEcho = (targetEcho: App.Api.Ech0.Echo) => {
  const text = targetEcho.content?.trim() || ''
  if (!text) return

  zoneStore.setPendingPrintEcho(targetEcho)

  router.push({
    name: 'zone',
    params: {
      echoId: String(targetEcho.id),
    },
  })
}

const goBack = () => {
  if (window.history.length > 2) {
    window.history.back()
  } else {
    router.push({ name: 'home' }) // 没有历史记录则跳首页
  }
}
onMounted(async () => {
  // 先尝试从 store 获取
  echo.value = getEchoFromStore()

  // 如果 store 里没有，再发请求兜底
  if (!echo.value) {
    const res = await fetchGetEchoById(echoId)
    if (res.code === 1) {
      echo.value = res.data
    }
  }
  isLoading.value = false
})
</script>

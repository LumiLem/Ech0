<template>
  <!-- Uppy Dashboard 容器 -->
  <div
    id="uppy-dashboard"
    class="rounded-md overflow-hidden shadow-inner ring-inset ring-1 ring-[var(--ring-color)]"
  ></div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { getAuthToken } from '@/service/request/shared'
import { useUserStore } from '@/stores/user'
import { theToast } from '@/utils/toast'
import { storeToRefs } from 'pinia'
import { ImageSource } from '@/enums/enums'
import { fetchGetPresignedUrl } from '@/service/api'
import { useEditorStore } from '@/stores/editor'
import { 
  detectLivePhotoPairs, 
  detectLivePhotoPairsFromMedia, 
  applyLivePhotoPairIds, 
  getBaseName,
  detectEmbeddedMotionPhoto,
  separateEmbeddedMotionPhoto
} from '@/utils/livephoto'
/* --------------- 与Uppy相关 ---------------- */
import Uppy from '@uppy/core'
import Dashboard from '@uppy/dashboard'
import XHRUpload from '@uppy/xhr-upload'
import AwsS3 from '@uppy/aws-s3'
import '@uppy/core/css/style.min.css'
import '@uppy/dashboard/css/style.min.css'
import zh_CN from '@uppy/locales/lib/zh_CN'
/* --------------- HEIC 转换 ---------------- */
import heic2any from 'heic2any'

let uppy: Uppy | null = null

const props = defineProps<{
  TheImageSource: string
}>()
// const emit = defineEmits(['uppyUploaded'])

const memorySource = ref<string>(props.TheImageSource) // 用于记住上传方式
const isUploading = ref<boolean>(false) // 是否正在上传
const files = ref<App.Api.Ech0.MediaToAdd[]>([]) // 已上传的文件列表
const tempFiles = ref<Map<string, { url: string; objectKey: string }>>(new Map()) // 用于S3临时存储文件回显地址的 Map(key: fileName, value: {url, objectKey})
// 用于保存原始文件名到 live_pair_id 的映射（在上传前检测，上传后应用）
const originalFilenameToPairId = ref<Map<string, string>>(new Map())
// 用于保存上传文件的原始文件名（key: uppy file id, value: original filename）
const uppyFileIdToOriginalName = ref<Map<string, string>>(new Map())
// 标记是否正在处理嵌入式实况照片分离（防止重复触发）
const isProcessingEmbedded = ref<boolean>(false)

const userStore = useUserStore()
const editorStore = useEditorStore()
const { isLogin } = storeToRefs(userStore)
const envURL = import.meta.env.VITE_SERVICE_BASE_URL as string
const backendURL = envURL.endsWith('/') ? envURL.slice(0, -1) : envURL

// ✨ 监听粘贴事件
const handlePaste = async (e: ClipboardEvent) => {
  if (!e.clipboardData) return

  for (const item of e.clipboardData.items) {
    if (item.type.startsWith('image/')) {
      const file = item.getAsFile()
      if (file) {
        const uniqueFile = new File([file], file.name, {
          type: file.type,
          lastModified: Date.now(),
        })

        uppy?.addFile({
          id: `pasted-image-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
          name: uniqueFile.name,
          type: uniqueFile.type,
          data: uniqueFile,
          source: 'PastedImage',
        })
        // 粘贴的图片会触发 files-added 事件，在那里统一处理上传
      }
    }
  }
}

// HEIC 转换函数
const convertHeicToJpeg = async (file: any): Promise<void> => {
  const isHeic = file.type === 'image/heic' || 
                 file.type === 'image/heif' ||
                 file.name.toLowerCase().endsWith('.heic') ||
                 file.name.toLowerCase().endsWith('.heif')
  
  if (!isHeic) return

  try {
    console.log('检测到 HEIC 文件，开始转换:', file.name, 'type:', file.type)
    theToast.info('正在转换 HEIC 格式...', { duration: 1500 })
    
    const convertedBlob = await heic2any({
      blob: file.data as Blob,
      toType: 'image/jpeg',
      quality: 0.9
    })
    
    // 处理可能返回的数组或单个 Blob
    const jpegBlob = Array.isArray(convertedBlob) ? convertedBlob[0] : convertedBlob
    
    if (!jpegBlob) {
      throw new Error('转换后的 Blob 为空')
    }
    
    // 更新文件数据和类型，保持原文件名不变
    uppy?.setFileState(file.id, {
      data: jpegBlob,
      type: 'image/jpeg'
    })
    
    console.log('HEIC 转换成功，文件类型已更新为 image/jpeg')
    theToast.success('HEIC 转换完成！', { duration: 1000 })
  } catch (error) {
    console.error('HEIC 转换失败:', error)
    theToast.warning('HEIC 转换失败，将上传原文件（部分浏览器可能无法查看）', { duration: 3000 })
    // 转换失败不影响上传，继续使用原文件
  }
}

// 嵌入式实况照片分离函数
const separateEmbeddedMotionPhotoFile = async (file: any): Promise<boolean> => {
  try {
    console.log('🔍 检测嵌入式实况照片:', file.name, (file.size / (1024 * 1024)).toFixed(2) + 'MB')
    
    // 将 Uppy 文件转换为标准 File 对象
    const standardFile = new File([file.data as Blob], file.name, {
      type: file.type,
      lastModified: Date.now(),
    })

    // 检测是否为嵌入式实况照片（统一在 livephoto.ts 中处理所有检测逻辑）
    const detection = await detectEmbeddedMotionPhoto(standardFile)
    if (!detection.isEmbedded) {
      return false
    }

    console.log('检测到嵌入式实况照片，开始分离:', file.name)
    theToast.info('检测到嵌入式实况照片，正在分离...', { duration: 1500 })

    // 分离图片和视频（使用已检测的数据，避免重复搜索）
    const result = await separateEmbeddedMotionPhoto(
      standardFile, 
      detection.ftypPosition, 
      detection.uint8Array
    )
    if (!result) {
      theToast.warning('嵌入式实况照片分离失败，将上传原文件', { duration: 2000 })
      return false
    }

    const { imageFile, videoFile } = result

    // 移除原文件
    // 移除原始嵌入式文件并添加分离后的文件
    uppy?.removeFile(file.id)

    // 添加分离后的图片文件
    const imageId = `embedded-image-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
    uppy?.addFile({
      id: imageId,
      name: imageFile.name,
      type: imageFile.type,
      data: imageFile,
      source: 'EmbeddedMotionPhoto',
    })

    // 添加分离后的视频文件
    const videoId = `embedded-video-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
    uppy?.addFile({
      id: videoId,
      name: videoFile.name,
      type: videoFile.type,
      data: videoFile,
      source: 'EmbeddedMotionPhoto',
    })

    console.log('✅ 实况照片分离完成:', imageFile.name, '+', videoFile.name)
    theToast.success('实况照片分离完成！', { duration: 1500 })

    return true
  } catch (error) {
    console.error('嵌入式实况照片分离失败:', error)
    theToast.warning('嵌入式实况照片分离失败，将上传原文件', { duration: 2000 })
    return false
  }
}

// 初始化 Uppy 实例
const initUppy = () => {
  // 创建 Uppy 实例
  uppy = new Uppy({
    restrictions: {
      maxNumberOfFiles: 50,
      allowedFileTypes: [
        'image/*', 
        'video/*',
        '.heic',  // 显式支持 HEIC 格式
        '.heif',  // 显式支持 HEIF 格式
      ],
    },
    autoProceed: false, // 关闭自动上传，等待 HEIC 转换完成后手动触发
  })

  // 使用 Dashboard 插件
  uppy.use(Dashboard, {
    inline: true,
    target: '#uppy-dashboard',
    hideProgressDetails: false,
    hideUploadButton: true, // 隐藏上传按钮，因为我们自动触发上传
    hideCancelButton: false,
    hideRetryButton: false,
    hidePauseResumeButton: false,
    proudlyDisplayPoweredByUppy: false,
    height: 200,
    locale: zh_CN,
    note: '支持粘贴或选择图片、视频上传哦！',
  })

  // 根据 props.TheImageSource 动态切换上传插件
  if (memorySource.value == ImageSource.LOCAL) {
    console.log('使用本地存储')
    uppy.use(XHRUpload, {
      endpoint: `${backendURL}/api/images/upload`, // 本地上传接口
      fieldName: 'file',
      formData: true,
      headers: {
        Authorization: `${getAuthToken()}`,
      },
    })
  } else if (memorySource.value == ImageSource.S3) {
    console.log('使用 S3 存储')
    uppy.use(AwsS3, {
      endpoint: '', // 走自定义的签名接口
      shouldUseMultipart: false, // 禁用分块上传
      // 每来一个文件都调用一次该函数，获取签名参数
      async getUploadParameters(file) {
        // console.log("Uploading to S3:", file)
        const fileName = file.name ? file.name : ''
        const contentType = file.type ? file.type : ''
        console.log('获取预签名fileName, contentType', fileName, contentType)

        const res = await fetchGetPresignedUrl(fileName, contentType)
        if (res.code !== 1) {
          throw new Error(res.msg || '获取预签名 URL 失败')
        }
        console.log('获取预签名成功!')
        const data = res.data as App.Api.Ech0.PresignResult
        tempFiles.value.set(data.file_name, { url: data.file_url, objectKey: data.object_key })
        return {
          method: 'PUT',
          url: data.presign_url, // 预签名 URL
          headers: {
            // 必须跟签名时的 Content-Type 完全一致
            'Content-Type': file.type,
          },
          // PUT 上传没有 fields
          fields: {},
        }
      },
    })
  }

  // 监听粘贴事件
  document.addEventListener('paste', handlePaste)

  // 添加文件时，检测实况照片对并生成 pairId
  uppy.on('files-added', async (addedFiles) => {
    if (!isLogin.value) {
      theToast.error('请先登录再上传图片 😢')
      uppy?.cancelAll()
      return
    }
    
    // 如果是嵌入式实况照片分离产生的文件，不要重复处理
    const isFromEmbedded = addedFiles.every(f => f.source === 'EmbeddedMotionPhoto')
    if (isFromEmbedded && isProcessingEmbedded.value) {
      console.log('检测到嵌入式分离产生的文件，跳过处理，等待批量检测')
      return
    }
    
    isUploading.value = true
    editorStore.MediaUploading = true
    isProcessingEmbedded.value = true
    
    // 1. 先转换 HEIC 文件
    for (const file of addedFiles) {
      await convertHeicToJpeg(file)
    }
    
    // 2. 检测并分离嵌入式实况照片
    // 注意：分离后会移除原文件并添加新文件，所以需要重新获取文件列表
    const filesToCheck = uppy?.getFiles() || []
    const embeddedSeparatedFiles: string[] = [] // 记录已分离的文件 ID
    
    for (const file of filesToCheck) {
      // 只检查新添加的文件（通过 source 判断，避免重复处理）
      if (file.source !== 'EmbeddedMotionPhoto') {
        const separated = await separateEmbeddedMotionPhotoFile(file)
        if (separated) {
          embeddedSeparatedFiles.push(file.id)
        }
      }
    }
    
    // 等待一小段时间，确保所有分离的文件都已添加
    if (embeddedSeparatedFiles.length > 0) {
      console.log('等待嵌入式实况照片分离完成...')
      await new Promise(resolve => setTimeout(resolve, 100))
    }
    
    // 3. 保存原始文件名映射（转换后的文件名）
    const currentFiles = uppy?.getFiles() || []
    for (const file of currentFiles) {
      uppyFileIdToOriginalName.value.set(file.id, file.name)
    }
    
    // 获取所有当前文件（包括之前添加的和分离后的）
    const allFiles = uppy?.getFiles() || []
    const fileObjects: File[] = allFiles.map(f => new File([f.data as Blob], f.name, { type: f.type }))
    
    // 打印所有文件信息用于调试
    console.log('=== 开始检测实况照片对 ===')
    console.log('当前文件列表:', allFiles.map(f => ({
      name: f.name,
      type: f.type,
      baseName: getBaseName(f.name)
    })))
    
    // 检测实况照片对
    const pairs = detectLivePhotoPairs(fileObjects)
    console.log('检测到的实况照片对数量:', pairs.length)
    
    // 为每个文件生成 pairId 映射（基于原始文件名）
    originalFilenameToPairId.value.clear()
    for (const pair of pairs) {
      const imgFile = allFiles[pair.imageIndex]
      const vidFile = allFiles[pair.videoIndex]
      if (imgFile && vidFile) {
        // 使用原始文件名作为 key
        originalFilenameToPairId.value.set(imgFile.name, pair.pairId)
        originalFilenameToPairId.value.set(vidFile.name, pair.pairId)
        console.log('✅ 预检测实况照片对:', {
          image: imgFile.name,
          video: vidFile.name,
          imageBaseName: getBaseName(imgFile.name),
          videoBaseName: getBaseName(vidFile.name),
          pairId: pair.pairId
        })
      }
    }
    console.log('=== 实况照片对检测完成 ===')
    
    isProcessingEmbedded.value = false
    
    // 4. 转换完成后手动触发上传
    uppy?.upload()
  })
  // 上传开始前，检查是否登录
  uppy.on('upload', () => {
    if (!isLogin.value) {
      theToast.error('请先登录再上传图片 😢')
      return
    }
    theToast.info('正在上传图片，请稍等... ⏳', { duration: 500 })
    isUploading.value = true
    editorStore.MediaUploading = true
  })
  // 单个文件上传失败后，显示错误信息
  uppy.on('upload-error', (file, error, response) => {
    if (props.TheImageSource === ImageSource.LOCAL) {
      type ResponseBody = {
        code: number
        msg: string
        // @ts-nocheck
        /* eslint-disable */
        data: any
      }

      // 判断文件类型以显示更具体的错误信息
      const isVideo = file?.type?.startsWith('video/')
      let errorMsg = isVideo ? '上传视频时发生错误 😢' : '上传图片时发生错误 😢'
      
      // @ts-nocheck
      /* eslint-disable */
      const resp = response as any // 忽略 TS 类型限制
      if (resp?.response) {
        let resObj: ResponseBody

        if (typeof resp.response === 'string') {
          resObj = JSON.parse(resp.response) as ResponseBody
        } else {
          resObj = resp.response as ResponseBody
        }

        if (resObj?.msg) {
          errorMsg = resObj.msg
        }
      }
      theToast.error(errorMsg, { duration: 3000 })
    } else if (props.TheImageSource === ImageSource.S3) {
      // S3上传失败的错误处理
      const isVideo = file?.type?.startsWith('video/')
      const errorMsg = isVideo ? '视频上传到S3失败 😢' : '图片上传到S3失败 😢'
      theToast.error(errorMsg, { duration: 3000 })
    }
    isUploading.value = false
    editorStore.MediaUploading = false
  })
  // 单个文件上传成功后，保存文件 URL 到 files 列表
  uppy.on('upload-success', (file, response) => {
    theToast.success(`好耶,上传成功！🎉`)

    // 判断文件类型
    const mediaType = file?.type?.startsWith('video/') ? 'video' : 'image'
    
    // 获取原始文件名（用于查找 pairId）
    const originalName = file?.name || ''
    const pairId = originalFilenameToPairId.value.get(originalName) || ''

    // 分两种情况: Local 或者 S3
    if (memorySource.value === ImageSource.LOCAL) {
      const res = response.body as unknown as App.Api.Response<App.Api.File.ImageDto>
      const fileUrl = String(res.data.url)
      const { width, height } = res.data
      const item: App.Api.Ech0.MediaToAdd = {
        media_url: fileUrl,
        media_type: mediaType,
        media_source: ImageSource.LOCAL,
        object_key: '',
        width: width,
        height: height,
        live_pair_id: pairId, // 应用预检测的 pairId
      }
      files.value.push(item)
      if (pairId) {
        console.log('上传成功，应用 pairId:', originalName, '->', fileUrl, 'pairId:', pairId)
      }
    } else if (memorySource.value === ImageSource.S3) {
      const uploadedFile = tempFiles.value.get(file?.name || '') || ''
      if (!uploadedFile) return

      const item: App.Api.Ech0.MediaToAdd = {
        media_url: uploadedFile.url,
        media_type: mediaType,
        media_source: ImageSource.S3,
        object_key: uploadedFile.objectKey,
        live_pair_id: pairId, // 应用预检测的 pairId
      }
      files.value.push(item)
        if (pairId) {
          console.log('上传成功，应用 pairId:', originalName, '->', uploadedFile.url, 'pairId:', pairId)
        }
    }
  })
  // 全部文件上传完成后，发射事件到父组件
  uppy.on('complete', () => {
    isUploading.value = false
    editorStore.MediaUploading = false
    
    // 打印实况照片对信息（用于调试）
    const pairGroups = new Map<string, number[]>()
    files.value.forEach((item, index) => {
      if (item.live_pair_id) {
        const group = pairGroups.get(item.live_pair_id) || []
        group.push(index)
        pairGroups.set(item.live_pair_id, group)
      }
    })
    
    // 打印检测到的实况照片对
    for (const [pairId, indexes] of pairGroups) {
      const items = indexes.map(idx => files.value[idx])
      const imageItem = items.find(item => item?.media_type === 'image')
      const videoItem = items.find(item => item?.media_type === 'video')
      if (imageItem && videoItem) {
        console.log('检测到实况照片对:', {
          pairId,
          image: imageItem.media_url,
          video: videoItem.media_url
        })
      }
    }
    
    const MediaToAddResult = [...files.value]
    editorStore.handleUppyUploaded(MediaToAddResult)
    files.value = []
    tempFiles.value.clear()
    originalFilenameToPairId.value.clear()
    uppyFileIdToOriginalName.value.clear()
  })
}

// 监听 props.TheImageSource 变化
watch(
  () => props.TheImageSource,
  (newSource, oldSource) => {
    if (newSource !== oldSource) {
      console.log('TheImageSource changed:', newSource, oldSource)
      if (!isUploading.value) {
        memorySource.value = newSource
        console.log('当前没有上传任务，可以切换上传方式')
        // 销毁旧的 Uppy 实例
        uppy?.destroy()
        uppy?.clear()
        files.value = [] // 清空已上传文件列表
        // 初始化新的 Uppy 实例
        initUppy()
      } else {
        theToast.error('当前有文件正在上传，请稍后再切换上传方式 😢')
      }
    }
  },
)

onMounted(() => {
  console.log('TheImageSource:', props.TheImageSource)
  initUppy()
})

onBeforeUnmount(() => {
  document.removeEventListener('paste', handlePaste)
})
</script>

<style scoped>
:deep(.uppy-Root) {
  border: transparent;
}

:deep(.uppy-Dashboard-innerWrap) {
  background-color: var(--image-uploader-bg-color);
}

:deep(.uppy-Dashboard-AddFiles) {
  /* 内阴影 */
  box-shadow:
    inset 0px 0px 2px rgba(80, 80, 80, 0.12),
    inset 0px 0px 2px rgba(80, 80, 80, 0.12);
}

:deep(.uppy-Dashboard-AddFiles-title) {
  color: #6f5427;
}

:deep(.uppy-Dashboard-browse) {
  color: #e5a437;
}
:deep(.uppy-StatusBar) {
  color: var(--text-color);
  background-color: var(--image-uploader-bar-bg-color);
}

:deep(.uppy-DashboardContent-bar) {
  color: var(--text-color);
  background-color: var(--image-uploader-bar-bg-color);
}

:deep(.uppy-StatusBar-statusPrimary) {
  color: var(--text-color);
}

:deep(.uppy-DashboardContent-back) {
  color: #cf8e12;
}

:deep(.uppy-DashboardContent-addMore) {
  color: #cf8e12;
}
</style>

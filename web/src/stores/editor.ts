import { defineStore } from 'pinia'
import { ref, watch, onMounted, computed } from 'vue'
import { theToast } from '@/utils/toast'
import { fetchAddEcho, fetchUpdateEcho, fetchAddTodo, fetchGetMusic, fetchGetEchoById } from '@/service/api'
import { Mode, ExtensionType, ImageSource, ImageLayout } from '@/enums/enums'
import { useEchoStore, useTodoStore, useInboxStore, usePwaStore } from '@/stores'
import { localStg } from '@/utils/storage'
import { getImageSize } from '@/utils/image'
import { useLayoutRecommend } from '@/composables/useLayoutRecommend'
import { useDebounceFn } from '@vueuse/core'

export const useEditorStore = defineStore('editorStore', () => {
  const echoStore = useEchoStore()
  const todoStore = useTodoStore()
  const inboxStore = useInboxStore()

  //================================================================
  // 编辑器状态控制
  //================================================================
  const ShowEditor = ref<boolean>(true) // 是否显示编辑器

  // ================================================================
  // 主编辑模式
  // ================================================================
  const currentMode = ref<Mode>(Mode.ECH0) // 默认为Echo编辑模式
  const currentExtensionType = ref<ExtensionType>() // 当前扩展类型（可为空）

  //================================================================
  // 编辑状态
  //================================================================
  const isSubmitting = ref<boolean>(false) // 是否正在提交
  const isUpdateMode = ref<boolean>(false) // 是否为编辑更新模式
  const MediaUploading = ref<boolean>(false) // 媒体是否正在上传

  //================================================================
  // 编辑器数据状态管理(待添加的Echo)
  //================================================================
  const echoToAdd = ref<App.Api.Ech0.EchoToAdd>({
    content: '', // 文字板块
    media: [], // 媒体板块（图片和视频）
    private: false, // 是否私密
    layout: ImageLayout.AUTO, // 媒体布局方式，默认为自动推荐
    extension: null, // 拓展内容（对于扩展类型所需的数据）
    extension_type: null, // 拓展内容类型（音乐/视频/链接/GITHUB项目）
  })

  const hasContent = computed(() => !!echoToAdd.value.content?.trim()) // 是否已填写内容
  const hasImage = computed(() => mediaListToAdd.value.length > 0) // 是否已添加图片
  const hasExtension = computed(() => {
    // 适合 Music/Video/Github
    const ext = extensionToAdd.value.extension
    const extType = extensionToAdd.value.extension_type

    // Website 多一层检测
    if (extType === ExtensionType.WEBSITE) {
      const { title, site } = websiteToAdd.value
      return !!title && !!site
    }

    return !!ext && !!extType
  })

  //================================================================
  // 编辑器数据状态管理(待添加的Todo)
  //================================================================
  const todoToAdd = ref<App.Api.Todo.TodoToAdd>({ content: '' })

  //================================================================
  // 辅助Echo的添加变量（媒体板块）
  //================================================================
  const mediaToAdd = ref<App.Api.Ech0.MediaToAdd>({
    media_url: '', // 媒体地址(依据存储方式不同而不同)
    media_type: 'image', // 媒体类型（图片/视频）
    media_source: ImageSource.LOCAL, // 媒体存储方式（本地/直链/S3）
    object_key: '', // 对象存储的Key (如果是本地存储或直链则为空)
  })
  const mediaListToAdd = ref<App.Api.Ech0.MediaToAdd[]>([]) // 最终要添加的媒体列表
  const mediaIndex = ref<number>(0) // 当前媒体索引（用于编辑媒体时定位）

  //================================================================
  // 辅助Echo的添加变量（扩展内容板块）
  //================================================================
  const websiteToAdd = ref({ title: '', site: '' }) // 辅助生成扩展内容（网站）的变量
  const videoURL = ref('') // 辅助生成扩展内容（视频）的变量
  const musicURL = ref('') // 辅助生成扩展内容（音乐）的变量
  const githubRepo = ref('') // 辅助生成扩展内容（GitHub项目）的变量
  const extensionToAdd = ref({ extension: '', extension_type: '' }) // 最终要添加的扩展内容
  const tagToAdd = ref<string>('')

  //================================================================
  // 其它状态变量
  //================================================================
  const PlayingMusicURL = ref('') // 当前正在播放的音乐URL
  const ShouldLoadMusic = ref(true) // 是否应该加载音乐（用于控制音乐播放器的加载）

  // AI 布局自动推荐
  const { recommendLayout, extractMediaInfo, extractContentInfo } = useLayoutRecommend()

  /**
   * 执行布局推荐
   * @param showToast 是否显示提示（手动调用时为 true，自动调用时为 false）
   * @returns 推荐的布局，失败返回 null
   */
  const doRecommendLayout = async (showToast: boolean = false): Promise<ImageLayout | null> => {
    if (mediaListToAdd.value.length === 0) {
      if (showToast) {
        theToast.info('请先添加图片/视频')
      }
      return null
    }

    try {
      const mediaInfo = extractMediaInfo(mediaListToAdd.value as App.Api.Ech0.Media[])
      const contentInfo = extractContentInfo(
        echoToAdd.value.content || '',
        echoToAdd.value.tags as { name: string }[] | undefined
      )
      const result = await recommendLayout(mediaInfo, contentInfo)
      echoToAdd.value.layout = result.layout

      if (showToast) {
        const layoutLabels: Record<string, string> = {
          [ImageLayout.AUTO]: '自动',
          [ImageLayout.WATERFALL]: '瀑布流',
          [ImageLayout.GRID]: '九宫格',
          [ImageLayout.CAROUSEL]: '单图轮播',
          [ImageLayout.HORIZONTAL]: '水平轮播',
        }
        const sourceIcon = result.source === 'ai' ? '🤖' : '📐'
        const reason = result.reason ? `（${result.reason}）` : ''
        theToast.success(`${sourceIcon} 推荐「${layoutLabels[result.layout]}」${reason}`)
      } else {
        console.log('[AI Layout] 自动推荐完成:', result.layout, '-', result.reason)
      }

      return result.layout
    } catch (e) {
      console.error('[AI Layout] 推荐失败:', e)
      if (showToast) {
        theToast.error('AI 推荐失败，请手动选择布局')
      }
      return null
    }
  }

  //===============================================================
  // 检测是否有实际变更（用于更新模式）
  //===============================================================
  const hasChanges = (): boolean => {
    const original = echoStore.echoToUpdate
    if (!original) return true // 没有原始数据，认为有变更

    // 比较基本内容
    if (echoToAdd.value.content !== original.content) return true
    if (echoToAdd.value.private !== original.private) return true
    if (echoToAdd.value.layout !== original.layout) return true

    // 比较标签
    const originalTagNames = original.tags?.map((tag) => tag.name).sort() || []
    const newTagName = tagToAdd.value?.trim() || ''
    const newTagNames = newTagName ? [newTagName] : []
    if (originalTagNames.length !== newTagNames.length) return true
    if (originalTagNames.some((name, index) => name !== newTagNames[index])) return true

    // 比较媒体数量
    const originalMediaCount = original.media?.length || 0
    const newMediaCount = mediaListToAdd.value.length
    if (originalMediaCount !== newMediaCount) return true

    // 比较媒体顺序（通过 URL 比较）
    const originalMediaUrls = original.media?.map((m) => m.media_url) || []
    const newMediaUrls = mediaListToAdd.value.map((m) => m.media_url)
    if (originalMediaUrls.some((url, index) => url !== newMediaUrls[index])) return true

    return false // 没有变更
  }

  //================================================================
  // 自动保存逻辑
  //================================================================
  const DRAFT_KEY = 'ech0_editor_draft'
  const isSaving = ref(false)
  const lastSavedTime = ref('')

  const saveDraft = () => {
    // 只有在非提交状态时才保存
    if (isSubmitting.value) return

    // 如果是更新模式，只有当内容发生实际变化时才保存草稿
    if (isUpdateMode.value && !hasChanges()) {
      // 如果当前没有变化，但本地存有此 updateId 的草稿，则清理草稿（可能是用户改了又改回去了）
      const draft = localStg.getItem<any>(DRAFT_KEY)
      if (draft && draft.isUpdateMode && draft.updateId === echoStore.echoToUpdate?.id) {
        clearDraft()
      }
      return
    }

    const draftData = {
      currentMode: currentMode.value,
      echoToAdd: {
        content: echoToAdd.value.content,
        private: echoToAdd.value.private,
        layout: echoToAdd.value.layout,
      },
      mediaListToAdd: mediaListToAdd.value,
      tagToAdd: tagToAdd.value,
      todoToAdd: todoToAdd.value,
      videoURL: videoURL.value,
      websiteToAdd: websiteToAdd.value,
      isUpdateMode: isUpdateMode.value,
      updateId: isUpdateMode.value ? echoStore.echoToUpdate?.id : null,
      timestamp: Date.now(),
    }

    // 检查是否为空，如果全是空的则不保存（或清理）
    if (!isUpdateMode.value) {
      const isEmpty =
        !draftData.echoToAdd.content &&
        draftData.mediaListToAdd.length === 0 &&
        !draftData.todoToAdd.content &&
        !draftData.videoURL &&
        !draftData.websiteToAdd.site

      if (isEmpty) {
        clearDraft()
        return
      }
    }

    isSaving.value = true
    localStg.setItem(DRAFT_KEY, draftData)
    lastSavedTime.value = new Date().toLocaleTimeString()

    setTimeout(() => {
      isSaving.value = false
    }, 1000)
  }

  const debouncedSave = useDebounceFn(saveDraft, 2000)

  // 监听所有相关数据的变化
  watch(
    [
      () => echoToAdd.value.content,
      () => echoToAdd.value.private,
      () => echoToAdd.value.layout,
      () => mediaListToAdd.value,
      () => tagToAdd.value,
      () => todoToAdd.value.content,
      () => videoURL.value,
      () => websiteToAdd.value.site,
      () => websiteToAdd.value.title,
    ],
    () => {
      debouncedSave()
    },
    { deep: true },
  )

  const loadDraft = () => {
    const draft = localStg.getItem<any>(DRAFT_KEY)
    if (!draft) return false

    // 如果处于更新模式且草稿是关于特定 Echo 的，则需要匹配 ID
    // 这里的逻辑比较微妙：如果用户打开了编辑器但没进更新模式，我们载入“新建”草稿
    // 如果进了更新模式，我们只在草稿也是同一个 ID 时载入（或者提示用户）

    // 简便起见，如果是新建模式（!isUpdateMode.value），则载入新建草稿
    if (!isUpdateMode.value && !draft.isUpdateMode) {
      echoToAdd.value.content = draft.echoToAdd?.content || ''
      echoToAdd.value.private = draft.echoToAdd?.private || false
      echoToAdd.value.layout = draft.echoToAdd?.layout || ImageLayout.AUTO
      mediaListToAdd.value = draft.mediaListToAdd || []
      tagToAdd.value = draft.tagToAdd || ''
      todoToAdd.value.content = draft.todoToAdd?.content || ''
      videoURL.value = draft.videoURL || ''
      websiteToAdd.value = draft.websiteToAdd || { title: '', site: '' }
      currentMode.value = draft.currentMode || Mode.ECH0
      return true
    }

    // 如果草稿是更新模式的，且当前也是更新模式且 ID 匹配
    if (isUpdateMode.value && draft.isUpdateMode && draft.updateId === echoStore.echoToUpdate?.id) {
      echoToAdd.value.content = draft.echoToAdd?.content || ''
      echoToAdd.value.private = draft.echoToAdd?.private || false
      echoToAdd.value.layout = draft.echoToAdd?.layout || ImageLayout.AUTO
      mediaListToAdd.value = draft.mediaListToAdd || []
      tagToAdd.value = draft.tagToAdd || ''
      todoToAdd.value.content = draft.todoToAdd?.content || ''
      videoURL.value = draft.videoURL || ''
      websiteToAdd.value = draft.websiteToAdd || { title: '', site: '' }
      return true
    }

    return false
  }

  const clearDraft = () => {
    localStg.removeItem(DRAFT_KEY)
    lastSavedTime.value = ''
  }

  //================================================================
  // 编辑器功能函数
  //================================================================
  // 设置当前编辑模式
  const setMode = (mode: Mode) => {
    currentMode.value = mode

    if (mode === Mode.Panel) {
      todoStore.setTodoMode(false)
      inboxStore.setInboxMode(false)
    }
  }
  // 切换当前编辑模式
  const toggleMode = () => {
    if (currentMode.value === Mode.ECH0)
      setMode(Mode.Panel) // 切换到面板模式
    else if (
      currentMode.value === Mode.TODO ||
      currentMode.value === Mode.INBOX ||
      currentMode.value === Mode.PlayMusic ||
      currentMode.value === Mode.EXTEN
    )
      setMode(Mode.Panel) // 扩展模式/TODO模式/音乐播放器模式均切换到面板模式
    else setMode(Mode.ECH0) // 其他模式均切换到Echo编辑模式
  }

  // 清空并重置编辑器
  const clearEditor = () => {
    const rememberedImageSource = ref<ImageSource>(
      localStg.getItem<ImageSource>('image_source') ?? ImageSource.LOCAL,
    )

    echoToAdd.value = {
      content: '',
      media: [],
      private: false,
      layout: ImageLayout.AUTO,
      extension: null,
      extension_type: null,
      tags: [],
    }
    mediaToAdd.value = {
      media_url: '',
      media_type: 'image',
      media_source: rememberedImageSource.value,
      object_key: '',
    }
    mediaListToAdd.value = []
    videoURL.value = ''
    musicURL.value = ''
    githubRepo.value = ''
    extensionToAdd.value = { extension: '', extension_type: '' }
    tagToAdd.value = ''
    todoToAdd.value = { content: '' }
  }

  const handleGetPlayingMusic = () => {
    ShouldLoadMusic.value = !ShouldLoadMusic.value
    fetchGetMusic().then((res) => {
      if (res.code === 1 && res.data) {
        PlayingMusicURL.value = res.data || ''
        ShouldLoadMusic.value = !ShouldLoadMusic.value
      }
    })
  }

  //===============================================================
  // 媒体模式功能函数
  //===============================================================
  // 添加更多媒体
  const handleAddMoreMedia = async () => {
    let width: number | undefined = mediaToAdd.value.width
    let height: number | undefined = mediaToAdd.value.height
    // 只对图片类型获取尺寸，视频类型跳过（避免 Image 加载视频 URL 报错）
    if ((width === undefined || height === undefined) && mediaToAdd.value.media_type === 'image') {
      try {
        const size = await getImageSize(mediaToAdd.value.media_url)
        width = size.width
        height = size.height
      } catch (error) {
        console.warn('获取图片尺寸失败:', error)
        // 获取失败时使用默认值
        width = 0
        height = 0
      }
    }
    mediaListToAdd.value.push({
      media_url: mediaToAdd.value.media_url,
      media_type: mediaToAdd.value.media_type,
      media_source: mediaToAdd.value.media_source,
      object_key: mediaToAdd.value.object_key ? mediaToAdd.value.object_key : '',
      live_pair_id: mediaToAdd.value.live_pair_id, // 传递实况照片配对ID
      width,
      height,
    })

    mediaToAdd.value = {
      media_url: '',
      media_type: 'image',
      media_source: mediaToAdd.value.media_source
        ? mediaToAdd.value.media_source
        : ImageSource.LOCAL, // 记忆存储方式
      object_key: '',
    }
  }

  const handleUppyUploaded = async (files: App.Api.Ech0.MediaToAdd[]) => {
    for (const file of files) {
      mediaToAdd.value = {
        media_url: file.media_url,
        media_type: file.media_type,
        media_source: file.media_source,
        object_key: file.object_key ? file.object_key : '',
        width: file.width,
        height: file.height,
        live_pair_id: file.live_pair_id, // 传递实况照片配对ID
      }
      await handleAddMoreMedia()
    }

    if (isUpdateMode.value && echoStore.echoToUpdate) {
      await handleAddOrUpdateEcho(true) // 仅同步媒体
    }
  }

  //===============================================================
  // 私密性切换
  //===============================================================
  const togglePrivate = () => {
    echoToAdd.value.private = !echoToAdd.value.private
  }

  //===============================================================
  // 添加或更新Echo
  //===============================================================
  const handleAddOrUpdateEcho = async (justSyncMedia: boolean) => {
    // 防止重复提交
    if (isSubmitting.value) return

    // 如果是更新模式且不是仅同步媒体，检测是否有实际变更
    if (!justSyncMedia && isUpdateMode.value && !hasChanges()) {
      theToast.info('没有需要更新的内容，已退出更新模式')

      // 保存要回到的Echo ID
      const echoId = echoStore.echoToUpdate?.id

      // 自动退出更新模式
      clearEditor()
      isUpdateMode.value = false
      echoStore.echoToUpdate = null
      setMode(Mode.ECH0)

      // 滚动回到编辑内容的位置
      if (echoId) {
        setTimeout(() => {
          scrollToEditedEcho(echoId)
        }, 100)
      }

      return
    }

    isSubmitting.value = true

    // 执行添加或更新
    try {
      // ========== 添加或更新前的检查和处理 ==========
      // 处理扩展板块
      checkEchoExtension()

      // 回填媒体板块
      echoToAdd.value.media = mediaListToAdd.value

      // 回填标签板块
      echoToAdd.value.tags = tagToAdd.value?.trim() ? [{ name: tagToAdd.value.trim() }] : []

      // 处理布局：如果是 AUTO，需要调用 AI 推荐或使用默认值
      // 注意：auto 只是前端选项，不能存入数据库
      if (echoToAdd.value.layout === ImageLayout.AUTO) {
        if (mediaListToAdd.value.length > 0) {
          // 有媒体时调用 AI 推荐（自动模式不显示toast）
          const layout = await doRecommendLayout(false)
          // 如果推荐失败，使用默认布局
          if (!layout) {
            echoToAdd.value.layout = ImageLayout.GRID
          }
        } else {
          // 没有媒体时使用默认布局
          echoToAdd.value.layout = ImageLayout.GRID
        }
      }

      // 检查Echo是否为空
      if (checkIsEmptyEcho(echoToAdd.value)) {
        const errMsg = isUpdateMode.value ? '待更新的Echo不能为空！' : '待添加的Echo不能为空！'
        theToast.error(errMsg)
        return
      }

      // ========= 添加模式 =========
      if (!isUpdateMode.value) {
        console.log('adding echo:', echoToAdd.value)
        theToast.promise(fetchAddEcho(echoToAdd.value), {
          loading: '🚀发布中...',
          success: (res) => {
            if (res.code === 1) {
              clearEditor()
              clearDraft() // 发布成功清理草稿
              echoStore.refreshEchos()
              setMode(Mode.ECH0)
              echoStore.getTags() // 刷新标签列表

              // PWA: 发布 Echo 成功后触发安装提示
              const pwaStore = usePwaStore()
              pwaStore.onEchoPublished()

              return '🎉发布成功！'
            } else {
              return '😭发布失败，请稍后再试！'
            }
          },
          error: '😭发布失败，请稍后再试！',
        })

        isSubmitting.value = false
        return
      }

      // ======== 更新模式 =========
      if (isUpdateMode.value) {
        if (!echoStore.echoToUpdate) {
          theToast.error('没有待更新的Echo！')
          return
        }

        // 回填 echoToUpdate
        echoStore.echoToUpdate.content = echoToAdd.value.content
        echoStore.echoToUpdate.private = echoToAdd.value.private
        echoStore.echoToUpdate.layout = echoToAdd.value.layout
        echoStore.echoToUpdate.media = echoToAdd.value.media
        echoStore.echoToUpdate.extension = echoToAdd.value.extension
        echoStore.echoToUpdate.extension_type = echoToAdd.value.extension_type
        echoStore.echoToUpdate.tags = echoToAdd.value.tags

        // 保存要更新的Echo ID，用于后续滚动定位
        const updatedEchoId = echoStore.echoToUpdate.id

        // 更新 Echo
        const updatePromise = fetchUpdateEcho(echoStore.echoToUpdate)

        theToast.promise(updatePromise, {
          loading: justSyncMedia ? '🔁同步图片/视频中...' : '🚀更新中...',
          success: (res) => {
            if (res.code === 1 && !justSyncMedia) {
              return '🎉更新成功！'
            } else if (res.code === 1 && justSyncMedia) {
              return '🔁发现图片/视频更改，已自动更新同步Echo！'
            } else {
              return '😭更新失败，请稍后再试！'
            }
          },
          error: '😭更新失败，请稍后再试！',
        })

        // 等待更新完成后，从服务器获取最新数据
        updatePromise.then(async (res) => {
          if (res.code === 1) {
            // 参考现有模式：直接使用 fetchGetEchoById 获取最新数据
            const latestRes = await fetchGetEchoById(String(updatedEchoId))
            if (latestRes.code === 1 && latestRes.data) {
              // 使用服务器最新数据更新本地
              echoStore.updateEcho(latestRes.data)
            }

            if (!justSyncMedia) {
              // 完整更新模式的后续处理
              clearEditor()
              clearDraft() // 更新成功清理草稿
              isUpdateMode.value = false
              echoStore.echoToUpdate = null
              setMode(Mode.ECH0)
              echoStore.getTags() // 刷新标签列表

              // 延迟滚动到编辑的内容位置
              setTimeout(() => {
                scrollToEditedEcho(updatedEchoId)
              }, 100)
            }
          }
        })
      }
    } finally {
      isSubmitting.value = false
    }
  }

  //===============================================================
  // 滚动到编辑的内容位置
  //===============================================================
  const scrollToEditedEcho = (echoId: number) => {
    // 查找对应的Echo元素
    const echoElement = document.querySelector(`[data-echo-id="${echoId}"]`)
    if (echoElement) {
      // 滚动到该元素，并留出一些顶部空间
      const elementTop = echoElement.getBoundingClientRect().top + window.pageYOffset
      const offsetTop = elementTop - 100 // 留出100px的顶部空间

      window.scrollTo({
        top: offsetTop,
        behavior: 'smooth'
      })
    }
  }

  function checkIsEmptyEcho(echo: App.Api.Ech0.EchoToAdd): boolean {
    return (
      !echo.content &&
      (!echo.media || echo.media.length === 0) &&
      !echo.extension &&
      !echo.extension_type
    )
  }

  function checkEchoExtension() {
    // 检查是否有设置扩展类型
    const { extension_type } = extensionToAdd.value
    if (extension_type) {
      // 设置了扩展类型，检查扩展内容是否为空

      switch (extension_type) {
        case ExtensionType.WEBSITE: // 处理网站扩展
          if (!handleWebsiteExtension()) {
            return
          }
          break
        default: // 其他扩展类型暂不处理
          break
      }

      // 同步至echo
      syncEchoExtension()
    } else {
      // 没有设置扩展类型，清空扩展内容
      clearExtension()
    }
  }

  function handleWebsiteExtension(): boolean {
    const { title, site } = websiteToAdd.value

    // 存在标题但无链接
    if (title && !site) {
      theToast.error('网站链接不能为空！')
      return false
    }

    // 如果有链接但没标题，补默认标题
    const finalTitle = title || (site ? '外部链接' : '')
    if (!finalTitle || !site) {
      clearExtension()
      return true
    }

    // 构建扩展内容
    extensionToAdd.value.extension = JSON.stringify({ title: finalTitle, site })
    extensionToAdd.value.extension_type = ExtensionType.WEBSITE

    return true
  }

  // 清空扩展内容
  function clearExtension() {
    extensionToAdd.value.extension = ''
    extensionToAdd.value.extension_type = ''
    echoToAdd.value.extension = null
    echoToAdd.value.extension_type = null
  }

  // 同步Echo的扩展内容
  function syncEchoExtension() {
    const { extension, extension_type } = extensionToAdd.value
    if (extension && extension_type) {
      echoToAdd.value.extension = extension
      echoToAdd.value.extension_type = extension_type
    } else {
      echoToAdd.value.extension = null
      echoToAdd.value.extension_type = null
    }
  }

  //===============================================================
  // 添加Todo
  //===============================================================
  const handleAddTodo = async () => {
    // 防止重复提交
    if (isSubmitting.value) return
    isSubmitting.value = true

    // 执行添加
    try {
      // 检查待办事项是否为空
      console.log('todo content:', todoToAdd.value.content)
      if (todoToAdd.value.content.trim() === '') {
        theToast.error('待办事项不能为空！')
        return
      }

      // 执行添加
      const res = await fetchAddTodo(todoToAdd.value)
      if (res.code === 1) {
        theToast.success('🎉添加成功！')
        todoToAdd.value = { content: '' }
        clearDraft() // 添加成功清理草稿
        todoStore.getTodos()
      }
    } finally {
      isSubmitting.value = false
    }
  }

  //===============================================================
  // 退出更新模式
  //===============================================================
  const handleExitUpdateMode = () => {
    // 保存要回到的Echo ID
    const echoId = echoStore.echoToUpdate?.id

    isUpdateMode.value = false
    echoStore.echoToUpdate = null
    clearEditor()
    clearDraft() // 退出更新模式清理该更新的草稿
    setMode(Mode.ECH0)
    theToast.info('已退出更新模式')

    // 延迟滚动回到原来编辑的内容位置
    if (echoId) {
      setTimeout(() => {
        scrollToEditedEcho(echoId)
      }, 100)
    }
  }

  //===============================================================
  // 处理不同模式下的添加或更新
  //===============================================================
  const handleAddOrUpdate = () => {
    if (todoStore.todoMode) handleAddTodo()
    else handleAddOrUpdateEcho(false)
  }

  //===============================================================
  // 处理来自系统的分享 (Web Share Target)
  //===============================================================
  const handleIncomingShare = (params: { title?: string; text?: string; url?: string }) => {
    const { title, text, url } = params
    let content = ''

    // 1. 处理标题 (过滤掉默认值和无意义的标题)
    const cleanTitle = title?.trim()
    if (cleanTitle && cleanTitle !== 'Ech0' && cleanTitle !== 'Home' && cleanTitle !== text?.trim() && cleanTitle !== url?.trim()) {
      content += `### ${cleanTitle}\n\n`
    }

    // 2. 处理文本和链接 (防重复逻辑，适配 Android)
    const cleanText = text?.trim() || ''
    const cleanUrl = url?.trim() || ''

    if (cleanText) {
      content += cleanText
      // 如果 URL 不包含在文本内，且不为空，则追加到末尾
      if (cleanUrl && !cleanText.includes(cleanUrl)) {
        content += `\n\n${cleanUrl}`
      }
    } else if (cleanUrl) {
      // 只有 URL 时直接填入
      content += cleanUrl
    }

    if (content.trim()) {
      echoToAdd.value.content = content.trim()
      setMode(Mode.ECH0)
      theToast.info('已为你填入分享的内容')
    }
  }

  const init = () => {
    handleGetPlayingMusic()
    // 载入草稿
    setTimeout(() => {
      const hasRestored = loadDraft()
      if (hasRestored) {
        theToast.info('已为你恢复上次未完成的编辑内容！')
      }
    }, 500)
  }

  return {
    // 状态
    ShowEditor,

    currentMode,
    currentExtensionType,

    isSubmitting,
    isUpdateMode,
    MediaUploading,

    echoToAdd,
    todoToAdd,

    hasContent,
    hasImage,
    hasExtension,

    mediaToAdd,
    mediaListToAdd,
    mediaIndex,

    websiteToAdd,
    videoURL,
    musicURL,
    githubRepo,
    extensionToAdd,
    tagToAdd,

    PlayingMusicURL,
    ShouldLoadMusic,

    // 方法
    init,
    setMode,
    toggleMode,
    clearEditor,
    handleGetPlayingMusic,
    handleAddMoreMedia,
    togglePrivate,
    handleAddTodo,
    handleAddOrUpdateEcho,
    handleAddOrUpdate,
    handleExitUpdateMode,
    checkIsEmptyEcho,
    checkEchoExtension,
    syncEchoExtension,
    clearExtension,
    handleUppyUploaded,
    scrollToEditedEcho,
    doRecommendLayout,
    handleIncomingShare,

    // 自动保存状态
    isSaving,
    lastSavedTime,
    loadDraft,
    clearDraft,
  }
})

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useFetch } from '@vueuse/core'
import { theToast } from '@/utils/toast'
import { useConnectStore } from './connect'

// 每个 Hub 的独立状态
interface HubState {
  url: string
  buffer: App.Api.Hub.Echo[] // 缓冲池
  currentPage: number // 独立分页
  hasMore: boolean
  isLoading: boolean
}

export const useHubStore = defineStore('hubStore', () => {
  /**
   * state
   */

  const connectStore = useConnectStore()

  // hub
  const hubList = ref<App.Api.Hub.HubList>([])
  const hubinfoList = ref<App.Api.Hub.HubInfoList>([])
  const hubInfoMap = ref<Map<string, App.Api.Hub.HubItemInfo>>(new Map())
  const hubStates = ref<Map<string, HubState>>(new Map()) // 各 Hub 的独立状态

  // echo
  const echoList = ref<App.Api.Hub.Echo[]>([]) // 存储Echo列表（展示列表）
  const existingIds = ref<Set<string>>(new Set()) // 已存在的 Echo ID，用于去重

  const isPreparing = ref<boolean>(true) // 是否正在准备数据
  const isLoading = ref<boolean>(false) // 是否正在加载数据
  const hasTriedInitialLoad = ref<boolean>(false) // 是否已尝试过首次加载（用于空态展示）
  const pageSize = ref<number>(10) // 每个 Hub 每次请求的数量
  const batchSize = ref<number>(10) // 每次归并取数的数量
  const hasMore = ref<boolean>(true) // 是否还有更多数据可加载

  /**
   * utils
   */

  // 数据标准化函数：将旧版本的 images 字段转换为 media 字段
  const normalizeEchoData = (echo: any, serverUrl: string): App.Api.Hub.Echo => {
    // 如果没有 media 字段或 media 为空，但有 images 字段，则进行转换
    if ((!echo.media || echo.media.length === 0) && echo.images && Array.isArray(echo.images)) {
      // 将 images 转换为 media 格式
      echo.media = echo.images.map((image: any) => ({
        id: image.id,
        message_id: image.message_id,
        media_url: image.image_url || image.media_url, // 兼容两种字段名
        media_type: 'image' as const, // 旧版本只支持图片
        media_source: image.image_source || image.media_source, // 兼容两种字段名
        object_key: image.object_key,
        width: image.width,
        height: image.height,
      }))

      // 开发环境日志
      if (import.meta.env.DEV) {
        console.log('[兼容性转换] 检测到旧版本数据格式，已转换 images → media', {
          echoId: echo.id,
          serverUrl: serverUrl,
          imagesCount: echo.images.length,
        })
      }
    }

    // 如果既没有 media 也没有 images，设置为空数组
    if (!echo.media || !Array.isArray(echo.media)) {
      echo.media = []
      if (import.meta.env.DEV && echo.images === undefined) {
        console.warn('[兼容性转换] Echo 数据缺少 media 和 images 字段', {
          echoId: echo.id,
          serverUrl: serverUrl,
        })
      }
    }

    return echo
  }

  /**
   * actions
   */

  // 1. 获取hubList
  const getHubList = async () => {
    isPreparing.value = true
    hasTriedInitialLoad.value = false
    await connectStore.getConnect()

    hubList.value = connectStore.connects
  }

  // 2. 根据hubList 获取每个item的info
  const getHubInfoList = async () => {
    if (hubList.value.length === 0) {
      theToast.info('Hub列表为空，请到设置中添加Connect吧~')
      isPreparing.value = false
      return
    }

    // 处理 hubList 中的每个Hub（末尾的 / 去除）
    hubList.value = hubList.value.map((item) => {
      return typeof item === 'string'
        ? item.endsWith('/')
          ? item.slice(0, -1)
          : item
        : item.connect_url.endsWith('/')
          ? {
            ...item,
            connect_url: item.connect_url.slice(0, -1),
          }
          : item
    })

    // 创建带超时的请求函数
    const fetchWithTimeout = async (
      url: string,
      timeout: number = 5000,
    ): Promise<App.Api.Hub.HubItemInfo | null> => {
      return new Promise((resolve) => {
        let isResolved = false

        // 设置超时
        const timeoutId = setTimeout(() => {
          if (!isResolved) {
            isResolved = true
            console.warn(`[Hub] 请求超时: ${url}`)
            resolve(null)
          }
        }, timeout)

          // 发起请求
          ; (async () => {
            try {
              const { error, data } = await useFetch<App.Api.Response<App.Api.Hub.HubItemInfo>>(
                `${url}/api/connect`,
              ).json()

              clearTimeout(timeoutId)
              if (!isResolved) {
                isResolved = true
                if (error.value || data.value?.code !== 1) {
                  console.warn(`[Hub] 请求失败: ${url}`, error.value)
                  resolve(null)
                } else {
                  resolve(data.value?.data || null)
                }
              }
            } catch (err) {
              clearTimeout(timeoutId)
              if (!isResolved) {
                isResolved = true
                console.error(`[Hub] 请求异常: ${url}`, err)
                resolve(null)
              }
            }
          })()
      })
    }

    // 使用 Promise.allSettled 来并行获取每个Hub的info
    const promises = hubList.value.map(async (hub) => {
      const url = typeof hub === 'string' ? hub : hub.connect_url
      return await fetchWithTimeout(url, 5000) // 5秒超时
    })

    const results = await Promise.allSettled(promises)

    // 收集成功的结果，并从 hubList 中移除失败的实例
    const validHubs: typeof hubList.value = []
    const failedHubs: string[] = []

    results.forEach((result, index) => {
      const hub = hubList.value[index]
      if (!hub) return // 防止 undefined

      const hubUrl = typeof hub === 'string' ? hub : hub.connect_url

      if (result.status === 'fulfilled' && result.value) {
        // 成功获取信息
        hubinfoList.value.push(result.value)
        validHubs.push(hub)

        // 将Hub信息存入Map
        if (typeof hubUrl === 'string') {
          hubInfoMap.value.set(hubUrl, result.value)
        }
      } else {
        // 失败的实例，记录并排除
        if (typeof hubUrl === 'string') {
          failedHubs.push(hubUrl)
          console.warn(`[Hub] 实例不可用，已排除: ${hubUrl}`)
        }
      }
    })

    // 更新 hubList，只保留可用的实例
    hubList.value = validHubs

    // 提示用户
    // if (failedHubs.length > 0) {
    //   theToast.warning(`${failedHubs.length} 个实例不可用，已自动排除`)
    // }

    // 处理结果
    if (hubList.value.length === 0) {
      theToast.info('当前Hub暂无可连接的实例。')
      isPreparing.value = false
      return
    }

    // 初始化各 Hub 的独立状态
    hubStates.value.clear()
    for (const hub of hubList.value) {
      const url = typeof hub === 'string' ? hub : hub.connect_url
      hubStates.value.set(url, {
        url,
        buffer: [],
        currentPage: 1,
        hasMore: true,
        isLoading: false,
      })
    }

    isPreparing.value = false
    theToast.success(`成功连接 ${hubList.value.length} 个实例，开始加载 Echos`)

    // 并行请求所有 Hub 的第一页，填充缓冲池
    await Promise.all(Array.from(hubStates.value.keys()).map((url) => fetchHubPage(url)))
  }

  // 3. 请求某个 Hub 的下一页数据到其缓冲池
  const fetchHubPage = async (hubUrl: string): Promise<void> => {
    const state = hubStates.value.get(hubUrl)
    if (!state || state.isLoading || !state.hasMore) return

    state.isLoading = true
    try {
      const { error, data } = await useFetch<App.Api.Response<App.Api.Ech0.PaginationResult>>(
        hubUrl + '/api/echo/page',
      )
        .post({
          page: state.currentPage,
          pageSize: pageSize.value,
        })
        .json()

      if (error.value || data.value?.code !== 1) {
        console.warn(`[Hub] 请求失败: ${hubUrl}`, error.value)
        state.hasMore = false
        return
      }

      const items = (data.value?.data.items || []).map((echo: App.Api.Ech0.Echo) => {
        // 先进行数据标准化（images → media 转换）
        const normalizedEcho = normalizeEchoData(echo, hubUrl)

        // 然后再添加 Hub 相关字段
        return {
          ...normalizedEcho,
          createdTs: new Date(normalizedEcho.created_at).getTime(),
          server_name: hubInfoMap.value.get(hubUrl)?.server_name || 'Ech0',
          server_url: hubUrl,
          // 设置echo.logo为站点Logo（来自/api/connect接口）
          logo:
            hubInfoMap.value.get(hubUrl)?.logo && hubInfoMap.value.get(hubUrl)?.logo !== ''
              ? hubInfoMap.value.get(hubUrl)?.logo
              : '/Ech0.svg',
        }
      })

      // 按时间降序排序后追加到缓冲池
      items.sort((a: App.Api.Hub.Echo, b: App.Api.Hub.Echo) => b.createdTs - a.createdTs)
      state.buffer.push(...items)
      state.currentPage++
      state.hasMore = items.length >= pageSize.value
    } catch (err) {
      console.error(`[Hub] 请求异常: ${hubUrl}`, err)
      state.hasMore = false
    } finally {
      state.isLoading = false
    }
  }

  // 4. 归并取数：从各 Hub 缓冲池中按时间顺序取出数据
  const loadEchoListPage = async () => {
    if (isLoading.value || isPreparing.value) return

    // 检查是否还有更多数据可加载
    const canLoadMore = Array.from(hubStates.value.values()).some(
      (s) => s.hasMore || s.buffer.length > 0,
    )
    if (!canLoadMore) {
      hasMore.value = false
      hasTriedInitialLoad.value = true
      return
    }



    isLoading.value = true
    try {
      const result: App.Api.Hub.Echo[] = []
      let attempts = 0
      const maxAttempts = batchSize.value * 3 // 防止死循环

      while (result.length < batchSize.value && attempts < maxAttempts) {
        attempts++

        // 1. 找出所有缓冲池中时间最新（createdTs 最大）的那条
        let maxTs = -1
        let maxHubUrl: string | null = null

        for (const [url, state] of hubStates.value) {
          const head = state.buffer[0]
          if (head) {
            const headTs = head.createdTs
            if (headTs > maxTs) {
              maxTs = headTs
              maxHubUrl = url
            }
          }
        }

        // 2. 如果所有缓冲池都空了，尝试补充
        if (maxHubUrl === null) {
          const emptyHubsWithMore = Array.from(hubStates.value.values()).filter(
            (s) => s.hasMore && !s.isLoading && s.buffer.length === 0,
          )

          if (emptyHubsWithMore.length === 0) {
            // 真的没有更多数据了
            break
          }

          // 并行补充所有空缓冲池
          await Promise.all(emptyHubsWithMore.map((s) => fetchHubPage(s.url)))
          continue
        }

        // 3. 取出这条数据
        const state = hubStates.value.get(maxHubUrl)!
        const echo = state.buffer.shift()!

        // 去重检查
        const key = `${echo.server_url}-${echo.id}`
        if (!existingIds.value.has(key)) {
          existingIds.value.add(key)
          result.push(echo)
        }

        // 4. 如果这个 Hub 的缓冲池快空了，提前补充（预加载）
        if (state.buffer.length < 3 && state.hasMore && !state.isLoading) {
          fetchHubPage(maxHubUrl) // 异步补充，不等待
        }
      }

      // 追加到展示列表（不重排序，已经是按时间顺序取出的）
      echoList.value.push(...result)

      // 更新 hasMore 状态
      hasMore.value = Array.from(hubStates.value.values()).some(
        (s) => s.hasMore || s.buffer.length > 0,
      )

      if (!hasMore.value && echoList.value.length > 0) {
        theToast.info('没有更多数据了🙃')
      }
    } finally {
      isLoading.value = false
      hasTriedInitialLoad.value = true
    }
  }

  return {
    echoList,
    hubList,
    hubInfoMap,
    hubinfoList,
    hubStates,
    isLoading,
    isPreparing,
    hasTriedInitialLoad,
    pageSize,
    batchSize,
    hasMore,
    getHubList,
    getHubInfoList,
    loadEchoListPage,
  }
})

import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { fetchGetConnectList, fetchGetAllConnectInfo } from '@/service/api'
import { localStg } from '@/utils/storage'

export const useConnectStore = defineStore('connectStore', () => {
  /**
   * State
   */
  const connects = ref<App.Api.Connect.Connected[]>([])
  const connectsInfo = ref<App.Api.Connect.Connect[]>([])
  const loading = ref<boolean>(false)
  const hubUpdateCount = ref<number>(0)
  const hubUpdateSites = ref<number>(0)
  const lastFetchTime = ref<number>(0) // 上次获取数据的时间
  const cacheTimeout = 3 * 60 * 1000 // 同步为3分钟，平衡实时性与负载
  let pollInterval: ReturnType<typeof setInterval> | null = null

  /**
   * Tooltip 提示
   */
  const hubUpdateTooltip = computed(() => {
    if (hubUpdateCount.value === 0) return ''
    return `来自 ${hubUpdateSites.value} 个站点的 ${hubUpdateCount.value} 条新动态`
  })

  /**
   * Actions
   */
  async function getConnect() {
    await fetchGetConnectList()
      .then((res) => {
        if (res.code === 1) {
          connects.value = res.data
        }
      })
      .catch((err) => {
        console.error(err)
      })
  }

  const getConnectInfo = (forceRefresh = false) => {
    const now = Date.now()

    // 如果有缓存数据且未过期，且不是强制刷新，则直接返回
    if (!forceRefresh && connectsInfo.value.length > 0 && (now - lastFetchTime.value) < cacheTimeout) {
      loading.value = false
      return
    }

    // 如果已经在加载中，且不是强制刷新，则直接返回，避免并发重复请求
    if (!forceRefresh && loading.value) {
      return
    }

    loading.value = true
    fetchGetAllConnectInfo()
      .then((res) => {
        if (res.code === 1) {
          connectsInfo.value = res.data
          lastFetchTime.value = now
          // 检查Hub更新
          checkHubUpdates()
        }
      })
      .catch((err) => {
        console.error('[Connect] 获取连接信息失败:', err)
      })
      .finally(() => {
        loading.value = false
      })
  }

  /**
   * 检查Hub更新数量
   */
  const checkHubUpdates = () => {
    if (connectsInfo.value.length === 0) return

    // 获取按站点存储的上次记录数量
    const lastSiteCounts = localStg.getItem<Record<string, number>>('hubSiteCounts') || {}

    let totalNewUpdates = 0
    let affectedSites = 0
    let storageDirty = false

    // 1. 检查更新并识别新站点
    connectsInfo.value.forEach((connect) => {
      const lastCount = lastSiteCounts[connect.server_url]

      if (lastCount !== undefined) {
        // 已存在的站点，检查增量
        if (connect.total_echos > lastCount) {
          totalNewUpdates += (connect.total_echos - lastCount)
          affectedSites++
        }
      } else {
        // 新发现的站点，存入记录但不触发红点
        lastSiteCounts[connect.server_url] = connect.total_echos
        storageDirty = true
      }
    })

    // 2. 清理已删除的站点记录
    const currentUrls = new Set(connectsInfo.value.map((c) => c.server_url))
    Object.keys(lastSiteCounts).forEach((url) => {
      if (!currentUrls.has(url)) {
        delete lastSiteCounts[url]
        storageDirty = true
      }
    })

    // 如果存储有变化（新增或删除站点），同步到 localStorage
    if (storageDirty) {
      localStg.setItem('hubSiteCounts', lastSiteCounts)
    }

    // 调试信息
    console.log(`[Hub] 新增: ${totalNewUpdates} 条, 涉及站点: ${affectedSites}`)

    hubUpdateCount.value = totalNewUpdates
    hubUpdateSites.value = affectedSites

    // 如果是首次访问（没有记录且还没初始化过）
    if (Object.keys(lastSiteCounts).length === 0 && connectsInfo.value.length > 0) {
      const currentSiteCounts: Record<string, number> = {}
      connectsInfo.value.forEach((connect) => {
        currentSiteCounts[connect.server_url] = connect.total_echos
      })
      localStg.setItem('hubSiteCounts', currentSiteCounts)
      hubUpdateCount.value = 0
      hubUpdateSites.value = 0
    }
  }

  /**
   * 清除Hub更新提示
   */
  const clearHubUpdates = () => {
    // 无论是否有更新，都更新基准记录，以保持最新同步
    const currentSiteCounts: Record<string, number> = {}
    connectsInfo.value.forEach((connect) => {
      currentSiteCounts[connect.server_url] = connect.total_echos
    })
    localStg.setItem('hubSiteCounts', currentSiteCounts)

    // 清除旧的聚合记录（如果有）
    localStg.removeItem('hubTotalEchos')

    if (hubUpdateCount.value > 0) {
      console.log(`[Hub] 已读 ${hubUpdateCount.value} 条更新`)
    }

    hubUpdateCount.value = 0
    hubUpdateSites.value = 0
  }

  /**
   * 开始定期轮询
   */
  const startPolling = () => {
    // 首次立即获取
    getConnectInfo()

    // 每3分钟刷新一次（与缓存周期同步）
    if (!pollInterval) {
      pollInterval = setInterval(() => {
        getConnectInfo()
      }, 3 * 60 * 1000)
    }
  }

  /**
   * 停止定期轮询
   */
  const stopPolling = () => {
    if (pollInterval) {
      clearInterval(pollInterval)
      pollInterval = null
    }
  }

  return {
    connects,
    connectsInfo,
    loading,
    hubUpdateCount,
    hubUpdateSites,
    hubUpdateTooltip,
    getConnect,
    getConnectInfo,
    checkHubUpdates,
    clearHubUpdates,
    startPolling,
    stopPolling,
  }
})

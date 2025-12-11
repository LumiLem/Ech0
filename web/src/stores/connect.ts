import { ref } from 'vue'
import { defineStore } from 'pinia'
import { fetchGetConnectList, fetchGetAllConnectInfo } from '@/service/api'
import { localStg } from '@/utils/storage'

export const useConnectStore = defineStore('connectStore', () => {
  /**
   * State
   */
  const connects = ref<App.Api.Connect.Connected[]>([])
  const connectsInfo = ref<App.Api.Connect.Connect[]>([])
  const loading = ref<boolean>(true)
  const hubUpdateCount = ref<number>(0)
  const lastFetchTime = ref<number>(0) // 上次获取数据的时间
  const cacheTimeout = 5 * 60 * 1000 // 5分钟缓存

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
    
    // 计算当前总数
    const currentTotal = connectsInfo.value.reduce((sum, connect) => sum + connect.total_echos, 0)
    
    // 获取上次记录的总数
    const lastTotal = localStg.getItem<number>('hubTotalEchos') || 0
    
    // 计算更新数量
    const updateCount = Math.max(0, currentTotal - lastTotal)
    
    // 调试信息
    console.log(`[Hub] 总数: ${currentTotal}, 基准: ${lastTotal}, 新增: ${updateCount}`)
    
    hubUpdateCount.value = updateCount
    
    // 如果是首次访问，不显示更新数量
    if (lastTotal === 0) {
      hubUpdateCount.value = 0
      // 保存当前总数作为基准
      localStg.setItem('hubTotalEchos', currentTotal)
    }
  }

  /**
   * 清除Hub更新提示
   */
  const clearHubUpdates = () => {
    if (hubUpdateCount.value > 0) {
      // 更新基准总数
      const currentTotal = connectsInfo.value.reduce((sum, connect) => sum + connect.total_echos, 0)
      console.log(`[Hub] 已读 ${hubUpdateCount.value} 条更新`)
      localStg.setItem('hubTotalEchos', currentTotal)
      hubUpdateCount.value = 0
    }
  }

  return {
    connects,
    connectsInfo,
    loading,
    hubUpdateCount,
    getConnect,
    getConnectInfo,
    checkHubUpdates,
    clearHubUpdates,
  }
})

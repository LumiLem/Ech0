import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { fetchGetEchosByPage, fetchGetTags, fetchGetEchosByTagId, fetchGetEchosByDate } from '@/service/api'

export const useEchoStore = defineStore('echoStore', () => {
  /**
   * state
   */
  const echoList = ref<App.Api.Ech0.Echo[]>([]) // 存储Echo列表
  const echoIndexMap = ref(new Map<number, number>()) // id -> index
  const isLoading = ref<boolean>(true) // 是否正在加载数据
  const total = ref<number>(0) // 总数据量
  const pageSize = ref<number>(5) // 每页显示的数量
  const page = ref<number>(0) // 当前页码，从0开始计数
  const current = ref<number>(1) // 当前页码，从1开始计数
  const searchValue = ref<string>('') // 搜索关键词
  const hasMore = computed(() => {
    return total.value > echoList.value.length
  }) // 是否还有更多数据可加载
  const searchingMode = computed(() => {
    return searchValue.value.length > 0
  }) // 是否处于搜索模式
  const echoToUpdate = ref<App.Api.Ech0.EchoToUpdate | null>(null) // 用于更新Echo的临时存储
  const tagList = ref<App.Api.Ech0.Tag[]>([]) // 标签列表
  const tagOptions = computed<string[]>(() => {
    return tagList.value.map((tag) => tag.name)
  })

  const isFilteringMode = ref<boolean>(false) // 是否正在通过标签过滤
  const filteredEchoList = ref<App.Api.Ech0.Echo[]>([]) // 过滤后的Echo列表
  const filteredEchoIndexMap = ref(new Map<number, number>()) // 过滤后的id -> index
  const filteredTotal = ref<number>(0) // 过滤后的总数据量
  const filteredPageSize = ref<number>(5) // 过滤后的每页显示的数量
  const filteredPage = ref<number>(0) // 过滤后的当前页码，从0开始计数
  const filteredCurrent = ref<number>(1) // 过滤后的当前页码，从1开始计数
  const filteredSearchValue = ref<string>('') // 过滤后的搜索关键词
  const filteredHasMore = computed(() => {
    return filteredTotal.value > filteredEchoList.value.length
  }) // 过滤后是否还有更多数据可加载
  const filteredTag = ref<App.Api.Ech0.Tag | null>(null) // 当前用于过滤的标签
  const filteredSearchingMode = computed(() => {
    return filteredSearchValue.value.length > 0
  }) // 过滤后是否处于搜索模式

  // 日期筛选相关
  const isDateFilteringMode = ref<boolean>(false) // 是否正在通过日期过滤
  const filteredDate = ref<string | null>(null) // 当前用于过滤的日期 (YYYY-MM-DD)
  const filteredYearMonth = ref<{ year: number; month: number } | null>(null) // 当前用于过滤的年月

  // 监听 searchingMode 的变化
  watch(searchingMode, (newValue, oldValue) => {
    // 如果从搜索模式切换到非搜索模式，重置当前页码和数据列表
    if (newValue === false && oldValue === true) {
      refreshEchos()
    }
  })

  // watch(filteredSearchingMode, (newValue, oldValue) => {
  //   // 如果从搜索模式切换到非搜索模式，重置当前页码和数据列表
  //   if (newValue === false && oldValue === true) {
  //     refreshEchosForFilter()
  //   }
  // })

  // 监听 isFilteringMode 的变化
  watch(isFilteringMode, (newValue, oldValue) => {
    // 如果从过滤模式切换到非过滤模式，重置当前页码和数据列表
    if (newValue === false && oldValue === true) {
      refreshEchosForFilter()
    }
  })

  // 监听 isDateFilteringMode 的变化
  watch(isDateFilteringMode, (newValue, oldValue) => {
    // 如果从日期过滤模式切换到非过滤模式，重置当前页码和数据列表
    if (newValue === false && oldValue === true) {
      refreshEchosForFilter()
    }
  })

  /**
   * actions
   */
  async function getEchosByPage() {
    if (current.value <= page.value) return

    isLoading.value = true

    await fetchGetEchosByPage({
      page: current.value,
      pageSize: pageSize.value,
      search: searchValue.value,
    })
      .then((res) => {
        if (res.code === 1) {
          total.value = res.data.total

          // 同步更新 echoMap
          res.data.items.forEach((item: App.Api.Ech0.Echo) => {
            const idx = echoIndexMap.value.get(item.id)
            if (idx !== undefined) {
              echoList.value[idx] = item // 更新已有数据
            } else {
              echoList.value.push(item) // 添加新数据
              echoIndexMap.value.set(item.id, echoList.value.length - 1)
            }
          })

          page.value += 1
        }
      })
      .finally(() => {
        isLoading.value = false
      })
  }

  const refreshEchos = () => {
    current.value = 1
    page.value = 0
    echoList.value = []
    echoIndexMap.value.clear()
    getEchosByPage()
  }

  const clearEchos = () => {
    current.value = 1
    page.value = 0
    echoList.value = []
    echoIndexMap.value.clear()
    total.value = 0
  }

  const refreshForSearch = () => {
    current.value = 1
    page.value = 0
    echoList.value = []
    echoIndexMap.value.clear()
  }

  const updateEcho = (echo: App.Api.Ech0.Echo) => {
    // 更新主列表
    const idx = echoIndexMap.value.get(echo.id)
    if (idx !== undefined) {
      echoList.value[idx] = echo // 更新
    }
    
    // 更新过滤列表（如果存在）
    const filteredIdx = filteredEchoIndexMap.value.get(echo.id)
    if (filteredIdx !== undefined) {
      filteredEchoList.value[filteredIdx] = echo // 更新过滤列表
    }
  }

  const updateLikeCount = (echoId: number, delta: number = 1) => {
    const idx = echoIndexMap.value.get(echoId)
    if (idx !== undefined) {
      const targetEcho = echoList.value[idx]
      if (targetEcho) {
        targetEcho.fav_count = (targetEcho.fav_count || 0) + delta
        echoList.value[idx] = { ...targetEcho } // 保证响应式触发
      }
    }
  }

  const getTags = async () => {
    const res = await fetchGetTags()
    if (res.code === 1) {
      // 清空原数组再插入新数据
      tagList.value.splice(0, tagList.value.length, ...res.data)
    }
  }

  const refreshEchosForFilter = () => {
    filteredCurrent.value = 1
    filteredPage.value = 0
    filteredEchoList.value = []
    filteredEchoIndexMap.value.clear()
  }

  async function getEchosByPageForFilter() {
    if (filteredCurrent.value <= filteredPage.value) return

    // 标签筛选模式
    if (filteredTag.value) {
      isLoading.value = true
      await fetchGetEchosByTagId(filteredTag.value.id, {
        page: filteredCurrent.value,
        pageSize: filteredPageSize.value,
        search: filteredSearchValue.value || '',
      })
      .then((res) => {
        if (res.code === 1) {
          filteredTotal.value = res.data.total

          // 同步更新 echoMap
          res.data.items.forEach((item: App.Api.Ech0.Echo) => {
            const idx = filteredEchoIndexMap.value.get(item.id)
            if (idx !== undefined) {
              filteredEchoList.value[idx] = item // 更新已有数据
            } else {
              filteredEchoList.value.push(item) // 添加新数据
              filteredEchoIndexMap.value.set(item.id, filteredEchoList.value.length - 1)
            }
          })

          filteredPage.value += 1
        }
      })
      .finally(() => {
        isLoading.value = false
      })
      return
    }

    // 日期筛选模式
    if (filteredDate.value || filteredYearMonth.value) {
      isLoading.value = true
      const params: App.Api.Ech0.ParamsByPagination & { date?: string; year?: number; month?: number } = {
        page: filteredCurrent.value,
        pageSize: filteredPageSize.value,
        search: filteredSearchValue.value || '',
      }
      if (filteredDate.value) {
        params.date = filteredDate.value
      } else if (filteredYearMonth.value) {
        params.year = filteredYearMonth.value.year
        params.month = filteredYearMonth.value.month
      }

      await fetchGetEchosByDate(params)
        .then((res) => {
          if (res.code === 1) {
            filteredTotal.value = res.data.total

            res.data.items.forEach((item: App.Api.Ech0.Echo) => {
              const idx = filteredEchoIndexMap.value.get(item.id)
              if (idx !== undefined) {
                filteredEchoList.value[idx] = item
              } else {
                filteredEchoList.value.push(item)
                filteredEchoIndexMap.value.set(item.id, filteredEchoList.value.length - 1)
              }
            })

            filteredPage.value += 1
          }
        })
        .finally(() => {
          isLoading.value = false
        })
    }
  }

  const refreshForFilterSearch = () => {
    filteredCurrent.value = 1
    filteredPage.value = 0
    filteredEchoList.value = []
    filteredEchoIndexMap.value.clear()
  }

  const init = () => {
    getTags()
  }

  return {
    echoList,
    echoIndexMap,
    isLoading,
    total,
    pageSize,
    page,
    current,
    searchValue,
    searchingMode,
    hasMore,
    echoToUpdate,
    tagList,
    tagOptions,
    isFilteringMode,
    filteredEchoList,
    filteredEchoIndexMap,
    filteredTotal,
    filteredPageSize,
    filteredPage,
    filteredCurrent,
    filteredSearchValue,
    filteredHasMore,
    filteredTag,
    filteredSearchingMode,
    isDateFilteringMode,
    filteredDate,
    filteredYearMonth,
    refreshForFilterSearch,
    getEchosByPageForFilter,
    refreshEchosForFilter,
    getEchosByPage,
    refreshEchos,
    clearEchos,
    refreshForSearch,
    updateEcho,
    updateLikeCount,
    getTags,
    init,
  }
})

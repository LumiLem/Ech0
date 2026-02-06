/**
 * PWA Store - 管理 PWA 安装提示和状态
 *
 * 核心设计原则：
 * 1. 只有支持 PWA 安装的浏览器才显示安装相关 UI
 * 2. beforeinstallprompt 事件不是每次页面加载都会触发
 * 3. 需要持久化记录浏览器是否曾经触发过安装事件
 * 4. iOS Safari 使用独立的后备方案
 * 5. 对于不支持 PWA 的浏览器，不显示任何安装内容
 *
 * 参考文档：
 * - https://web.dev/articles/customize-install
 * - https://web.dev/learn/pwa/installation-prompt
 * - https://web.dev/articles/promote-install
 */
import { ref, computed, watch } from 'vue'
import { defineStore } from 'pinia'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { useInboxStore } from './inbox'
import { useTodoStore } from './todo'
import { useConnectStore } from './connect'
import router from '@/router'

// BeforeInstallPromptEvent 类型定义（浏览器原生事件）
interface BeforeInstallPromptEvent extends Event {
  readonly platforms: string[]
  readonly userChoice: Promise<{
    outcome: 'accepted' | 'dismissed'
    platform: string
  }>
  prompt(): Promise<void>
}

// PWA 存储键名
const PWA_STORAGE_KEY = 'ech0_pwa_state'

// PWA 持久化状态
interface PwaStorageState {
  // 用户是否已关闭/拒绝安装提示
  dismissed: boolean
  // 关闭时间戳（用于冷却期计算）
  dismissedAt: number
  // 访问次数
  visitCount: number
  // 上次访问时间
  lastVisit: number
  // 是否曾经收到过 beforeinstallprompt 事件
  // 这是关键：用于在刷新页面后判断浏览器是否支持 PWA 安装
  hasSeenInstallPrompt: boolean
}

const getDefaultStorageState = (): PwaStorageState => ({
  dismissed: false,
  dismissedAt: 0,
  visitCount: 0,
  lastVisit: 0,
  hasSeenInstallPrompt: false,
})

export const usePwaStore = defineStore('pwaStore', () => {
  // ================================================================
  // State - 响应式状态
  // ================================================================

  // 保存的 beforeinstallprompt 事件（不持久化，页面刷新后为 null）
  const deferredPrompt = ref<BeforeInstallPromptEvent | null>(null)

  // 是否已安装 PWA（通过 display-mode 检测）
  const isInstalled = ref(false)

  // 是否正在安装中
  const isInstalling = ref(false)

  // 是否为 iOS 设备
  const isIOS = ref(false)

  // 是否为 Safari 浏览器（包括 iOS Safari 和 macOS Safari）
  const isSafari = ref(false)

  // 是否为 Firefox 浏览器（不支持 beforeinstallprompt）
  const isFirefox = ref(false)

  // 当前会话是否已显示过提示
  const hasShownPromptThisSession = ref(false)

  // 浏览器是否支持 PWA 安装（持久化缓存的结果）
  const hasPwaSupport = ref(false)

  // ================================================================
  // Stores - 关联的 Store
  // ================================================================
  const inboxStore = useInboxStore()
  const todoStore = useTodoStore()
  const connectStore = useConnectStore()

  // ================================================================
  // Computed - 派生状态
  // ================================================================

  /**
   * 是否可以显示标准安装按钮（非 iOS）
   *
   * 条件：
   * 1. 未安装
   * 2. 当前有 deferredPrompt（可以触发安装）
   *
   * 注意：这个条件比较严格，只有在 deferredPrompt 可用时才返回 true
   * 这意味着刷新页面后，如果 beforeinstallprompt 没有再次触发，按钮会消失
   * 这是符合 Google 官方建议的行为
   */
  const canShowInstall = computed(() => {
    return !isInstalled.value && !!deferredPrompt.value
  })

  /**
   * 是否可以显示"设置页面"的安装入口
   *
   * 这个条件更宽松，用于在设置页面显示一个提示或按钮
   * 即使 deferredPrompt 不可用，只要曾经收到过该事件，就显示入口
   * 点击后会：
   * - 如果 deferredPrompt 可用：触发安装
   * - 如果不可用：显示手动安装说明
   */
  const canShowInstallEntry = computed(() => {
    if (isInstalled.value) return false
    if (isIOS.value) return false // iOS 使用独立的入口

    // 获取持久化状态
    const state = getStorageState()
    // 有当前事件，或者曾经收到过事件
    return !!deferredPrompt.value || state.hasSeenInstallPrompt
  })

  /**
   * 是否可以显示 iOS 的安装入口
   *
   * iOS Safari 不支持 beforeinstallprompt，但支持"添加到主屏幕"
   * 需要显示引导用户手动操作的说明
   */
  const canShowIOSInstallEntry = computed(() => {
    return isIOS.value && !isInstalled.value
  })

  /**
   * 是否支持 PWA 安装功能
   *
   * 用于判断是否应该显示任何与 PWA 安装相关的 UI
   * 对于不支持的浏览器（如 Firefox、一些老旧浏览器），不显示 any 安装内容
   */
  const isPwaSupported = computed(() => {
    // iOS Safari 支持 PWA（通过"添加到主屏幕"）
    if (isIOS.value) return true

    // 检查是否有当前事件或曾经收到过事件
    const state = getStorageState()
    return !!deferredPrompt.value || state.hasSeenInstallPrompt || hasPwaSupport.value
  })

  /**
   * 是否支持通知功能
   */
  const isNotificationSupported = computed(() => {
    return typeof Notification !== 'undefined'
  })

  // 通知权限状态
  const notificationPermission = ref<NotificationPermission>(
    isNotificationSupported.value ? Notification.permission : 'default'
  )

  /**
   * 是否需要显示"开启通知"按钮
   * 条件：支持通知且权限不是 granted
   */
  const canShowNotificationButton = computed(() => {
    return isNotificationSupported.value && notificationPermission.value !== 'granted'
  })

  /**
   * 请求通知权限
   */
  const requestNotificationPermission = async () => {
    if (!isNotificationSupported.value) return

    try {
      const permission = await Notification.requestPermission()
      notificationPermission.value = permission

      if (permission === 'granted') {
        theToast.success('通知已开启')
        showNotification('通知已开启', {
          body: '您现在可以接收到来自 Ech0 的重要提醒',
          tag: 'notification-enabled',
          vibrate: [200, 100, 200],
        })
        // 尝试注册周期性同步
        registerPeriodicSync()
      } else if (permission === 'denied') {
        theToast.warning('通知权限被拒绝，请在浏览器设置中开启')
      }
    } catch (error) {
      console.error('[PWA] Failed to request notification permission:', error)
    }
  }

  /**
   * 注册周期性后台同步 (Periodic Background Sync)
   * 允许在应用未开启时在后台更新数据
   */
  const registerPeriodicSync = async () => {
    if (!('serviceWorker' in navigator)) return

    try {
      const registration = await navigator.serviceWorker.ready
      if ('periodicSync' in (registration as any)) {
        const status = await (navigator as any).permissions.query({
          name: 'periodic-background-sync',
        })

        if (status.state === 'granted') {
          await (registration as any).periodicSync.register('fetch-updates', {
            minInterval: 24 * 60 * 60 * 1000, // 浏览器通常限制最小 24 小时
          })
          console.log('[PWA] Periodic Sync registered')
        }
      }
    } catch (e) {
      console.warn('[PWA] Periodic Sync registration failed:', e)
    }
  }

  /**
   * 显示持久化通知 (通过 Service Worker)
   */
  const showNotification = async (
    title: string,
    options?: NotificationOptions & {
      url?: string
      actions?: any[] // 使用 any 避免 SW 类型冲突
      vibrate?: number[]
      data?: any
    },
  ) => {
    if (!isNotificationSupported.value || notificationPermission.value !== 'granted') return

    try {
      const registration = await navigator.serviceWorker.ready

      // 准备完整的通知选项
      const fullOptions: NotificationOptions = {
        icon: '/Ech0.png',
        badge: '/favicon.svg',
        vibrate: options?.vibrate || [100],
        ...options,
        data: {
          url: options?.url || '/',
          ...(options?.data || {}),
        },
      }

      // 1. 尝试使用 Service Worker 发送持久通知 (即使在后台也有效)
      await registration.showNotification(title, fullOptions)
    } catch (error) {
      console.warn('[PWA] Service Worker notification failed, fallback to basic:', error)
      // 2. 回退到普通前台通知
      try {
        const n = new Notification(title, options)
        if (options?.url) {
          n.onclick = () => {
            window.focus()
            router.push(options.url!)
            n.close()
          }
        }
      } catch (e) {
        console.error('[PWA] Fallback notification failed', e)
      }
    }
  }

  /**
   * 抓取指定站点的最新一条动态内容
   */
  const fetchLatestEchoContent = async (serverUrl: string): Promise<string | null> => {
    try {
      // 使用 GET 方法带参数获取最新一页的第一条，更符合规范且性能更好
      const response = await fetch(`${serverUrl}/api/echo/page?page=1&pageSize=1`, {
        method: 'GET',
        headers: { 'Accept': 'application/json' },
      })
      const json = (await response.json()) as any

      if (json?.code === 1 && json?.data?.items?.length > 0) {
        return json.data.items[0].content
      }
    } catch (e) {
      console.warn(`[PWA] Failed to fetch latest echo from ${serverUrl}`, e)
    }
    return null
  }

  /**
   * 将当前前台状态同步给 Service Worker 的 Cache Storage
   * 解决 LocalStorage 与 SW 存储不一致导致的重复通知问题
   */
  const syncStateToServiceWorker = async () => {
    if (!('serviceWorker' in navigator) || !('caches' in window)) return

    try {
      const cache = await caches.open('ech0-sync-state')

      // 构造当前状态快照
      const hubCounts: Record<string, number> = {}
      connectStore.connectsInfo.forEach(s => {
        hubCounts[s.server_url] = s.total_echos
      })

      const lastInboxId = inboxStore.unreadItems?.length > 0
        ? Math.max(...inboxStore.unreadItems.map(i => i.id))
        : 0

      const lastTodoId = todoStore.todos?.length > 0
        ? Math.max(...todoStore.todos.map(t => t.id))
        : 0

      // 获取认证 Token
      const token = localStg.getItem('token') || ''

      const state = { lastInboxId, lastTodoId, hubCounts, token }

      // 写入 Cache API (SW 周期性同步会读取此文件)
      await cache.put('/state.json', new Response(JSON.stringify(state), {
        headers: { 'Content-Type': 'application/json' }
      }))
    } catch (e) {
      console.warn('[PWA] Failed to sync state to SW cache:', e)
    }
  }

  /**
   * 清除 Service Worker 的缓存状态
   * 通常在退出登录时调用，防止 Token 泄露
   */
  const clearServiceWorkerState = async () => {
    if (!('serviceWorker' in navigator) || !('caches' in window)) return

    try {
      await caches.delete('ech0-sync-state')
      console.log('[PWA] Service Worker state cleared')
    } catch (e) {
      console.warn('[PWA] Failed to clear SW state:', e)
    }
  }

  /**
   * 计算总角标数量
   * 包含：Hub 更新、未读收件箱、待办事项
   */
  const totalBadgeCount = computed(() => {
    // 1. Hub 更新数量
    const hubCount = connectStore.hubUpdateCount || 0

    // 2. 收件箱未读数量
    const inboxCount = inboxStore.unreadItems?.length || 0

    // 3. 待办事项数量 (status === 0 表示未完成)
    const todoCount = todoStore.todos?.filter((t) => t.status === 0).length || 0

    return hubCount + inboxCount + todoCount
  })

  // ================================================================
  // 持久化状态管理
  // ================================================================

  const getStorageState = (): PwaStorageState => {
    const stored = localStg.getItem(PWA_STORAGE_KEY)
    if (stored) {
      try {
        return { ...getDefaultStorageState(), ...JSON.parse(stored as string) }
      } catch {
        return getDefaultStorageState()
      }
    }
    return getDefaultStorageState()
  }

  const saveStorageState = (state: Partial<PwaStorageState>) => {
    const current = getStorageState()
    localStg.setItem(PWA_STORAGE_KEY, JSON.stringify({ ...current, ...state }))
  }

  // ================================================================
  // 核心方法
  // ================================================================

  /**
   * 初始化 PWA Store
   * 应在应用启动时调用
   */
  const init = () => {
    // 检测设备和浏览器
    detectPlatform()

    // 检测是否已安装
    checkInstallState()

    // 检测浏览器 PWA 支持情况
    checkPwaSupport()

    // 监听 beforeinstallprompt 事件
    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt)

    // 监听安装完成事件
    window.addEventListener('appinstalled', handleAppInstalled)

    // 监听总数变化并更新角标
    watch(
      totalBadgeCount,
      (count) => {
        refreshBadge(count)
      },
      { immediate: true },
    )

    // 如果通知权限已授予，尝试注册周期性同步
    if (notificationPermission.value === 'granted') {
      registerPeriodicSync()
    }


    // 监听 Hub 更新并触发通知
    watch(
      () => connectStore.hubUpdateCount,
      async (newCount, oldCount) => {
        // 仅在更新数量增加时触发通知
        if (newCount > (oldCount || 0)) {
          const lastSiteCounts = localStg.getItem<Record<string, number>>('hubSiteCounts') || {}
          const updates = connectStore.connectsInfo.filter(
            (s) => s.total_echos > (lastSiteCounts[s.server_url] || 0),
          )

          if (updates.length > 0) {
            const first = updates[0]!
            const title = updates.length === 1 ? `✨ ${first.server_name} 更新了` : '✨ Hub 发现新动态'

            let body = ''
            if (updates.length === 1) {
              // 只有一个站点更新，尝试获取具体内容
              const latestContent = await fetchLatestEchoContent(first.server_url)
              const snippet = latestContent
                ? latestContent.length > 50
                  ? latestContent.slice(0, 50) + '...'
                  : latestContent
                : `发布了 ${first.total_echos - (lastSiteCounts[first.server_url] || 0)} 条新内容`

              body = `${snippet}\n点击查看详情`
            } else {
              // 多个站点更新，显示列表
              body = updates
                .map((s) => `• ${s.server_name} (+${s.total_echos - (lastSiteCounts[s.server_url] || 0)})`)
                .join('\n')
            }

            showNotification(title, {
              body,
              tag: 'hub-update',
              renotify: true,
              url: '/hub',
              vibrate: [50],
              data: { type: 'hub' },
            } as any)

            // 通知发送后同步状态，确保后台同步知道我们已经处理了这批更新
            syncStateToServiceWorker()
          }
        }
      },
    )

    // 监听收件箱更新并触发通知
    watch(
      () => inboxStore.unreadItems,
      (newList, oldList) => {
        // 初始加载或列表缩减时不触发通知
        if (!oldList || newList.length <= oldList.length) return

        const oldIds = new Set(oldList.map((i) => i.id))
        const added = newList.filter((i) => !oldIds.has(i.id))

        added.forEach((item) => {
          showNotification(`来自 ${item.source} 的新消息`, {
            body: item.content,
            tag: `inbox-${item.id}`,
            url: '/?mode=inbox',
            vibrate: [200, 100, 200],
            actions: [{ action: 'inbox-read', title: '设为已读' }],
            data: { type: 'inbox', inboxId: item.id },
          } as any)
        })
      },
      { deep: true },
    )

    // 监听待办事项更新并触发通知
    watch(
      () => todoStore.todos,
      (newList, oldList) => {
        const newIncomplete = newList.filter((t) => t.status === 0)
        const oldIncomplete = (oldList || []).filter((t) => t.status === 0)

        // 仅在未完成项增加时触发
        if (newIncomplete.length > oldIncomplete.length) {
          const oldIds = new Set(oldIncomplete.map((t) => t.id))
          const added = newIncomplete.filter((t) => !oldIds.has(t.id))

          added.forEach((todo) => {
            showNotification('新待办事项', {
              body: todo.content,
              tag: `todo-${todo.id}`,
              url: '/?mode=todo',
              vibrate: [100, 50, 100],
              actions: [{ action: 'todo-done', title: '完成任务' }],
              data: { type: 'todo', todoId: todo.id },
            } as any)
          })
        }
      },
      { deep: true },
    )

    // 更新访问计数
    updateVisitCount()

    const state = getStorageState()
    console.log('[PWA] Store initialized', {
      isInstalled: isInstalled.value,
      isIOS: isIOS.value,
      isSafari: isSafari.value,
      isFirefox: isFirefox.value,
      hasSeenInstallPrompt: state.hasSeenInstallPrompt,
      hasPwaSupport: hasPwaSupport.value,
      badgeCount: totalBadgeCount.value,
    })

    // 同步初始状态到 SW 缓存，防止重推
    syncStateToServiceWorker()
  }

  /**
   * 检测平台信息
   */
  const detectPlatform = () => {
    const ua = navigator.userAgent
    isIOS.value = /iPad|iPhone|iPod/.test(ua) && !(window as any).MSStream
    isSafari.value = /^((?!chrome|android).)*safari/i.test(ua)
    // Firefox 不支持 beforeinstallprompt，需要特殊处理
    isFirefox.value = /Firefox/i.test(ua) && !/Seamonkey/i.test(ua)
  }

  /**
   * 检测浏览器 PWA 支持情况
   *
   * 主要检测：
   * 1. beforeinstallprompt 事件是否存在于 window 上
   * 2. 是否曾经收到过该事件（持久化记录）
   * 3. Service Worker 是否支持
   *
   * 注意：Firefox 虽然支持 Service Worker，但不支持 beforeinstallprompt
   * 用户只能通过浏览器菜单手动安装
   */
  const checkPwaSupport = () => {
    // 检查浏览器是否支持 beforeinstallprompt
    // Chrome、Edge、Samsung Internet 等支持此事件
    hasPwaSupport.value = 'onbeforeinstallprompt' in window

    // 如果 Service Worker 不支持，则 PWA 也不支持
    if (!('serviceWorker' in navigator)) {
      hasPwaSupport.value = false
    }

    // Firefox 虽然支持 PWA（可通过菜单安装），但不支持 beforeinstallprompt
    // 所以我们不在 Firefox 中显示安装按钮，让用户使用浏览器菜单
    if (isFirefox.value) {
      hasPwaSupport.value = false
    }
  }

  /**
   * 检测安装状态
   *
   * 使用多种方法检测 PWA 是否已安装：
   * 1. display-mode 媒体查询（最可靠）
   * 2. navigator.standalone（iOS Safari）
   * 3. TWA referrer 检测
   */
  const checkInstallState = () => {
    // 方法1: 通过 display-mode 媒体查询检测
    // 检测多种可能的 PWA 显示模式
    const standaloneQuery = window.matchMedia('(display-mode: standalone)')
    const minimalUiQuery = window.matchMedia('(display-mode: minimal-ui)')
    const fullscreenQuery = window.matchMedia('(display-mode: fullscreen)')
    const wcoQuery = window.matchMedia('(display-mode: window-controls-overlay)')

    isInstalled.value =
      standaloneQuery.matches ||
      minimalUiQuery.matches ||
      fullscreenQuery.matches ||
      wcoQuery.matches

    // 监听 display-mode 变化（例如用户安装或卸载 PWA）
    const handleDisplayModeChange = () => {
      isInstalled.value =
        standaloneQuery.matches ||
        minimalUiQuery.matches ||
        fullscreenQuery.matches ||
        wcoQuery.matches
      console.log('[PWA] Display mode changed, isInstalled:', isInstalled.value)
    }

    standaloneQuery.addEventListener('change', handleDisplayModeChange)
    minimalUiQuery.addEventListener('change', handleDisplayModeChange)
    fullscreenQuery.addEventListener('change', handleDisplayModeChange)
    wcoQuery.addEventListener('change', handleDisplayModeChange)

    // 方法2: iOS Safari 特殊检测
    // navigator.standalone 是 iOS Safari 独有的属性
    if ((navigator as any).standalone === true) {
      isInstalled.value = true
    }

    // 方法3: 检查是否通过 TWA (Trusted Web Activity) 启动
    if (document.referrer.startsWith('android-app://')) {
      isInstalled.value = true
    }
  }

  /**
   * 更新访问计数
   */
  const updateVisitCount = () => {
    const state = getStorageState()
    const now = Date.now()
    const oneDay = 24 * 60 * 60 * 1000
    const oneHour = 60 * 60 * 1000

    // 如果距离上次访问超过1小时，增加访问计数
    if (now - state.lastVisit > oneHour) {
      saveStorageState({
        visitCount: state.visitCount + 1,
        lastVisit: now,
      })
    }

    // 如果用户之前关闭了提示，检查是否已过冷却期（7天）
    if (state.dismissed && now - state.dismissedAt > 7 * oneDay) {
      saveStorageState({
        dismissed: false,
        dismissedAt: 0,
      })
    }
  }

  /**
   * 刷新 PWA 角标 (App Badging API)
   *
   * @param count 数量，0 则清除
   */
  const refreshBadge = async (count: number) => {
    // 检查支持情况
    if (!('setAppBadge' in navigator)) {
      return
    }

    try {
      if (count > 0) {
        // 设置角标
        // 注意：某些平台（如 Android）可能只显示一个点，而不显示具体数字
        // 某些平台（如 macOS/Windows）可以直接显示数字
        await (navigator as any).setAppBadge(count)
        console.debug(`[PWA] App badge set to: ${count}`)
      } else {
        // 清除角标
        await (navigator as any).clearAppBadge()
        console.debug('[PWA] App badge cleared')
      }
    } catch (error) {
      console.warn('[PWA] Failed to update app badge:', error)
    }
  }

  /**
   * 处理 beforeinstallprompt 事件
   *
   * 这是 PWA 安装的核心事件
   * 当浏览器检测到网站符合 PWA 安装条件时触发
   */
  const handleBeforeInstallPrompt = (e: Event) => {
    // 阻止浏览器默认的迷你信息栏（在移动设备上）
    e.preventDefault()

    // 保存事件以便稍后使用
    deferredPrompt.value = e as BeforeInstallPromptEvent

    // 记录浏览器支持 PWA 安装（持久化）
    // 这样即使页面刷新后事件没有再次触发，我们也知道浏览器支持
    saveStorageState({ hasSeenInstallPrompt: true })
    hasPwaSupport.value = true

    console.log('[PWA] beforeinstallprompt event captured and stored')
  }

  /**
   * 处理应用安装完成事件
   */
  const handleAppInstalled = () => {
    isInstalled.value = true
    deferredPrompt.value = null
    isInstalling.value = false

    // ⭐ 关键改进：安装成功后，清除持久化的提示状态
    // 这样在普通浏览器窗口再次打开时，就不会因为之前的记忆而显示安装按钮
    saveStorageState({
      hasSeenInstallPrompt: false,
      dismissed: false,
    })

    console.log('[PWA] App installed successfully')
    theToast.success('🎉 Ech0 已安装到您的设备！')
  }

  // ================================================================
  // 安装提示相关方法
  // ================================================================

  /**
   * 检查是否应该自动显示安装提示
   * @returns 是否应该显示
   */
  const shouldShowPrompt = (): boolean => {
    // 已安装则不显示
    if (isInstalled.value) return false

    // 本次会话已显示过则不重复显示
    if (hasShownPromptThisSession.value) return false

    // 检查用户是否已关闭提示（冷却期内）
    const state = getStorageState()
    if (state.dismissed) return false

    return true
  }

  /**
   * 显示 Snackbar 风格的安装提示
   * 适用于支持 beforeinstallprompt 的浏览器（Chrome、Edge 等）
   */
  const showInstallPrompt = () => {
    if (!shouldShowPrompt()) return
    if (!deferredPrompt.value) return

    hasShownPromptThisSession.value = true

    theToast.info('安装 Ech0 到桌面，获得更流畅的体验', {
      duration: 7000,
      action: {
        label: '立即安装',
        onClick: () => installApp(true),
      },
    })
  }

  /**
   * 显示 iOS Safari 的安装引导
   * 因为 iOS 不支持 beforeinstallprompt，需要引导用户手动操作
   */
  const showIOSInstallGuide = () => {
    if (!shouldShowPrompt()) return
    if (!isIOS.value) return

    hasShownPromptThisSession.value = true

    theToast.info('点击底部「分享」按钮，选择「添加到主屏幕」即可安装', {
      duration: 10000,
      action: {
        label: '知道了',
        onClick: () => {
          saveStorageState({ dismissed: true, dismissedAt: Date.now() })
        },
      },
    })
  }

  /**
   * 智能显示安装提示
   * 根据平台自动选择合适的提示方式
   */
  const smartShowInstallPrompt = () => {
    if (!shouldShowPrompt()) return

    if (isIOS.value && !isInstalled.value) {
      showIOSInstallGuide()
    } else if (deferredPrompt.value) {
      showInstallPrompt()
    }
  }

  /**
   * 执行安装
   * 调用浏览器原生的安装流程
   *
   * @param fromSnackbar 是否从 Snackbar 提示触发（用于决定取消时是否显示引导提示）
   */
  const installApp = async (fromSnackbar = false) => {
    // 如果没有 deferredPrompt，显示手动安装说明
    if (!deferredPrompt.value) {
      showManualInstallGuide()
      return
    }

    isInstalling.value = true

    try {
      // 触发浏览器安装提示
      await deferredPrompt.value.prompt()

      // 等待用户响应
      const { outcome } = await deferredPrompt.value.userChoice

      console.log('[PWA] User choice:', outcome)

      if (outcome === 'accepted') {
        // 安装成功，appinstalled 事件会处理后续逻辑
      } else {
        // 用户取消安装
        // 只有从 Snackbar 触发的取消才显示引导提示
        if (fromSnackbar) {
          theToast.info('下次可以在设置页面找到安装入口')
        }
        saveStorageState({ dismissed: true, dismissedAt: Date.now() })
      }
    } catch (error) {
      console.error('[PWA] Install error:', error)
      theToast.error('安装出现问题，请稍后再试')
    } finally {
      isInstalling.value = false
      // deferredPrompt 只能使用一次，之后需要等待新的事件
      deferredPrompt.value = null
    }
  }

  /**
   * 显示手动安装说明
   *
   * 当 deferredPrompt 不可用时（例如页面刷新后），
   * 引导用户通过浏览器菜单手动安装
   */
  const showManualInstallGuide = () => {
    const ua = navigator.userAgent
    const isChrome = /Chrome/i.test(ua) && !/Edge/i.test(ua) && !/OPR/i.test(ua)
    const isEdge = /Edg/i.test(ua)
    const isOpera = /OPR/i.test(ua) || /Opera/i.test(ua)
    const isSamsungInternet = /SamsungBrowser/i.test(ua)

    let message = '请通过浏览器菜单安装此应用'

    if (isChrome) {
      message = '请点击浏览器地址栏右侧的安装图标，或通过菜单选择「安装 Ech0」'
    } else if (isEdge) {
      message = '请点击浏览器地址栏右侧的安装图标，或通过菜单选择「应用」→「将此站点作为应用安装」'
    } else if (isFirefox.value) {
      message = '请点击浏览器菜单，选择「安装」或「添加到主屏幕」'
    } else if (isOpera) {
      message = '请点击浏览器菜单，选择「安装 Ech0」'
    } else if (isSamsungInternet) {
      message = '请点击浏览器菜单，选择「添加页面到」→「主屏幕」'
    }

    theToast.info(message, { duration: 8000 })
  }

  // ================================================================
  // 触发时机相关方法
  // ================================================================

  /**
   * 用户登录成功后调用
   * 在用户表现出对应用的兴趣后显示提示
   */
  const onUserLoggedIn = () => {
    // 登录成功表示用户对应用有较高参与度
    // 延迟 3 秒显示，避免与登录成功提示冲突
    setTimeout(() => {
      smartShowInstallPrompt()
    }, 3000)
  }

  /**
   * 用户发布 Echo 成功后调用
   * 在用户完成核心操作后显示提示
   */
  const onEchoPublished = () => {
    const state = getStorageState()
    // 只在第一次发布时提示
    if (state.visitCount <= 1) {
      setTimeout(() => {
        smartShowInstallPrompt()
      }, 2000)
    }
  }

  /**
   * 基于访问次数的自动提示
   * 当用户访问次数达到阈值时自动显示
   */
  const checkAutoPrompt = () => {
    const state = getStorageState()
    // 访问次数达到 3 次以上时自动提示
    if (state.visitCount >= 3) {
      // 延迟 10 秒，让用户先熟悉页面内容
      setTimeout(() => {
        smartShowInstallPrompt()
      }, 10000)
    }
  }

  // ================================================================
  // 清理方法
  // ================================================================

  /**
   * 重置提示状态（用于测试或调试）
   */
  const resetPromptState = () => {
    localStg.removeItem(PWA_STORAGE_KEY)
    hasShownPromptThisSession.value = false
    console.log('[PWA] Prompt state reset')
  }

  return {
    // State
    deferredPrompt,
    isInstalled,
    isInstalling,
    isIOS,
    isSafari,
    isFirefox,
    hasPwaSupport,
    notificationPermission,

    // Computed
    canShowInstall,
    canShowInstallEntry,
    canShowIOSInstallEntry,
    isPwaSupported,
    canShowNotificationButton,

    // Methods
    init,
    showInstallPrompt,
    showIOSInstallGuide,
    showManualInstallGuide,
    smartShowInstallPrompt,
    installApp,
    requestNotificationPermission,
    showNotification,
    clearServiceWorkerState,

    // Triggers
    onUserLoggedIn,
    onEchoPublished,
    checkAutoPrompt,

    // Utils
    shouldShowPrompt,
    resetPromptState,
    syncStateToServiceWorker,
  }
})

// 监听 Service Worker 发来的消息 (用于软跳转等)
if ('serviceWorker' in navigator) {
  navigator.serviceWorker.addEventListener('message', (event) => {
    if (event.data) {
      // 1. 处理导航指令
      if (event.data.type === 'NAVIGATE') {
        const url = event.data.url
        if (url) {
          router.push(url).catch(err => {
            if (err.name !== 'NavigationDuplicated') console.warn(err)
          })
        }
      }

      // 2. 处理刷新指令 (来自 SW 后台操作)
      if (event.data.type === 'REFRESH') {
        const target = event.data.target
        if (target === 'todo') {
          // 刷新待办事项
          // 需要在此作用域内获取 store 实例
          const todoStore = useTodoStore()
          todoStore.getTodos()
        } else if (target === 'inbox') {
          // 刷新收件箱 (包括未读数)
          // 需要在此作用域内获取 store 实例
          const inboxStore = useInboxStore()
          // 假设 unreadItems 包含最新未读数据，fetchUnread 是获取方法
          inboxStore.refresh()
        }
      }
    }
  })
}

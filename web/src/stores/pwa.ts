/**
 * PWA Store - 管理 PWA 安装提示和状态
 *
 * 使用 Snackbar（临时界面）模式展示安装提示，与现有 Toast 系统保持一致
 * 支持 beforeinstallprompt 事件和 iOS 后备方案
 */
import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'

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
const PWA_STORAGE_KEY = 'ech0_pwa_prompt'

// 获取 PWA 提示状态
interface PwaPromptState {
  dismissed: boolean // 用户是否已关闭提示
  dismissedAt: number // 关闭时间戳
  visitCount: number // 访问次数
  lastVisit: number // 上次访问时间
}

const getDefaultPromptState = (): PwaPromptState => ({
  dismissed: false,
  dismissedAt: 0,
  visitCount: 0,
  lastVisit: 0,
})

export const usePwaStore = defineStore('pwaStore', () => {
  // ================================================================
  // State
  // ================================================================

  // 保存的 beforeinstallprompt 事件
  const deferredPrompt = ref<BeforeInstallPromptEvent | null>(null)

  // 是否已安装（通过 display-mode 检测）
  const isInstalled = ref(false)

  // 是否可以显示安装按钮（有 deferredPrompt 且未安装）
  const canShowInstall = computed(() => !!deferredPrompt.value && !isInstalled.value)

  // 是否正在安装中
  const isInstalling = ref(false)

  // 是否为 iOS 设备
  const isIOS = ref(false)

  // 是否为 Safari 浏览器
  const isSafari = ref(false)

  // 当前会话是否已显示过提示
  const hasShownPromptThisSession = ref(false)

  // ================================================================
  // 持久化状态管理
  // ================================================================

  const getPromptState = (): PwaPromptState => {
    const stored = localStg.getItem(PWA_STORAGE_KEY)
    if (stored) {
      try {
        return JSON.parse(stored as string)
      } catch {
        return getDefaultPromptState()
      }
    }
    return getDefaultPromptState()
  }

  const savePromptState = (state: Partial<PwaPromptState>) => {
    const current = getPromptState()
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

    // 监听 beforeinstallprompt 事件
    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt)

    // 监听安装完成事件
    window.addEventListener('appinstalled', handleAppInstalled)

    // 更新访问计数
    updateVisitCount()

    console.log('[PWA] Store initialized', {
      isInstalled: isInstalled.value,
      isIOS: isIOS.value,
      isSafari: isSafari.value,
    })
  }

  /**
   * 检测平台信息
   */
  const detectPlatform = () => {
    const ua = navigator.userAgent
    isIOS.value = /iPad|iPhone|iPod/.test(ua) && !(window as any).MSStream
    isSafari.value = /^((?!chrome|android).)*safari/i.test(ua)
  }

  /**
   * 检测安装状态
   */
  const checkInstallState = () => {
    // 方法1: 通过 display-mode 媒体查询检测
    const standaloneQuery = window.matchMedia('(display-mode: standalone)')
    isInstalled.value = standaloneQuery.matches

    // 监听 display-mode 变化
    standaloneQuery.addEventListener('change', (e) => {
      isInstalled.value = e.matches
    })

    // 方法2: iOS Safari 特殊检测
    if ((navigator as any).standalone === true) {
      isInstalled.value = true
    }
  }

  /**
   * 更新访问计数
   */
  const updateVisitCount = () => {
    const state = getPromptState()
    const now = Date.now()
    const oneDay = 24 * 60 * 60 * 1000

    // 如果距离上次访问超过1小时，增加访问计数
    if (now - state.lastVisit > 60 * 60 * 1000) {
      savePromptState({
        visitCount: state.visitCount + 1,
        lastVisit: now,
      })
    }

    // 如果用户之前关闭了提示，检查是否已过冷却期（7天）
    if (state.dismissed && now - state.dismissedAt > 7 * oneDay) {
      savePromptState({
        dismissed: false,
        dismissedAt: 0,
      })
    }
  }

  /**
   * 处理 beforeinstallprompt 事件
   */
  const handleBeforeInstallPrompt = (e: Event) => {
    // 阻止默认的迷你信息栏
    e.preventDefault()

    // 保存事件以便稍后使用
    deferredPrompt.value = e as BeforeInstallPromptEvent

    console.log('[PWA] beforeinstallprompt event captured')
  }

  /**
   * 处理应用安装完成事件
   */
  const handleAppInstalled = () => {
    isInstalled.value = true
    deferredPrompt.value = null
    isInstalling.value = false

    console.log('[PWA] App installed successfully')
    theToast.success('🎉 Ech0 已安装到您的设备！')
  }

  // ================================================================
  // 安装提示相关方法
  // ================================================================

  /**
   * 检查是否应该显示安装提示
   * @returns 是否应该显示
   */
  const shouldShowPrompt = (): boolean => {
    // 已安装则不显示
    if (isInstalled.value) return false

    // 本次会话已显示过则不重复显示
    if (hasShownPromptThisSession.value) return false

    // 检查用户是否已关闭提示（冷却期内）
    const state = getPromptState()
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
          savePromptState({ dismissed: true, dismissedAt: Date.now() })
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
   * @param fromSnackbar 是否从 Snackbar 提示触发（用于决定取消时是否显示引导提示）
   */
  const installApp = async (fromSnackbar = false) => {
    if (!deferredPrompt.value) {
      console.warn('[PWA] No deferred prompt available')
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
        savePromptState({ dismissed: true, dismissedAt: Date.now() })
      }
    } catch (error) {
      console.error('[PWA] Install error:', error)
      theToast.error('安装出现问题，请稍后再试')
    } finally {
      isInstalling.value = false
      // deferredPrompt 只能使用一次
      deferredPrompt.value = null
    }
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
    const state = getPromptState()
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
    const state = getPromptState()
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
    canShowInstall,
    isInstalling,
    isIOS,
    isSafari,

    // Methods
    init,
    showInstallPrompt,
    showIOSInstallGuide,
    smartShowInstallPrompt,
    installApp,

    // Triggers
    onUserLoggedIn,
    onEchoPublished,
    checkAutoPrompt,

    // Utils
    shouldShowPrompt,
    resetPromptState,
  }
})

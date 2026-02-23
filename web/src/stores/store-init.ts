import { useThemeStore } from './theme'
import { useUserStore } from './user'
import { useSettingStore } from './setting'
import { useTodoStore } from './todo'
import { useEchoStore } from './echo'
import { useZoneStore } from './zone'
import { useEditorStore } from './editor'
import { useInboxStore } from './inbox'
import { usePwaStore } from './pwa'
import { useConnectStore } from './connect'

export async function initStores() {
  const themeStore = useThemeStore()
  const userStore = useUserStore()
  const settingStore = useSettingStore()
  const todoStore = useTodoStore()
  const echoStore = useEchoStore()
  const zoneStore = useZoneStore()
  const editorStore = useEditorStore()
  const inboxStore = useInboxStore()
  const pwaStore = usePwaStore()
  const connectStore = useConnectStore()

  themeStore.init()
  await userStore.init()
  await settingStore.init()
  todoStore.init()
  editorStore.init()
  echoStore.init()
  zoneStore.init()
  inboxStore.init()
  pwaStore.init()

  // 全局启动 Hub 更新轮询（确保从任意路由入口进入都能正常工作）
  connectStore.startPolling()

  // PWA: 基于访问次数的自动提示检查（在所有 store 初始化完成后）
  pwaStore.checkAutoPrompt()
}

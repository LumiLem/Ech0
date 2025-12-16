import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'
import {
  fetchInboxList,
  fetchUnreadInbox,
  fetchMarkInboxRead,
  fetchDeleteInbox,
  fetchClearInbox,
} from '@/service/api'
import { theToast } from '@/utils/toast'
import { useUserStore } from './user'

export const useInboxStore = defineStore('inboxStore', () => {
  const items = ref<App.Api.Inbox.Inbox[]>([])
  const unreadItems = ref<App.Api.Inbox.Inbox[]>([])
  const total = ref(0)
  const loading = ref(false)
  const page = ref(1)
  const pageSize = ref(12)
  const search = ref('')
  const initialized = ref(false)

  const userStore = useUserStore()
  const isAdmin = computed(() => Boolean(userStore.user?.is_admin))
  const canAccess = computed(() => Boolean(userStore.isLogin && isAdmin.value))
  const hasMore = computed(() => items.value.length < total.value)
  const currentSearch = computed(() => search.value)

  function resetState() {
    items.value = []
    unreadItems.value = []
    total.value = 0
    page.value = 1
    search.value = ''
  }

  async function fetchList(options: { append?: boolean } = {}) {
    if (!canAccess.value) {
      return false
    }

    loading.value = true
    try {
      const res = await fetchInboxList({
        page: page.value,
        pageSize: pageSize.value,
        search: search.value,
      })

      if (res.code === 1) {
        const payload = res.data?.items ?? []
        total.value = res.data?.total ?? 0
        items.value = options.append ? [...items.value, ...payload] : payload
        return true
      }

      return false
    } finally {
      loading.value = false
    }
  }

  async function fetchUnread() {
    if (!canAccess.value) {
      return false
    }

    const res = await fetchUnreadInbox()
    if (res.code === 1) {
      unreadItems.value = res.data ?? []
      return true
    }
    return false
  }

  async function refresh() {
    page.value = 1
    await fetchList()
    await fetchUnread()
  }

  async function loadMore() {
    if (!canAccess.value || loading.value || !hasMore.value) {
      return
    }
    page.value += 1
    const ok = await fetchList({ append: true })
    if (!ok) {
      page.value -= 1
    }
  }

  async function updateSearch(keyword: string) {
    if (!canAccess.value) {
      search.value = ''
      return
    }
    const normalized = keyword.trim()
    if (search.value === normalized) {
      return
    }
    search.value = normalized
    page.value = 1
    await fetchList()
  }

  async function markAsRead(id: number) {
    const res = await fetchMarkInboxRead(id)
    if (res.code === 1) {
      items.value = items.value.map((item) =>
        item.id === id
          ? {
              ...item,
              read: true,
            }
          : item,
      )
      unreadItems.value = unreadItems.value.filter((item) => item.id !== id)
      return true
    }
    return false
  }

  async function deleteInbox(id: number) {
    const res = await fetchDeleteInbox(id)
    if (res.code === 1) {
      items.value = items.value.filter((item) => item.id !== id)
      unreadItems.value = unreadItems.value.filter((item) => item.id !== id)
      total.value = Math.max(0, total.value - 1)
      return true
    }
    return false
  }

  async function clearInbox() {
    const res = await fetchClearInbox()
    if (res.code === 1) {
      resetState()
      theToast.success('收件箱已清空')
      return true
    }
    return false
  }

  function init() {
    if (initialized.value) {
      return
    }
    initialized.value = true

    watch(
      [() => userStore.isLogin, () => isAdmin.value],
      ([isLogin, admin]) => {
        if (isLogin && admin) {
          void refresh()
        } else {
          resetState()
        }
      },
      { immediate: true },
    )
  }

  return {
    items,
    unreadItems,
    total,
    loading,
    hasMore,
    canAccess,
    currentSearch,
    refresh,
    loadMore,
    updateSearch,
    markAsRead,
    deleteInbox,
    clearInbox,
    init,
  }
})

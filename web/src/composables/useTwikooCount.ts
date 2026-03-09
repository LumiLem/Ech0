import { ref, nextTick } from 'vue'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

declare global {
    // @ts-nocheck
    /* eslint-disable */
    interface Window {
        twikoo: any
    }
}

const commentCaches = ref<Record<string, number>>({})
let scriptLoadPromise: Promise<void> | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null
const queuedUrls = new Set<string>()

export function useTwikooCount() {
    const { CommentSetting } = storeToRefs(useSettingStore())

    function loadScript(src: string): Promise<void> {
        if (scriptLoadPromise) return scriptLoadPromise

        scriptLoadPromise = new Promise((resolve, reject) => {
            const existingScript = document.querySelector(`script[src="${src}"]`)
            if (existingScript) {
                resolve()
                return
            }

            const script = document.createElement('script')
            script.src = src
            script.onload = () => resolve()
            script.onerror = (error) => {
                console.error('Failed to load Twikoo script:', error)
                reject(error)
            }
            document.head.appendChild(script)
        })

        return scriptLoadPromise
    }

    async function fetchCountsForUrls(urls: string[]) {
        if (!CommentSetting.value?.enable_comment || !CommentSetting.value?.comment_api) {
            return
        }

        urls.forEach(url => {
            if (commentCaches.value[url] === undefined) {
                queuedUrls.add(url)
            }
        })

        if (debounceTimer) {
            clearTimeout(debounceTimer)
        }

        debounceTimer = setTimeout(async () => {
            const urlsToFetch = Array.from(queuedUrls)
            if (urlsToFetch.length === 0) return
            queuedUrls.clear()

            try {
                await loadScript('/others/scripts/twikoo.all.min.js')
                await nextTick()

                if (!window.twikoo) {
                    throw new Error('Twikoo is not available')
                }

                const res = await window.twikoo.getCommentsCount({
                    envId: CommentSetting.value.comment_api,
                    urls: urlsToFetch,
                    includeReply: true
                })

                res.forEach((item: any) => {
                    commentCaches.value[item.url] = item.count
                })
            } catch (e) {
                console.error('Failed to fetch Twikoo counts:', e)
            }
        }, 100)
    }

    return {
        commentCaches,
        fetchCountsForUrls
    }
}

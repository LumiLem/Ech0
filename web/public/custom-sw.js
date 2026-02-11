/**
 * Custom Service Worker Scripts for Ech0
 * Included via workbox.importScripts in vite.config.ts
 */

// 1. 处理通知点击事件与 Actions
self.addEventListener('notificationclick', (event) => {
    const notification = event.notification;
    const action = event.action;
    const urlToOpen = notification.data?.url || '/';

    notification.close();

    // 辅助函数：获取带 Token 的 Headers
    const getAuthHeaders = async () => {
        try {
            const cache = await caches.open('ech0-sync-state');
            const stateRes = await cache.match('/state.json');
            if (stateRes) {
                const state = await stateRes.json();
                if (state.token) {
                    return { 'Authorization': `Bearer ${state.token}` };
                }
            }
        } catch (e) {
            console.error('[SW] Failed to get auth token', e);
        }
        return {};
    };

    // 处理具体的 Action 按钮
    if (action === 'todo-done') {
        const todoId = notification.data?.todoId;
        if (todoId) {
            event.waitUntil(
                getAuthHeaders().then(headers =>
                    fetch(`/api/todo/${todoId}`, {
                        method: 'PUT',
                        headers: { ...headers }
                    })
                )
                    .then(() => {
                        // 发送刷新指令给前端
                        clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
                            windowClients.forEach((client) => {
                                client.postMessage({ type: 'REFRESH', target: 'todo' });
                            });
                        });
                    })
                    .catch(err => console.error('Failed to mark todo done in background', err))
            );
        }
        return;
    }

    if (action === 'inbox-read') {
        const inboxId = notification.data?.inboxId;
        if (inboxId) {
            event.waitUntil(
                getAuthHeaders().then(headers =>
                    fetch(`/api/inbox/${inboxId}/read`, {
                        method: 'PUT',
                        headers: { ...headers }
                    })
                )
                    .then(() => {
                        // 发送刷新指令给前端
                        clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
                            windowClients.forEach((client) => {
                                client.postMessage({ type: 'REFRESH', target: 'inbox' });
                            });
                        });
                    })
                    .catch(err => console.error('Failed to mark inbox read in background', err))
            );
        }
        return;
    }

    // 默认行为：点击通知打开/聚焦应用
    event.waitUntil(
        clients.matchAll({ type: 'window', includeUncontrolled: true }).then(async (windowClients) => {
            // 1. 尝试寻找已打开的且属于同源的窗口
            const matchingClient = windowClients.find(client => {
                const clientUrl = new URL(client.url);
                const targetUrl = new URL(urlToOpen, self.location.origin);
                return clientUrl.origin === targetUrl.origin;
            });

            if (matchingClient) {
                // 2. 如果找到了，先聚焦
                await matchingClient.focus();

                // 3. 尝试发送消息给客户端进行 Vue Router 软跳转 (无刷新)
                return matchingClient.postMessage({
                    type: 'NAVIGATE',
                    url: urlToOpen
                });
            }

            // 4. 没找到，才打开新窗口
            if (clients.openWindow) {
                return clients.openWindow(urlToOpen);
            }
        })
    );
});

// 2. 周期性后台同步 (Periodic Background Sync)
self.addEventListener('periodicsync', (event) => {
    if (event.tag === 'fetch-updates') {
        event.waitUntil(checkUpdatesAndNotify());
    }
});

// 3. Background Sync 成功事件 (离线发帖队列同步完成)
self.addEventListener('sync', (event) => {
    if (event.tag === 'api-mutation-queue') {
        event.waitUntil(
            notifyClientsOfSync('mutation')
        );
    }
});

/**
 * 通知所有客户端同步已完成
 */
async function notifyClientsOfSync(target) {
    const windowClients = await clients.matchAll({ type: 'window', includeUncontrolled: true });
    windowClients.forEach((client) => {
        client.postMessage({
            type: 'SYNC_COMPLETE',
            target: target,
            message: '离线发布的内容已成功同步'
        });
    });

    if (windowClients.length === 0) {
        await self.registration.showNotification('📤 离线内容已同步', {
            body: '之前离线发布的内容已成功同步到服务器',
            icon: '/api/icon?s=192',
            badge: '/api/icon?s=96',
            tag: 'offline-sync-success',
            data: { url: '/' }
        });
    }
}

/**
 * 核心：聚合接口调用与通知展示
 */
async function checkUpdatesAndNotify() {
    try {
        const cache = await caches.open('ech0-sync-state');
        const stateRes = await cache.match('/state.json');
        const cachedState = stateRes ? await stateRes.json() : {};
        const token = cachedState.token || '';

        if (!token) {
            console.warn('[SW] No auth token, skipping check');
            return;
        }

        const authHeaders = { 'Authorization': `Bearer ${token}` };

        // [优化核心] 调用一站式聚合接口
        const response = await fetch('/api/pwa/aggregate', { headers: authHeaders });
        const json = await response.json();

        if (json?.code !== 1) return;

        const data = json.data;

        // 1. 设置应用角标 (App Badging API)
        if ('setAppBadge' in navigator) {
            const total = (data.inboxCount || 0) + (data.todoCount || 0) + (data.hubDiff || 0);
            if (total > 0) {
                await navigator.setAppBadge(total);
            } else {
                await navigator.clearAppBadge();
            }
        }

        // 2. 展示通知
        if (data.hasUpdate && data.notifications?.length > 0) {
            for (const note of data.notifications) {
                await self.registration.showNotification(note.title, {
                    body: note.body,
                    tag: note.tag,
                    icon: note.icon || '/api/icon?s=192',
                    badge: '/api/icon?s=96',
                    renotify: true,
                    data: note.data,
                    actions: note.tag.startsWith('inbox') ? [{ action: 'inbox-read', title: '设为已读' }] :
                        note.tag.startsWith('todo') ? [{ action: 'todo-done', title: '完成任务' }] : []
                });
            }

            // 3. 内容预取
            const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
            const isSaveData = connection && connection.saveData;

            // 如果只有 Hub 更新且非省流量模式，尝试预取
            if (!isSaveData) {
                const hubNote = data.notifications.find(n => n.tag === 'hub-update');
                if (hubNote) {
                    // 后端在 aggregate 中已经能确定哪些站更新了，这里为了简化，SW 可以预取主要 Hub 列表
                    fetch('/api/connect', { headers: authHeaders }).catch(() => { });
                }
            }
        }

    } catch (e) {
        console.error('[SW] Aggregated check failed', e);
    }
}

// 4. 处理 Web Push 推送事件
self.addEventListener('push', (event) => {
    if (event.data) {
        try {
            const data = event.data.json();
            const title = data.title || 'Ech0';
            const options = {
                body: data.body || '',
                icon: data.icon || '/api/icon?s=192',
                badge: data.badge || '/api/icon?s=96',
                tag: data.tag,
                data: data.data || {},
                actions: data.actions || [],
                renotify: data.renotify || false,
                requireInteraction: data.requireInteraction || false,
                vibrate: data.vibrate || [100]
            };
            event.waitUntil(self.registration.showNotification(title, options));
        } catch (e) {
            console.error('[SW] Push event error', e);
        }
    }
});

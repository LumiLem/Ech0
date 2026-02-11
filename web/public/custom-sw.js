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
// 注意：Workbox 的 backgroundSync 插件会自动处理队列重试
// 我们通过监听 sync 事件来通知前端刷新
self.addEventListener('sync', (event) => {
    if (event.tag === 'api-mutation-queue') {
        event.waitUntil(
            // 同步完成后通知前端（刷新所有可能受影响的内容）
            notifyClientsOfSync('mutation')
        );
    }
});

/**
 * 通知所有客户端同步已完成
 * @param {string} target - 同步目标类型：'echo', 'todo', 'mutation'(所有变更)
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

    // 可选：显示系统通知（如果用户不在前台）
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
 * 抓取指定站点的最新一条动态内容
 */
async function fetchLatestEchoContent(serverUrl) {
    try {
        const response = await fetch(`${serverUrl}/api/echo/page?page=1&pageSize=1`, {
            method: 'GET',
            headers: { 'Accept': 'application/json' },
        });
        const json = await response.json();
        if (json?.code === 1 && json?.data?.items?.length > 0) {
            return json.data.items[0].content;
        }
    } catch (e) {
        console.warn(`[SW] Failed to fetch snippet from ${serverUrl}`);
    }
    return null;
}

async function checkUpdatesAndNotify() {
    try {
        // Token 仍从 Cache Storage 读取（前端初始化时写入）
        const cache = await caches.open('ech0-sync-state');
        const stateRes = await cache.match('/state.json');
        const cachedState = stateRes ? await stateRes.json() : {};
        const token = cachedState.token || '';

        if (!token) {
            console.warn('[SW] No auth token, skipping periodic check');
            return;
        }

        const authHeaders = { 'Authorization': `Bearer ${token}` };

        // 从后端获取统一快照（唯一数据源）
        let snapshot;
        try {
            const snapRes = await fetch('/api/pwa/snapshot', { headers: authHeaders });
            const snapJson = await snapRes.json();
            snapshot = snapJson?.code === 1 ? snapJson.data : null;
        } catch (e) {
            console.warn('[SW] Failed to fetch snapshot from backend');
            return;
        }
        if (!snapshot) {
            snapshot = { lastInboxId: 0, lastTodoId: 0, lastTodoRemindAt: 0, readHubCounts: {}, notifiedHubCounts: {} };
        }
        if (!snapshot.readHubCounts) snapshot.readHubCounts = {};
        if (!snapshot.notifiedHubCounts) snapshot.notifiedHubCounts = {};

        let snapshotDirty = false;

        // 1. 检查收件箱
        try {
            const inboxRes = await fetch('/api/inbox/unread', { headers: authHeaders });
            const inboxJson = await inboxRes.json();
            if (inboxJson?.code === 1 && inboxJson.data?.length > 0) {
                const newItems = inboxJson.data.filter(item => item.id > snapshot.lastInboxId);
                for (const item of newItems) {
                    await self.registration.showNotification(`📩 来自 ${item.source} 的新消息`, {
                        body: item.content,
                        icon: '/icons/notification-inbox.png',
                        badge: '/api/icon?s=96',
                        tag: `inbox-${item.id}`,
                        data: { url: '/?mode=inbox', inboxId: item.id },
                        actions: [{ action: 'inbox-read', title: '设为已读' }]
                    });
                    snapshot.lastInboxId = Math.max(snapshot.lastInboxId, item.id);
                    snapshotDirty = true;
                }
            }
        } catch (e) {
            console.warn('[SW] Inbox check failed', e);
        }

        // 2. 检查待办事项：未完成待办每 4 小时提醒一次
        try {
            const todoRemindInterval = 4 * 60 * 60;
            const nowSec = Math.floor(Date.now() / 1000);
            const todoRes = await fetch('/api/todo', { headers: authHeaders });
            const todoJson = await todoRes.json();
            const incompleteTodos = (todoJson?.data || []).filter(t => t.status === 0);

            if (incompleteTodos.length > 0 && (nowSec - (snapshot.lastTodoRemindAt || 0) >= todoRemindInterval)) {
                for (const todo of incompleteTodos) {
                    await self.registration.showNotification('⏰ 待办事项未完成', {
                        body: todo.content,
                        icon: '/icons/notification-todo.png',
                        badge: '/api/icon?s=96',
                        tag: `todo-${todo.id}`,
                        renotify: true,
                        data: { url: '/?mode=todo', todoId: todo.id },
                        actions: [{ action: 'todo-done', title: '完成任务' }]
                    });
                }
                snapshot.lastTodoRemindAt = nowSec;
                snapshotDirty = true;
            }

            for (const t of incompleteTodos) {
                if (t.id > snapshot.lastTodoId) {
                    snapshot.lastTodoId = t.id;
                    snapshotDirty = true;
                }
            }
        } catch (e) {
            console.warn('[SW] Todo check failed', e);
        }

        // 3. 检查 Hub 更新
        try {
            const connectRes = await fetch('/api/connect', { headers: authHeaders });
            const connectJson = await connectRes.json();
            if (connectJson?.code === 1 && connectJson.data?.length > 0) {
                const allSites = connectJson.data;

                const updates = allSites.filter(s => {
                    if (!(s.server_url in snapshot.notifiedHubCounts)) return false;
                    return s.total_echos > snapshot.notifiedHubCounts[s.server_url];
                });

                if (updates.length > 0) {
                    const first = updates[0];
                    const title = updates.length === 1 ? `✨ ${first.server_name} 发布了新动态` : '✨ Hub 发现了新动态';
                    let body = '';

                    if (updates.length === 1) {
                        const content = await fetchLatestEchoContent(first.server_url);
                        body = content
                            ? (content.length > 50 ? content.slice(0, 50) + '...' : content)
                            : `发布了 ${first.total_echos - snapshot.notifiedHubCounts[first.server_url]} 条新内容`;
                    } else {
                        const totalNewEchos = updates.reduce((sum, s) => sum + (s.total_echos - snapshot.notifiedHubCounts[s.server_url]), 0);
                        if (updates.length <= 3) {
                            const names = updates.map(s => s.server_name).join('、');
                            body = `${names} 更新了 ${totalNewEchos} 条动态`;
                        } else {
                            const firstTwo = updates.slice(0, 2).map(s => s.server_name).join('、');
                            body = `${firstTwo} 等 ${updates.length} 个站点更新了 ${totalNewEchos} 条动态`;
                        }
                    }

                    let icon = '/icons/notification-hub.png';
                    if (updates.length === 1 && first.logo && first.logo.startsWith('http')) {
                        icon = first.logo;
                    }

                    await self.registration.showNotification(title, {
                        body,
                        icon,
                        badge: '/api/icon?s=96',
                        tag: 'hub-update',
                        renotify: true,
                        data: { url: '/hub' }
                    });
                }

                // 记录所有站点的当前计数（包括新站点）到通知水位线
                allSites.forEach(s => { snapshot.notifiedHubCounts[s.server_url] = s.total_echos; });
                const currentUrls = new Set(allSites.map(s => s.server_url));
                Object.keys(snapshot.notifiedHubCounts).forEach(url => {
                    if (!currentUrls.has(url)) delete snapshot.notifiedHubCounts[url];
                });
                snapshotDirty = true;
            }
        } catch (e) {
            console.warn('[SW] Hub check failed', e);
        }

        // 将更新后的快照写回后端（唯一数据源）
        if (snapshotDirty) {
            try {
                await fetch('/api/pwa/snapshot', {
                    method: 'POST',
                    headers: {
                        ...authHeaders,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(snapshot)
                });
            } catch (e) {
                console.warn('[SW] Failed to save snapshot to backend', e);
            }
        }

    } catch (e) {
        console.error('[SW] Periodic sync check failed', e);
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


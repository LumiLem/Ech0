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
        // 使用 Cache Storage 存储同步状态（SW 无权访问 localStorage）
        const cache = await caches.open('ech0-sync-state');
        const stateRes = await cache.match('/state.json');
        let state = stateRes ? await stateRes.json() : { lastInboxId: 0, lastTodoId: 0, hubCounts: {}, token: '' };

        // 构造认证头
        const authHeaders = state.token ? { 'Authorization': `Bearer ${state.token}` } : {};

        // 1. 检查收件箱 (需要认证)
        const inboxRes = await fetch('/api/inbox/unread', { headers: authHeaders });
        const inboxJson = await inboxRes.json();
        if (inboxJson?.code === 1 && inboxJson.data?.length > 0) {
            const newItems = inboxJson.data.filter(item => item.id > state.lastInboxId);
            for (const item of newItems) {
                await self.registration.showNotification(`📩 来自 ${item.source} 的新消息`, {
                    body: item.content,
                    icon: '/icons/notification-inbox.png', // 使用蓝色的 Inbox 图标
                    badge: '/api/icon?s=96',
                    tag: `inbox-${item.id}`,
                    data: { url: '/?mode=inbox', inboxId: item.id },
                    actions: [{ action: 'inbox-read', title: '设为已读' }]
                });
                state.lastInboxId = Math.max(state.lastInboxId, item.id);
            }
        }

        // 2. 检查待办事项 (需要认证)
        const todoRes = await fetch('/api/todo', { headers: authHeaders });
        const todoJson = await todoRes.json();
        const todos = todoJson?.data || [];
        if (todos.length > 0) {
            const newTodos = todos.filter(t => t.id > state.lastTodoId && t.status === 0);
            for (const todo of newTodos) {
                await self.registration.showNotification('📋 待办事项提醒', {
                    body: todo.content,
                    icon: '/icons/notification-todo.png', // 使用绿色的 Todo 图标
                    badge: '/api/icon?s=96',
                    tag: `todo-${todo.id}`,
                    data: { url: '/?mode=todo', todoId: todo.id },
                    actions: [{ action: 'todo-done', title: '完成任务' }]
                });
                state.lastTodoId = Math.max(state.lastTodoId, todo.id);
            }
        }

        // 3. 检查 Hub 更新 (部分公开，部分可能需要认证，安全起见带上)
        const connectRes = await fetch('/api/connect', { headers: authHeaders });
        const connectJson = await connectRes.json();
        if (connectJson?.code === 1 && connectJson.data?.length > 0) {
            const updates = connectJson.data.filter(s => s.total_echos > (state.hubCounts[s.server_url] || 0));

            if (updates.length > 0) {
                const first = updates[0];
                const title = updates.length === 1 ? `✨ ${first.server_name} 发布了新动态` : '✨ Hub 发现了新动态';
                let body = '';

                if (updates.length === 1) {
                    const content = await fetchLatestEchoContent(first.server_url);
                    body = content
                        ? (content.length > 50 ? content.slice(0, 50) + '...' : content)
                        : `发布了 ${first.total_echos - (state.hubCounts[first.server_url] || 0)} 条新内容`;
                } else {
                    // 多个站点更新，使用紧凑格式
                    const totalNewEchos = updates.reduce((sum, s) => sum + (s.total_echos - (state.hubCounts[s.server_url] || 0)), 0);
                    if (updates.length <= 3) {
                        // 3个及以内：列出所有站点名称
                        const names = updates.map(s => s.server_name).join('、');
                        body = `${names} 更新了 ${totalNewEchos} 条动态`;
                    } else {
                        // 超过3个：显示前两个 + "等X个站点"
                        const firstTwo = updates.slice(0, 2).map(s => s.server_name).join('、');
                        body = `${firstTwo} 等 ${updates.length} 个站点更新了 ${totalNewEchos} 条动态`;
                    }
                }

                // 动态选择图标：单站更新优先尝试使用该站点 Logo，否则使用默认 Hub 图标
                let icon = '/icons/notification-hub.png';
                if (updates.length === 1 && first.logo) {
                    // 简单的 URL 检查，如果不是 http 开头，尝试拼接（假设是当前站点的相对路径，虽然 Hub 站点通常是外部的）
                    // 注意：跨域图片在某些浏览器通知中可能无法显示，但尝试一下无妨
                    if (first.logo.startsWith('http')) {
                        icon = first.logo;
                    }
                }

                await self.registration.showNotification(title, {
                    body,
                    icon,
                    badge: '/api/icon?s=96',
                    tag: 'hub-update',
                    data: { url: '/hub' }
                });

                // 更新 Hub 计数快照
                updates.forEach(s => { state.hubCounts[s.server_url] = s.total_echos; });
            }
        }

        // 保存最新状态（保持 Token 不变）
        await cache.put('/state.json', new Response(JSON.stringify(state)));

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


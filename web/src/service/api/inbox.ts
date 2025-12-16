import { request } from '../request'

export function fetchInboxList(params: App.Api.Inbox.InboxListParams) {
  const query = new URLSearchParams({
    page: String(params.page),
    pageSize: String(params.pageSize),
  })

  const trimmedSearch = params.search?.trim()
  if (trimmedSearch) {
    query.append('search', trimmedSearch)
  }

  return request<App.Api.Inbox.InboxListResult>({
    url: `/inbox?${query.toString()}`,
    method: 'GET',
  })
}

export function fetchUnreadInbox() {
  return request<App.Api.Inbox.Inbox[]>({
    url: `/inbox/unread`,
    method: 'GET',
  })
}

export function fetchMarkInboxRead(id: number) {
  return request({
    url: `/inbox/${id}/read`,
    method: 'PUT',
  })
}

export function fetchDeleteInbox(id: number) {
  return request({
    url: `/inbox/${id}`,
    method: 'DELETE',
  })
}

export function fetchClearInbox() {
  return request({
    url: `/inbox`,
    method: 'DELETE',
  })
}

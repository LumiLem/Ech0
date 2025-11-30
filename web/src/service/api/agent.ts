import { request } from '../request'

// 获取近况总结
export function fetchGetRecent() {
  return request<string>({
    url: '/agent/recent',
    method: 'GET',
  })
}

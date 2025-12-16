import { request } from '../request'

// 上传媒体文件（图片或视频）
export function fetchUploadMedia(file: File, source?: string) {
  const formData = new FormData()
  formData.append('file', file)

  if (source) {
    formData.append('ImageSource', source)
  }

  return request<App.Api.File.ImageDto>({
    url: `/images/upload`,
    method: 'POST',
    data: formData,
  })
}

// 上传图片（向后兼容）
export const fetchUploadImage = fetchUploadMedia

// 删除媒体文件（图片或视频）
export function fetchDeleteMedia(media: App.Api.Ech0.MediaToDelete) {
  return request({
    url: `/images/delete`,
    method: 'DELETE',
    data: media,
  })
}

// 删除Image（向后兼容）
export const fetchDeleteImage = fetchDeleteMedia


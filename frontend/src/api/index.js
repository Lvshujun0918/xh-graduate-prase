import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 120000
})

// 上传 HTML 文件
export function uploadHTML(file) {
  const formData = new FormData()
  formData.append('file', file)
  return api.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 获取图片信息
export function getImageInfo(taskId) {
  return api.get(`/images/${taskId}`)
}

// 获取图片代理 URL
export function getProxyImageUrl(url) {
  return `/api/proxy-image?url=${encodeURIComponent(url)}`
}

// 下载图片
export function downloadImages(taskId, indices) {
  return api.post('/download', { taskId, indices }, {
    responseType: 'blob'
  })
}

// 获取版本信息
export function getVersion() {
  return api.get('/version')
}

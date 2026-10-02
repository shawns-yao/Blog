import { API_BASE_URL, ApiError, getAuthToken, request } from './http'

import type { ApiEnvelope } from './http'

export interface MusicStatus {
  enabled: boolean
  configured: boolean
  available: boolean
  message?: string
  maxUploadBytes: number
  maxBitRate: number
  scan?: { scanning: boolean; count: number }
}

export interface MusicUpload {
  id: string
  filename: string
  title: string
  artist: string
  size: number
  duplicate: boolean
}

export interface MusicSong {
  id: string
  title: string
  artist: string
  album: string
  duration: number
  public: boolean
}

export const getMusicStatus = () => request<MusicStatus>('/admin/music/status')
export const setMusicVisibility = (id: string, isPublic: boolean) =>
  request<MusicSong>(`/admin/music/visibility/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: { public: isPublic },
  })
export const scanMusic = () =>
  request<{ scanning: boolean; count: number }>('/admin/music/scan', { method: 'POST' })
export const getMusicCatalog = (query = '', offset = 0) =>
  request<{ songs: MusicSong[]; hasMore: boolean }>('/admin/music/catalog', {
    query: { query, offset, limit: 30 },
  })

export function uploadMusicAsset(id: string, kind: 'cover' | 'lyrics', file: File) {
  const body = new FormData()
  body.append('file', file)
  return request<{ songId: string; kind: string }>(
    `/admin/music/assets/${encodeURIComponent(id)}/${kind}`,
    { method: 'POST', body },
  )
}

export function uploadMusic(
  file: File,
  onProgress: (percent: number) => void,
  signal: AbortSignal,
) {
  return new Promise<MusicUpload>((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    const abort = () => xhr.abort()
    const cleanup = () => signal.removeEventListener('abort', abort)
    xhr.open('POST', `${API_BASE_URL}/admin/music/upload`)
    xhr.timeout = 15 * 60 * 1000
    const token = getAuthToken()
    if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`)
    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable) onProgress(Math.round((event.loaded * 100) / event.total))
    }
    xhr.onload = () => {
      cleanup()
      try {
        const body = JSON.parse(xhr.responseText) as ApiEnvelope<MusicUpload>
        if (xhr.status < 200 || xhr.status >= 300 || body.code !== 0)
          reject(new ApiError(body.msg || '音乐上传失败', { status: xhr.status, code: body.code }))
        else resolve(body.data)
      } catch {
        reject(new ApiError('音乐上传响应无法解析', { status: xhr.status }))
      }
    }
    xhr.onerror = () => {
      cleanup()
      reject(new ApiError('上传连接中断，请重新上传'))
    }
    xhr.ontimeout = () => {
      cleanup()
      reject(new ApiError('上传超时，请重新上传'))
    }
    xhr.onabort = () => {
      cleanup()
      reject(new DOMException('上传已取消', 'AbortError'))
    }
    if (signal.aborted) {
      reject(new DOMException('上传已取消', 'AbortError'))
      return
    }
    signal.addEventListener('abort', abort, { once: true })
    const body = new FormData()
    body.append('file', file)
    xhr.send(body)
  })
}

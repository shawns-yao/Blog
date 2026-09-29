import { request } from './http'

export interface RagTuning {
  chunkSize: number
  chunkOverlap: number
  indexVersion: string
  vectorTopK: number
  keywordTopK: number
  topK: number
  minSimilarity: number
  rrfK: number
  rrfVectorWeight: number
  rrfKeywordWeight: number
  rerankEnabled: boolean
  rerankCandidateTopK: number
  rerankThreshold: number
  rerankFallback: boolean
}

export interface RagChatChannel {
  name: string
  model: string
  protocol: string
  priority: number
  configured: boolean
  default: boolean
  reasoningEffort?: string
}

export interface RagSettings {
  chatChannels: RagChatChannel[]
  tuning: RagTuning
  enabled: boolean
  primaryModel: string
  fallbackModel: string
  embeddingModel: string
  rerankModel: string
  primaryConfigured: boolean
  fallbackConfigured: boolean
  embeddingConfigured: boolean
  rerankConfigured: boolean
}

export type RagIndexStatus =
  'unindexed' | 'outdated' | 'pending' | 'running' | 'ready' | 'failed' | 'excluded'

export interface RagIndexStats {
  unindexed: number
  outdated: number
  pending: number
  running: number
  ready: number
  failed: number
  excluded: number
  chunks: number
  embeddingDimension: number
}

export interface RagDocument {
  momentId: number
  title: string
  contentKind: string
  status: RagIndexStatus
  published: boolean
  chunks: number
  attempts: number
  lastError: string
  updatedAt: string | null
  indexedAt: string | null
  indexDurationMs: number | null
}

export interface RagChunk {
  seq: number
  content: string
  contextHeader: string
  kind: string
  start: number
  end: number
}

export interface RagPage<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
}

export interface RagMetrics {
  days: number
  requests: number
  answered: number
  noEvidence: number
  unavailable: number
  primaryFailures: number
  fallbackRequests: number
  rerankDegraded: number
  avgDurationMs: number | null
  avgEmbeddingMs: number | null
  avgRetrievalMs: number | null
  avgRerankMs: number | null
  avgGenerationMs: number | null
  failures: { reason: string; count: number }[]
}

export const getRagSettings = () => request<RagSettings>('/admin/rag/settings')
export const getRagIndex = () => request<RagIndexStats>('/admin/rag/index')
export const getRagMetrics = () => request<RagMetrics>('/admin/rag/metrics')
export const saveRagSettings = (tuning: RagTuning) =>
  request<RagSettings>('/admin/rag/settings', { method: 'PUT', body: tuning, retry: 0 })
export const saveRagChatPriority = (priority: string[]) =>
  request<RagSettings>('/admin/rag/chat-priority', { method: 'PUT', body: { priority }, retry: 0 })
export const reindexRag = () =>
  request<{ queued: boolean }>('/admin/rag/reindex', { method: 'POST', retry: 0 })
export const reindexRagDocument = (id: number) =>
  request<{ queued: boolean }>(`/admin/rag/documents/${id}/reindex`, { method: 'POST', retry: 0 })

export function listRagDocuments(
  params: {
    page: number
    pageSize: number
    search?: string
    status?: string
    contentKind?: string
  },
  signal?: AbortSignal,
) {
  return request<RagPage<RagDocument>>('/admin/rag/documents', { query: params, signal })
}

export function getRagChunks(id: number, page: number, signal?: AbortSignal) {
  return request<RagPage<RagChunk>>(`/admin/rag/documents/${id}/chunks`, {
    query: { page, pageSize: 10 },
    signal,
  })
}

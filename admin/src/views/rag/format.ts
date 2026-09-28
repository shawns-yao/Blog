import type { RagIndexStatus } from '@/services/rag'
import type { TagProps } from 'naive-ui'

export const statusLabels: Record<RagIndexStatus, string> = {
  unindexed: '尚未入库',
  outdated: '索引过期',
  pending: '等待索引',
  running: '索引中',
  ready: '已索引',
  failed: '索引失败',
  excluded: '不参与检索',
}

export function statusType(status: RagIndexStatus): TagProps['type'] {
  if (status === 'ready') return 'success'
  if (status === 'failed') return 'error'
  if (status === 'running' || status === 'outdated') return 'warning'
  if (status === 'pending') return 'info'
  return 'default'
}

export function duration(ms: number | null | undefined) {
  if (ms == null) return '—'
  return ms < 1000 ? `${Math.round(ms)} ms` : `${(ms / 1000).toFixed(2)} s`
}

export function dateTime(value: string | null) {
  return value ? new Date(value).toLocaleString('zh-CN', { hour12: false }) : '—'
}

export function failureLabel(reason: string) {
  const labels: Record<string, string> = {
    empty_content: '正文为空',
    oversized_atomic_block: '代码、表格或公式超过分块上限',
    embedding_unavailable: '嵌入服务不可用',
    embedding_dimension_changed: '嵌入维度发生变化',
    configuration_changed: '配置发生变化',
    source_changed: '来源内容发生变化',
    index_write_failed: '索引写入失败',
    lease_expired: '索引任务超时',
    busy: '请求并发达到上限',
    disabled: '问答未启用',
    not_configured: '模型或检索配置不可用',
    index_not_ready: '公开内容索引未就绪',
    index_unavailable: '索引服务不可用',
    retrieval_unavailable: '检索服务不可用',
    rerank_unavailable: '重排序服务不可用',
    generation_unavailable: '问答模型未返回有效回答',
  }
  return labels[reason] || reason || '—'
}

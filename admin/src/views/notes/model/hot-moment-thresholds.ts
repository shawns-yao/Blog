import type { SysConfigGroup, SysConfigItem, SysConfigTreeResponse } from '@/services/sysconfig'

export interface HotMomentThresholds {
  views: number
  likes: number
  comments: number
}

export const DEFAULT_HOT_MOMENT_THRESHOLDS: HotMomentThresholds = {
  views: 100,
  likes: 10,
  comments: 5,
}

function collectGroupItems(groups: SysConfigGroup[], result: SysConfigItem[]) {
  for (const group of groups) {
    if (group.items) result.push(...group.items)
    if (group.children) collectGroupItems(group.children, result)
  }
}

function getAllItems(tree: SysConfigTreeResponse) {
  const result = [...(tree.items ?? [])]
  collectGroupItems(tree.groups ?? [], result)
  return result
}

function parseThreshold(value: unknown, fallback: number) {
  if (value === null || value === undefined || String(value).trim() === '') return fallback
  const parsed = typeof value === 'number' ? value : Number(String(value ?? '').trim())
  return Number.isSafeInteger(parsed) && parsed >= 0 ? parsed : fallback
}

export function resolveHotMomentThresholds(tree: SysConfigTreeResponse): HotMomentThresholds {
  const items = new Map(getAllItems(tree).map((item) => [item.key, item]))
  const read = (key: string, fallback: number) => {
    const item = items.get(key)
    return parseThreshold(item?.value ?? item?.defaultValue, fallback)
  }

  return {
    views: read('moment.hot.views', DEFAULT_HOT_MOMENT_THRESHOLDS.views),
    likes: read('moment.hot.likes', DEFAULT_HOT_MOMENT_THRESHOLDS.likes),
    comments: read('moment.hot.comments', DEFAULT_HOT_MOMENT_THRESHOLDS.comments),
  }
}

export function formatHotMomentThresholds(thresholds: HotMomentThresholds) {
  return `热门标准：浏览量 ≥ ${thresholds.views}、点赞数 ≥ ${thresholds.likes} 或评论数 ≥ ${thresholds.comments}`
}

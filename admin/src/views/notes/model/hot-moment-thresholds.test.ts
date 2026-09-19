import { describe, expect, it } from 'vitest'

import {
  DEFAULT_HOT_MOMENT_THRESHOLDS,
  formatHotMomentThresholds,
  resolveHotMomentThresholds,
} from './hot-moment-thresholds'

import type { SysConfigItem, SysConfigTreeResponse } from '@/services/sysconfig'

function configItem(key: string, value: unknown): SysConfigItem {
  return {
    key,
    value,
    valueType: 'number',
    enumOptions: [],
    visibleWhen: [],
    sort: 0,
    meta: {},
    isSensitive: false,
    createdAt: '',
    updatedAt: '',
  }
}

describe('hot moment thresholds', () => {
  it('reads current values from root and nested config items', () => {
    const tree: SysConfigTreeResponse = {
      items: [configItem('moment.hot.views', '320')],
      groups: [
        {
          key: 'moment',
          path: 'moment',
          label: '手记',
          children: [
            {
              key: 'hot',
              path: 'moment/hot',
              label: '热门手记',
              items: [configItem('moment.hot.likes', 24), configItem('moment.hot.comments', '8')],
            },
          ],
        },
      ],
    }

    expect(resolveHotMomentThresholds(tree)).toEqual({ views: 320, likes: 24, comments: 8 })
  })

  it('falls back to backend defaults for missing or invalid values', () => {
    expect(
      resolveHotMomentThresholds({
        groups: [],
        items: [configItem('moment.hot.views', 'invalid')],
      }),
    ).toEqual(DEFAULT_HOT_MOMENT_THRESHOLDS)
  })

  it('formats all three configured conditions with the actual inclusive comparison', () => {
    expect(formatHotMomentThresholds({ views: 100, likes: 10, comments: 5 })).toBe(
      '热门标准：浏览量 ≥ 100、点赞数 ≥ 10 或评论数 ≥ 5',
    )
  })
})

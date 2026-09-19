import { onMounted, shallowRef } from 'vue'

import { listSysConfigs } from '@/services/sysconfig'

import {
  formatHotMomentThresholds,
  resolveHotMomentThresholds,
} from '../model/hot-moment-thresholds'

const HOT_MOMENT_CONFIG_KEYS = ['moment.hot.views', 'moment.hot.likes', 'moment.hot.comments']

export function useHotMomentThresholds() {
  const description = shallowRef('正在读取热门手记配置…')

  async function load() {
    try {
      const tree = await listSysConfigs(HOT_MOMENT_CONFIG_KEYS)
      description.value = formatHotMomentThresholds(resolveHotMomentThresholds(tree))
    } catch {
      description.value = '当前手记已满足后台配置的热门标准'
    }
  }

  onMounted(load)

  return { description }
}

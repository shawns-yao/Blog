/**
 * Applies enabled feature configs to the backend.
 * Used by the init wizard to persist the chosen feature toggles.
 */
import {
  updateFederationConfigs,
  updateActivityPubConfigs,
  updateSysConfigs,
} from '@/services/sysconfig'

import type { FeatureToggleGroup } from './registry'
import type { SysConfigUpdateItem } from '@/services/sysconfig'

type Endpoint = 'federation' | 'activitypub' | 'sysconfig'

const updaters: Record<Endpoint, (items: SysConfigUpdateItem[]) => Promise<unknown>> = {
  federation: updateFederationConfigs,
  activitypub: updateActivityPubConfigs,
  sysconfig: updateSysConfigs,
}

export async function applyEnabledFeatures(
  groups: FeatureToggleGroup[],
  states: Record<string, boolean>,
  sitePublicUrl?: string,
) {
  // Group config writes by endpoint
  const batches = new Map<Endpoint, SysConfigUpdateItem[]>()

  for (const group of groups) {
    for (const feature of group.features) {
      if (!states[feature.id]) continue

      for (const cfg of feature.configs) {
        const list = batches.get(cfg.endpoint) ?? []
        list.push({ key: cfg.key, value: cfg.enableValue })
        batches.set(cfg.endpoint, list)
      }

      // Auto-fill instanceURL from public_url
      if (feature.autoFillInstanceURL && sitePublicUrl) {
        const { key, endpoint } = feature.autoFillInstanceURL
        const list = batches.get(endpoint) ?? []
        list.push({ key, value: sitePublicUrl })
        batches.set(endpoint, list)
      }
    }
  }

  // Execute all batches in parallel
  const tasks = Array.from(batches.entries()).map(([endpoint, items]) => updaters[endpoint](items))
  if (tasks.length > 0) {
    await Promise.all(tasks)
  }
}

/**
 * Feature Toggle Registry
 *
 * Declares the optional features the init wizard offers as toggles.
 * Each group lists its features together with the sys_config keys to write
 * when the feature is enabled; the UI renders them dynamically.
 */

export interface FeatureToggleConfig {
  /** sys_config key, e.g. "federation.enabled" */
  key: string
  /** Which config endpoint to call */
  endpoint: 'federation' | 'activitypub' | 'sysconfig'
  /** Value to write when the feature is toggled on */
  enableValue: unknown
}

export interface FeatureToggle {
  /** Unique id within this group */
  id: string
  /** Iconify class name */
  icon: string
  /** Display label */
  label: string
  /** Short description shown below the label */
  description: string
  /** Config keys to set when the user enables this feature */
  configs: FeatureToggleConfig[]
  /**
   * If set, also write the site's public_url into this config key
   * when the feature is enabled (useful for instanceURL fields).
   */
  autoFillInstanceURL?: {
    key: string
    endpoint: 'federation' | 'activitypub' | 'sysconfig'
  }
}

export interface FeatureToggleGroup {
  /** Stable identifier for the group */
  id: string
  /** Version the group belongs to, e.g. "2.1" */
  version: string
  /** Tag shown above the title, e.g. "v2.1 新功能" */
  tag: string
  /** Section heading */
  title: string
  /** Paragraph below the heading */
  description: string
  /** Info alert text shown below the feature list */
  hint: string
  /** Toggleable features */
  features: FeatureToggle[]
}

// ---------------------------------------------------------------------------
// REGISTRY — add new groups here
// ---------------------------------------------------------------------------

export const featureToggleRegistry: FeatureToggleGroup[] = [
  {
    id: '2.1-overview-and-features',
    version: '2.1',
    tag: 'v2.1 新功能',
    title: '联合与互联',
    description: '选择是否为您的站点启用联合功能。这些选项可以随时在设置中更改。',
    hint: '联合功能开启后，系统会自动生成签名密钥，更多高级选项可在「设置 > 联合」中配置。所有选项均可在设置中随时更改。',
    features: [
      {
        id: 'federation',
        icon: 'ph--circles-three',
        label: '站点联合',
        description:
          '启用后，您的站点可以与其他博客实例或支持联合协议的博客系统建立连接，互相交换友链申请、手记引用和提及通知。',
        configs: [{ key: 'federation.enabled', endpoint: 'federation', enableValue: true }],
        autoFillInstanceURL: { key: 'federation.instanceURL', endpoint: 'federation' },
      },
      {
        id: 'activitypub',
        icon: 'ph--broadcast',
        label: 'ActivityPub',
        description:
          '启用后，Mastodon、Misskey 等 Fediverse 平台的用户可以直接搜索并关注您的站点，新发布的手记将自动推送到他们的时间线。',
        configs: [{ key: 'activitypub.enabled', endpoint: 'activitypub', enableValue: true }],
        autoFillInstanceURL: { key: 'activitypub.instanceURL', endpoint: 'activitypub' },
      },
    ],
  },
]

/** Return all toggle groups — used by the init wizard to show every feature. */
export function getFeatureToggleGroups(): FeatureToggleGroup[] {
  return featureToggleRegistry
}

import type { MenuMixedOptions } from './interface'

export const routeRecordRaw: MenuMixedOptions[] = [
  {
    path: 'dashboard',
    name: 'dashboard',
    icon: 'icon-[mage--dashboard-chart]',
    label: '仪表板',
    meta: {
      componentName: 'Dashboard',
      pinned: true,
      showTab: true,
    },
    component: 'dashboard/index',
  },
  {
    path: 'notes',
    name: 'noteManagement',
    icon: 'iconify ph--aperture-thin',
    label: '手记管理',
    redirect: 'notes/list',
    children: [
      {
        path: 'list',
        name: 'noteList',
        label: '手记列表',
        icon: 'iconify ph--note',
        meta: {
          componentName: 'NoteList',
          showTab: true,
        },
        component: 'notes/index',
      },
      {
        path: 'edit/new',
        name: 'noteCreate',
        label: '新建手记',
        icon: 'iconify ph--pencil-simple-line',
        meta: {
          componentName: 'NoteEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle() {
            return '新建手记'
          },
        },
        component: 'notes/edit',
      },
      {
        path: 'edit/:id',
        name: 'noteEdit',
        label: '编辑手记',
        icon: 'iconify ph--pencil-simple-line',
        show: false,
        meta: {
          componentName: 'NoteEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle({ id }) {
            return `编辑手记${id ? `-${id}` : ''}`
          },
        },
        component: 'notes/edit',
      },
    ],
  },
  {
    path: 'albums',
    name: 'albumManagement',
    icon: 'iconify ph--image',
    label: '相册管理',
    redirect: 'albums/list',
    children: [
      {
        path: 'list',
        name: 'albumList',
        label: '相册列表',
        icon: 'iconify ph--images',
        meta: {
          componentName: 'AlbumList',
          showTab: true,
        },
        component: 'albums/index',
      },
      {
        path: 'edit/new',
        name: 'albumCreate',
        label: '新建相册',
        icon: 'iconify ph--pencil-simple-line',
        meta: {
          componentName: 'AlbumEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle() {
            return '新建相册'
          },
        },
        component: 'albums/edit',
      },
      {
        path: 'edit/:id',
        name: 'albumEdit',
        label: '编辑相册',
        show: false,
        meta: {
          componentName: 'AlbumEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle({ id }) {
            return `编辑相册${id ? `-${id}` : ''}`
          },
        },
        component: 'albums/edit',
      },
    ],
  },
  {
    path: 'comments',
    name: 'commentManagement',
    icon: 'iconify ph--chat-circle-text',
    label: '评论管理',
    meta: {
      componentName: 'CommentList',
      showTab: true,
    },
    component: 'comments/index',
  },
  {
    path: 'taxonomy',
    name: 'taxonomyManagement',
    icon: 'iconify ph--tree-structure',
    label: '内容分类',
    redirect: 'taxonomy/columns',
    children: [
      {
        path: 'columns',
        name: 'noteColumnManagement',
        icon: 'iconify ph--rows',
        label: '手记专栏',
        meta: {
          componentName: 'MomentColumnManagement',
          showTab: true,
        },
        component: 'taxonomy/columns/index',
      },
      {
        path: 'tags',
        name: 'tagManagement',
        icon: 'iconify ph--tag',
        label: '标签管理',
        meta: {
          componentName: 'TagManagement',
          showTab: true,
        },
        component: 'taxonomy/tags/index',
      },
    ],
  },
  {
    path: 'audience',
    name: 'audienceManagement',
    icon: 'iconify ph--users-three',
    label: '用户与访客',
    redirect: 'audience/users',
    children: [
      {
        path: 'users',
        name: 'siteUserManagement',
        icon: 'iconify ph--users',
        label: '本站用户',
        meta: {
          componentName: 'SiteUserManagement',
          showTab: true,
        },
        component: 'users/index',
      },
      {
        path: 'visitors',
        name: 'visitorProfileManagement',
        icon: 'iconify ph--users-three',
        label: '访客画像',
        meta: {
          componentName: 'VisitorProfileList',
          showTab: true,
        },
        component: 'visitors/index',
      },
      {
        path: 'rss',
        name: 'rssAccessStats',
        icon: 'iconify ph--rss',
        label: 'RSS访问统计',
        meta: {
          componentName: 'RssAccessStats',
          showTab: true,
        },
        component: 'rss/index',
      },
    ],
  },
  {
    path: 'friend-links',
    name: 'friendLinkManagement',
    icon: 'iconify ph--link',
    label: '友链',
    redirect: 'friend-links/list',
    children: [
      {
        path: 'list',
        name: 'friendLinkList',
        label: '友链列表',
        icon: 'iconify ph--link',
        meta: {
          componentName: 'FriendLinkList',
          showTab: true,
        },
        component: 'friend-links/index',
      },
      {
        path: 'applications',
        name: 'friendLinkApplications',
        label: '申请审核',
        icon: 'iconify ph--checks',
        meta: {
          componentName: 'FriendLinkApplications',
          showTab: true,
        },
        component: 'friend-links/applications',
      },
      {
        path: 'sync-jobs',
        name: 'friendLinkSyncJobs',
        label: '同步任务',
        icon: 'iconify ph--clock-counter-clockwise',
        meta: {
          componentName: 'FriendLinkSyncJobs',
          showTab: true,
        },
        component: 'friend-links/sync-jobs',
      },
    ],
  },
  {
    path: 'federation',
    name: 'unionManagement',
    icon: 'iconify ph--circles-three',
    label: '联合',
    redirect: 'federation/instances',
    children: [
      {
        path: 'instances',
        name: 'federationInstances',
        label: '联合实例',
        icon: 'iconify ph--network',
        meta: {
          componentName: 'FederationInstances',
          showTab: true,
        },
        component: 'federation/instances/index',
      },
      {
        path: 'outbound',
        name: 'federationOutbound',
        label: '出站记录',
        icon: 'iconify ph--paper-plane-tilt',
        meta: {
          componentName: 'FederationOutbound',
          showTab: true,
        },
        component: 'federation/outbound/index',
      },
      {
        path: 'activitypub-outbox',
        name: 'activityPubOutbox',
        label: 'ActivityPub 出站',
        icon: 'iconify ph--broadcast',
        meta: {
          componentName: 'ActivityPubOutbox',
          showTab: true,
        },
        component: 'federation/activitypub-outbox/index',
      },
      {
        path: 'reviews',
        name: 'federationReviews',
        label: '审核队列',
        icon: 'iconify ph--check-square',
        meta: {
          componentName: 'FederationReviews',
          showTab: true,
        },
        component: 'federation/reviews/index',
      },
      {
        path: 'debug',
        name: 'federationDebug',
        label: '联合调试',
        icon: 'iconify ph--bug',
        show: false, // Hidden from menu, accessed via Instances page
        meta: {
          componentName: 'FederationDebug',
          showTab: true,
        },
        component: 'federation/debug/OutboundRequest',
      },
    ],
  },
  // Legacy redirects for federation settings routes
  {
    path: 'federation/settings',
    name: 'unionSettingsLegacy',
    show: false,
    redirect: '/settings?tab=federation',
  },
  {
    path: 'federation/activitypub-settings',
    name: 'activityPubSettingsLegacy',
    show: false,
    redirect: '/settings?tab=federation',
  },
  {
    path: 'notifications',
    name: 'adminNotificationList',
    label: '通知中心',
    show: false,
    meta: {
      componentName: 'AdminNotificationList',
      showTab: true,
    },
    component: 'admin-notifications/index',
  },
  {
    path: 'files',
    name: 'fileManagement',
    icon: 'icon-[fluent--cloud-arrow-up-24-regular]',
    label: '文件管理',
    redirect: 'files/list',
    children: [
      {
        path: 'list',
        name: 'fileList',
        label: '文件列表',
        icon: 'icon-[fluent--cloud-arrow-up-24-regular]',
        meta: {
          componentName: 'FileList',
          showTab: true,
        },
        component: 'uploads/index',
      },
    ],
  },
  {
    path: 'webhooks',
    name: 'webhookList',
    icon: 'iconify ph--webhooks-logo',
    label: 'Webhook',
    meta: {
      componentName: 'WebhookList',
      showTab: true,
    },
    component: 'webhooks/index',
  },
  {
    path: 'global-notifications',
    name: 'globalNotificationList',
    icon: 'iconify ph--megaphone',
    label: '全站通知',
    meta: {
      componentName: 'GlobalNotificationList',
      showTab: true,
    },
    component: 'global-notifications/index',
  },
  {
    path: 'ai',
    name: 'aiManagement',
    icon: 'iconify ph--brain',
    label: 'AI',
    redirect: 'ai/task-logs',
    children: [
      {
        path: 'task-logs',
        name: 'aiTaskLogs',
        label: '任务日志',
        icon: 'iconify ph--list-checks',
        meta: {
          componentName: 'AITaskLogs',
          showTab: true,
        },
        component: 'ai/tasks/index',
      },
    ],
  },
  {
    path: 'email',
    name: 'emailManagement',
    icon: 'iconify ph--envelope',
    label: '邮件管理',
    redirect: 'email/templates',
    children: [
      {
        path: 'templates',
        name: 'emailTemplateList',
        label: '邮件模版',
        icon: 'iconify ph--scroll',
        meta: {
          componentName: 'EmailTemplateList',
          showTab: true,
        },
        component: 'email/templates/index',
      },
      {
        path: 'templates/new',
        name: 'emailTemplateCreate',
        label: '新建模版',
        show: false,
        meta: {
          componentName: 'EmailTemplateEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle() {
            return '新建模版'
          },
        },
        component: 'email/templates/edit',
      },
      {
        path: 'templates/:code',
        name: 'emailTemplateEdit',
        label: '编辑模版',
        show: false,
        meta: {
          componentName: 'EmailTemplateEdit',
          showTab: true,
          enableMultiTab: true,
          renderTabTitle({ code }) {
            return `编辑模版-${code}`
          },
        },
        component: 'email/templates/edit',
      },
      {
        path: 'subscriptions',
        name: 'emailSubscriptionList',
        label: '订阅管理',
        icon: 'iconify ph--users',
        meta: {
          componentName: 'EmailSubscriptionList',
          showTab: true,
        },
        component: 'email/subscriptions/index',
      },
      {
        path: 'outbox',
        name: 'emailOutbox',
        label: '出站队列',
        icon: 'iconify ph--paper-plane-right',
        meta: {
          componentName: 'EmailOutbox',
          showTab: true,
        },
        component: 'email/outbox/index',
      },
      {
        path: 'test',
        name: 'emailTest',
        label: '邮件测试',
        icon: 'iconify ph--paper-plane-tilt',
        meta: {
          componentName: 'EmailTest',
          showTab: true,
        },
        component: 'email/test/index',
      },
    ],
  },
  {
    path: 'settings',
    name: 'settings',
    icon: 'iconify ph--gear-six',
    label: '设置',
    meta: {
      componentName: 'UnifiedSettings',
      showTab: true,
    },
    component: 'settings/unified/index',
  },
  // Legacy redirects for old settings routes
  {
    path: 'settings/site-info',
    name: 'siteInfoLegacy',
    show: false,
    redirect: '/settings?tab=site-info',
  },
  {
    path: 'settings/login-methods',
    name: 'loginMethodsLegacy',
    show: false,
    redirect: '/settings?tab=security',
  },
  {
    path: 'settings/api-tokens',
    name: 'adminTokensLegacy',
    show: false,
    redirect: '/settings?tab=api-tokens',
  },
  {
    path: 'settings/system',
    name: 'systemSettingsLegacy',
    show: false,
    redirect: '/settings?tab=advanced',
  },
  {
    path: 'monitoring',
    name: 'systemMonitor',
    icon: 'iconify ph--activity',
    label: '系统监控',
    redirect: 'monitoring/overview',
    children: [
      {
        path: 'overview',
        name: 'systemMonitorOverview',
        label: '系统监控',
        icon: 'iconify ph--activity',
        meta: {
          componentName: 'SystemMonitor',
          showTab: true,
        },
        component: 'monitoring/index',
      },
      {
        path: 'logs',
        name: 'systemLogs',
        label: '系统日志',
        icon: 'iconify ph--scroll',
        meta: {
          componentName: 'SystemLogs',
          showTab: true,
        },
        component: 'monitoring/logs',
      },
      {
        path: 'advanced',
        name: 'advancedInfo',
        label: '系统信息',
        icon: 'iconify ph--info',
        meta: {
          componentName: 'AdvancedInfo',
          showTab: true,
        },
        component: 'advanced/index',
      },
      {
        path: 'about',
        name: 'about',
        label: '关于',
        icon: 'iconify ph--info',
        meta: {
          showTab: true,
        },
        component: 'about/index',
      },
    ],
  },
  {
    path: 'user-center',
    name: 'userCenter',
    label: '个人中心',
    icon: 'iconify ph--user',
    show: false,
    meta: {
      componentName: 'UserCenter',
      showTab: true,
    },
    component: 'user-center/index',
  },
]

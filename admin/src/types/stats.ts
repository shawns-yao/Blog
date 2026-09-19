export interface OverviewStats {
  users: number
  momentsTotal: number
  momentsPublished: number
  momentsDraft: number
  columnsTotal: number
  tagsTotal: number
}

export interface InteractionStats {
  viewsTotal: number
  likesTotal: number
  commentsTotal: number

  momentViews: number
  momentLikes: number
  momentComments: number
}

export interface WordCountStats {
  total: number
  moments: number
}

export interface PendingStats {
  unviewedComments: number
  friendLinkApplications: number
}

export interface PublishTrendPoint {
  date: string
  moments: number
}

export interface DayCountPoint {
  date: string
  count: number
}

export interface OnlineTrendPoint {
  hour: string
  peak: number
  avg: number
}

export interface DistributionItem {
  name: string
  count: number
}

export interface TopMomentItem {
  id: number
  title: string
  shortUrl: string
  views: number
  likes: number
  comments: number
  score: number
  createdAt: string
}

export interface DashboardStats {
  generatedAt: string
  cached: boolean
  overview: OverviewStats
  interaction: InteractionStats
  words: WordCountStats
  pending: PendingStats
  trend: PublishTrendPoint[]
  viewTrend: DayCountPoint[]
  commentTrend: DayCountPoint[]
  online24h: OnlineTrendPoint[]
  currentOnline: number
  todayPeakOnline: number
  columns: DistributionItem[]
  tagTop: DistributionItem[]
  platformTop: DistributionItem[]
  browserTop: DistributionItem[]
  locationTop: DistributionItem[]
  topMoments: TopMomentItem[]
}

export interface Hitokoto {
  id: number
  uuid: string
  hitokoto: string
  from: string
  from_who: string | null
  creator: string
  creator_uid: number
  reviewer: number
  commit_from: string
  created_at: string
  length: number
}

export interface HitokotoResponse {
  sentence: Hitokoto
  cached: boolean
}

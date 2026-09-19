export type TimelineMoment = {
	title: string;
	shortUrl: string;
	url: string;
	image?: string;
	publishedAt: string;
};

export type TimelineYearSummary = TimelineMoment;

export type TimelineYearData = {
	yearSummary?: TimelineYearSummary;
	moments: TimelineMoment[];
};

export type TimelineByYearResponse = Record<string, TimelineYearData>;

export type TimelineItemType = 'moment' | 'yearSummary';

export type TimelineStats = {
	moments: number;
};

export type UnifiedTimelineItem = {
	id: string;
	type: TimelineItemType;
	title?: string;
	content?: string;
	url: string;
	image?: string;
	publishedAt: Date;
	year: string;
	// Layout properties calculated at runtime
	targetX?: number;
	targetY?: number;
	monthIndex?: number;
};

export type MobileTimelineEntry = {
	item: UnifiedTimelineItem;
	side: 'left' | 'right';
};

export type MobileTimelineMonth = {
	month: number;
	stats: TimelineStats;
	entries: MobileTimelineEntry[];
};

export type MobileTimelineYear = {
	year: string;
	stats: TimelineStats;
	summary?: UnifiedTimelineItem;
	months: MobileTimelineMonth[];
};

import type { MomentDetail, MomentSummary } from './types';

type PreviewEntry = Pick<
	MomentSummary,
	'title' | 'shortUrl' | 'summary' | 'topics' | 'views' | 'likes' | 'comments'
> & {
	createdAt: string;
	cover?: string;
};

const entries: PreviewEntry[] = [
	{
		title: '九点半，窗边的光刚刚好',
		shortUrl: 'layout-preview-morning',
		summary: '把咖啡放到手边，读完昨晚折角的那一页。窗外很亮，房间里安静得只剩翻书声。',
		topics: ['日常', '阅读'],
		createdAt: '2026-09-22T09:30:00+08:00',
		views: 128,
		likes: 18,
		comments: 3
	},
	{
		title: '把读到一半的书留在桌上',
		shortUrl: 'layout-preview-afternoon',
		summary: '有些内容不必急着读完。停在恰好的地方，等下一次回到桌前，故事仍会从书签旁边继续。',
		topics: ['片刻', '书房'],
		createdAt: '2026-09-22T15:20:00+08:00',
		views: 96,
		likes: 12,
		comments: 1
	},
	{
		title: '台灯亮起以后',
		shortUrl: 'layout-preview-evening',
		summary: '天色暗得很慢，笔尖停在纸面上，刚好听见窗外第一声晚风。',
		topics: ['夜晚', '随想'],
		createdAt: '2026-09-22T20:10:00+08:00',
		views: 73,
		likes: 9,
		comments: 0
	},
	{
		title: '月亮越过窗框',
		shortUrl: 'layout-preview-moon',
		summary: '写下今天最后一句话，把书页轻轻压平。',
		topics: ['夜晚'],
		createdAt: '2026-09-22T23:06:00+08:00',
		views: 51,
		likes: 7,
		comments: 0
	},
	{
		title: '秋天好像真的来了',
		shortUrl: 'layout-preview-autumn',
		summary: '傍晚的风开始有了凉意，楼下的梧桐叶也变黄了。',
		topics: ['日常'],
		createdAt: '2026-09-23T08:45:00+08:00',
		views: 81,
		likes: 11,
		comments: 2
	},
	{
		title: '午后的一小段空白',
		shortUrl: 'layout-preview-blank',
		summary: '没有安排的十分钟，也值得被单独留下。',
		topics: ['片刻'],
		createdAt: '2026-09-23T13:20:00+08:00',
		views: 42,
		likes: 6,
		comments: 0
	},
	{
		title: '钢笔应该放在右手边',
		shortUrl: 'layout-preview-pen',
		summary: '整理桌面时突然发现，熟悉的位置也有自己的秩序。',
		topics: ['日常', '书房'],
		createdAt: '2026-09-23T16:10:00+08:00',
		views: 39,
		likes: 5,
		comments: 0
	},
	{
		title: '晚饭后的短散步',
		shortUrl: 'layout-preview-walk',
		summary: '绕着街角走了一圈，风里已经有桂花的味道。',
		topics: ['日常'],
		createdAt: '2026-09-23T18:35:00+08:00',
		views: 31,
		likes: 4,
		comments: 0
	},
	{
		title: '给明天留一张便签',
		shortUrl: 'layout-preview-note',
		summary: '先写下最重要的一件事，其他的等太阳升起来再说。',
		topics: ['计划'],
		createdAt: '2026-09-23T20:40:00+08:00',
		views: 28,
		likes: 3,
		comments: 0
	},
	{
		title: '听完一首旧歌',
		shortUrl: 'layout-preview-song',
		summary: '熟悉的旋律经过很多年，还是会把人带回同一扇窗前。',
		topics: ['片刻', '音乐'],
		createdAt: '2026-09-23T22:05:00+08:00',
		views: 24,
		likes: 3,
		comments: 0
	},
	{
		title: '今天写到这里',
		shortUrl: 'layout-preview-goodnight',
		summary: '合上电脑之前，再看一眼窗外安静的月亮。',
		topics: ['夜晚'],
		createdAt: '2026-09-23T23:18:00+08:00',
		views: 19,
		likes: 2,
		comments: 0
	}
];

export const layoutPreviewMoments: MomentSummary[] = entries.map((entry, index) => ({
	...entry,
	id: -index - 1,
	isTop: index === 0,
	isHot: false,
	isOriginal: true,
	contentUpdatedAt: entry.createdAt,
	updatedAt: entry.createdAt
}));

export function getLayoutPreviewDetail(slug: string): MomentDetail | null {
	const index = layoutPreviewMoments.findIndex((item) => item.shortUrl === slug);
	if (index < 0) return null;
	const item = layoutPreviewMoments[index];
	return {
		...item,
		contentKind: 'note',
		content: item.summary,
		contentHash: '',
		authorId: 0,
		isPublished: false,
		topics: item.topics.map((name, topicIndex) => ({ id: topicIndex, name }))
	};
}

import { describe, expect, it } from 'vitest';
import type { TimelineByYearResponse, UnifiedTimelineItem } from './types';
import { flattenAndLayoutTimeline, groupTimelineForMobile } from './utils';

const item = (
	id: string,
	type: UnifiedTimelineItem['type'],
	publishedAt: string,
	year = publishedAt.slice(0, 4)
): UnifiedTimelineItem => ({
	id,
	type,
	url: `/${id}/`,
	publishedAt: new Date(publishedAt),
	year
});

describe('groupTimelineForMobile', () => {
	it('groups entries by year and month while keeping summaries outside the alternating flow', () => {
		const years = groupTimelineForMobile([
			item('summary-2024', 'yearSummary', '2024-12-31T00:00:00Z'),
			item('moment-1', 'moment', '2024-01-03T00:00:00Z'),
			item('moment-2', 'moment', '2024-02-03T00:00:00Z'),
			item('moment-3', 'moment', '2025-01-01T00:00:00Z')
		]);

		expect(years.map((year) => year.year)).toEqual(['2024', '2025']);
		expect(years[0].summary?.id).toBe('summary-2024');
		expect(years[0].months.map((month) => month.month)).toEqual([1, 2]);
		expect(years[0].months.flatMap((month) => month.entries.map((entry) => entry.side))).toEqual([
			'left',
			'right'
		]);
		expect(years[0].stats).toEqual({ moments: 2 });
		expect(years[1].months[0].entries[0].side).toBe('left');
	});
});

describe('flattenAndLayoutTimeline', () => {
	it('carries the moment image through to the unified item', () => {
		const data: TimelineByYearResponse = {
			'2024': {
				moments: [
					{
						title: 'Moment',
						shortUrl: 'moment',
						url: '/moments/2024/01/03/moment',
						image: 'https://example.com/one.jpg',
						publishedAt: '2024-01-03T00:00:00Z'
					}
				]
			}
		};

		expect(flattenAndLayoutTimeline(data).items[0]?.image).toBe('https://example.com/one.jpg');
	});
});

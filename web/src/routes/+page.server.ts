import { getHomeActivityPulse } from '$lib/features/home/api';
import { resolveHomeThemeConfig } from '$lib/features/home/theme';
import { getRecentMoments } from '$lib/features/moment/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;
	const parentData = await event.parent();
	const homeTheme = resolveHomeThemeConfig(parentData.websiteInfo);
	trackISRDeps(event, 'home:recent-moments', 'home:activity-pulse', 'home:inspiration-stats');

	const [recentResult, activityResult] = await Promise.allSettled([
		getRecentMoments(fetch),
		getHomeActivityPulse(fetch, { days: 365 })
	]);

	return {
		recentMoments:
			recentResult.status === 'fulfilled'
				? recentResult.value
				: { items: [], total: 0, page: 1, size: 3 },
		recentMomentsFailed: recentResult.status === 'rejected',
		activityPulse: activityResult.status === 'fulfilled' ? activityResult.value : null,
		activityPulseFailed: activityResult.status === 'rejected',
		homeTheme
	};
};

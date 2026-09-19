import { getHomeActivityPulse, getHomeInspirationStats } from '$lib/features/home/api';
import { resolveHomeThemeConfig } from '$lib/features/home/theme';
import { getRecentMoments } from '$lib/features/moment/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;
	const parentData = await event.parent();
	const homeTheme = resolveHomeThemeConfig(parentData.websiteInfo);
	const configuredRangeDays = homeTheme.activityPulse?.rangeDays;
	const activityDays =
		configuredRangeDays === 'all'
			? 'all'
			: configuredRangeDays && configuredRangeDays > 0
				? configuredRangeDays
				: 365;

	trackISRDeps(
		event,
		'home:recent-moments',
		'home:activity-pulse',
		'home:inspiration-stats'
	);

	const [recentMoments, activityPulse, inspirationStats] = await Promise.all([
		getRecentMoments(fetch),
		getHomeActivityPulse(fetch, { days: activityDays }),
		getHomeInspirationStats(fetch, { githubUsername: homeTheme.inspiration?.github?.username })
	]);

	return {
		recentMoments,
		activityPulse,
		inspirationStats,
		homeTheme
	};
};

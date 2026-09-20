import { getApi } from '$lib/shared/clients/api';
import type { HomeActivityPulseData, HomeInspirationStatsData } from './types';

type GetHomeActivityPulseOptions = {
	days?: number | 'all';
};

type GetHomeInspirationStatsOptions = {
	githubUsername?: string;
};

export const getHomeActivityPulse = async (
	fetcher?: typeof fetch,
	options: GetHomeActivityPulseOptions = {}
): Promise<HomeActivityPulseData> => {
	const api = getApi(fetcher);
	const days =
		options.days === 'all'
			? 'all'
			: options.days && options.days > 0
				? String(Math.floor(options.days))
				: '365';
	const query = new URLSearchParams({
		days
	});
	const result = await api<HomeActivityPulseData>(
		`/public/home/activity-pulse?${query.toString()}`
	);
	if (!result || !Array.isArray(result.points) || result.points.length === 0) {
		throw new Error('创作记录返回数据不完整');
	}
	return result;
};

export const getHomeInspirationStats = async (
	fetcher?: typeof fetch,
	options: GetHomeInspirationStatsOptions = {}
): Promise<HomeInspirationStatsData> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams();
	if (options.githubUsername) {
		const username = options.githubUsername.trim();
		if (username.length > 0) {
			query.set('githubUsername', username);
		}
	}
	const suffix = query.size > 0 ? `?${query.toString()}` : '';
	const result = await api<HomeInspirationStatsData>(`/public/home/inspiration-stats${suffix}`);
	return (
		result ?? {
			words: {
				total: 0,
				moments: 0
			}
		}
	);
};

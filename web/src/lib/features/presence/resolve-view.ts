import { buildMomentPath } from '$lib/shared/utils/content-path';
import type { PresenceClientReport } from '$lib/features/presence/types';

type RouteData = {
	moment?: {
		shortUrl?: string | null;
		createdAt?: string | null;
	} | null;
};

const normalizePath = (pathname: string): string => {
	if (!pathname) return '/';
	if (pathname !== '/' && pathname.endsWith('/')) {
		return pathname.slice(0, -1);
	}
	return pathname;
};

export const resolvePresenceView = (
	pathname: string,
	data: unknown
): PresenceClientReport | null => {
	const currentPath = normalizePath(pathname);
	const routeData = (data ?? {}) as RouteData;

	const momentShortUrl = routeData.moment?.shortUrl;
	const momentCreatedAt = routeData.moment?.createdAt;
	if (momentShortUrl && momentCreatedAt) {
		return {
			contentType: 'moment',
			url: buildMomentPath(momentShortUrl, momentCreatedAt)
		};
	}

	if (currentPath.startsWith('/internal/')) {
		return null;
	}

	return {
		contentType: 'page',
		url: currentPath
	};
};

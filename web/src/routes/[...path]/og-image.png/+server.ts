import type { RequestHandler } from './$types';
import { fetchWebsiteInfo } from '$lib/features/website-info/api';
import type { WebsiteInfoMap } from '$lib/features/website-info/types';
import { getMomentDetail } from '$lib/features/moment/api';
import { resolveSeoMeta, resolveOgTag } from '$lib/shared/seo/metadata';
import { renderOgImage } from '$lib/server/og-image-renderer';
import { error } from '@sveltejs/kit';

export const trailingSlash = 'never';

const MOMENT_PATTERN = /^moments\/(\d{4})\/(\d{2})\/(\d{2})\/([^/]+)$/;

const KNOWN_LIST_ROUTES = new Set([
	'moments',
	'tags',
	'timeline',
	'friends',
	'friends-timeline',
	'statistics'
]);

function isKnownListRoute(path: string): boolean {
	const first = path.split('/')[0] ?? '';
	return KNOWN_LIST_ROUTES.has(first);
}

async function fetchRouteData(
	path: string,
	fetcher: typeof fetch
): Promise<{ pathname: string; routeData: Record<string, unknown> }> {
	const pathname = `/${path}`;

	const momentMatch = path.match(MOMENT_PATTERN);
	if (momentMatch) {
		const slug = momentMatch[4]!;
		const moment = await getMomentDetail(fetcher, slug);
		if (!moment) error(404, 'Moment not found');
		return { pathname, routeData: { moment } };
	}

	if (isKnownListRoute(path)) {
		return { pathname, routeData: {} };
	}

	// Fallback — return generic metadata
	return { pathname, routeData: {} };
}

export const GET: RequestHandler = async ({ params, fetch, url }) => {
	const path = params.path;
	if (!path) error(404, 'Not found');

	const websiteInfo: WebsiteInfoMap = await fetchWebsiteInfo(fetch).catch(() => ({}));
	const { pathname, routeData } = await fetchRouteData(path, fetch);

	const seo = resolveSeoMeta({
		pathname,
		routeData,
		websiteInfo,
		origin: url.origin
	});

	// If the page has a content image (e.g. moment cover), no generated OG image is needed
	const hasContentImage = seo.ogImage && seo.ogImageType === null && seo.ogImageWidth === null;
	if (hasContentImage) {
		return new Response(null, { status: 204 });
	}

	const png = await renderOgImage(
		{
			title: seo.ogTitle,
			subtitle: seo.ogDescription,
			site: seo.ogSiteName,
			tag: resolveOgTag(pathname, seo.ogType),
			iconUrl: websiteInfo.favicon || '',
			fallbackIconUrl: ''
		},
		fetch,
		url
	);

	return new Response(png, {
		headers: {
			'content-type': 'image/png',
			'content-length': String(png.byteLength),
			'cache-control': 'public, max-age=0, s-maxage=86400, stale-while-revalidate=604800'
		}
	});
};

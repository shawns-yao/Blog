import { getApi } from '$lib/shared/clients/api';
import { brand } from '$lib/shared/brand/brand';
import type { WebsiteInfoMap } from './types';

export async function fetchWebsiteInfo(fetcher?: typeof fetch): Promise<WebsiteInfoMap> {
	const api = getApi(fetcher);
	const info = await api<WebsiteInfoMap>('/public/website-info');
	for (const key of ['website_name', 'home_title', 'og_site_name', 'og_title'] as const) {
		if (!info[key] || /^(?:blog|shawn|grtblog(?:-v\d+)?)$/i.test(info[key].trim())) {
			info[key] = brand.name;
		}
	}
	return info;
}

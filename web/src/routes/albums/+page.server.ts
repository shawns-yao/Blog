import { redirect } from '@sveltejs/kit';
import { getAlbumList } from '$lib/features/album/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const TRACKED_ALBUM_LIST_PAGES = 3;
const DEFAULT_PAGE_SIZE = 20;

export const load: PageServerLoad = async (event) => {
	const { fetch, url } = event;
	const rawPage = Number(url.searchParams.get('page') ?? '1');
	const page = Number.isFinite(rawPage) && rawPage > 0 ? rawPage : 1;
	if (page > 1) {
		throw redirect(308, `/albums/page/${page}`);
	}
	if (page <= TRACKED_ALBUM_LIST_PAGES) {
		trackISRDeps(event, `album:list:page:${page}`);
	}

	const data = await getAlbumList(fetch, { page, pageSize: DEFAULT_PAGE_SIZE });
	return {
		albums: data
	};
};

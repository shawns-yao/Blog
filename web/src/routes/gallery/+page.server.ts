import { getAlbumList } from '$lib/features/album/api';
import { getMomentList } from '$lib/features/moment/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const PAGE_SIZE = 24;

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;

	const [moments, albums] = await Promise.all([
		getMomentList(fetch, { page: 1, pageSize: PAGE_SIZE }),
		getAlbumList(fetch, { page: 1, pageSize: PAGE_SIZE })
	]);

	trackISRDeps(event, 'gallery:list');

	return { moments, albums };
};

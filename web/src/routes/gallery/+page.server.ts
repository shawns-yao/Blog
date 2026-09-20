import { getMomentList } from '$lib/features/moment/api';
import { getColumns } from '$lib/features/taxonomy/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const PAGE_SIZE = 24;

export const load: PageServerLoad = async (event) => {
	const { fetch, url } = event;
	const rawPage = Number(url.searchParams.get('page') ?? 1);
	const page = Number.isSafeInteger(rawPage) && rawPage > 0 ? rawPage : 1;
	const rawColumn = Number(url.searchParams.get('column'));
	const columnId = Number.isSafeInteger(rawColumn) && rawColumn > 0 ? rawColumn : undefined;
	const [moments, columns] = await Promise.all([
		getMomentList(fetch, { page, pageSize: PAGE_SIZE, contentKind: 'article', columnId }),
		getColumns(fetch)
	]);

	trackISRDeps(event, 'gallery:list');
	trackISRDeps(event, 'column:list');

	return { moments, columns, columnId };
};

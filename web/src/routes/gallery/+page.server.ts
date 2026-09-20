import { error } from '@sveltejs/kit';
import { getMomentList, getMomentDetail } from '$lib/features/moment/api';
import { getColumns } from '$lib/features/taxonomy/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const PAGE_SIZE = 24;

export const load: PageServerLoad = async (event) => {
	const { fetch, url } = event;
	const rawPage = Number(url.searchParams.get('page') ?? 1);
	const page = Number.isSafeInteger(rawPage) && rawPage > 0 ? rawPage : 1;
	const rawColumn = Number(url.searchParams.get('column'));
	let columnId = Number.isSafeInteger(rawColumn) && rawColumn > 0 ? rawColumn : undefined;
	const slug = url.searchParams.get('read');
	let selected = slug ? await getMomentDetail(fetch, slug) : null;
	if (slug && (!selected || selected.contentKind !== 'article' || !selected.isPublished)) {
		error(404, '文章不存在');
	}
	if (selected) {
		if (columnId && selected.columnId !== columnId) error(404, '文章不属于当前分类');
		columnId = selected.columnId ?? undefined;
	}
	const query = (url.searchParams.get('q') ?? '').trim().slice(0, 200);
	const [moments, columns] = await Promise.all([
		getMomentList(fetch, {
			page,
			pageSize: PAGE_SIZE,
			contentKind: 'article',
			columnId,
			search: selected ? undefined : query
		}),
		getColumns(fetch)
	]);
	const column = columns.find((item) => item.id === columnId) ?? null;
	if (columnId && !column) error(404, '分类不存在');

	trackISRDeps(event, 'gallery:list');
	trackISRDeps(event, 'column:list');
	if (selected) trackISRDeps(event, `moment:detail:${selected.id}`);

	return { moments, columns, columnId, column, selected, reading: !!selected, query };
};

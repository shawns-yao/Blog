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
	const columns = await getColumns(fetch);
	let column = columns.find((item) => item.id === columnId) ?? null;
	if (columnId && !column) error(404, '分类不存在');
	let selected = slug ? await getMomentDetail(fetch, slug) : null;
	if (slug && (!selected || selected.contentKind !== 'article' || !selected.isPublished)) {
		error(404, '文章不存在');
	}
	if (selected) {
		if (columnId && selected.columnId !== columnId) error(404, '文章不属于当前分类');
		columnId = selected.columnId ?? undefined;
		column = columns.find((item) => item.id === columnId) ?? null;
	}
	const root = column?.parentId
		? (columns.find((item) => item.id === column.parentId) ?? null)
		: column;
	const secondary = column?.parentId ? column : null;
	const listView = url.searchParams.get('view') === 'list';
	const query = (url.searchParams.get('q') ?? '').trim().slice(0, 200);
	const moments = await getMomentList(fetch, {
		page,
		pageSize: PAGE_SIZE,
		contentKind: 'article',
		columnId,
		search: query,
		includeChildren: !!root && !secondary,
		newestFirst: true
	});
	if (secondary && !selected && !listView && !query && moments.items.length) {
		selected = await getMomentDetail(fetch, moments.items[0].shortUrl);
		if (
			!selected ||
			!selected.isPublished ||
			selected.contentKind !== 'article' ||
			selected.columnId !== secondary.id
		) {
			error(404, '文章已变更，请重新选择分类');
		}
	}

	trackISRDeps(event, 'gallery:list');
	trackISRDeps(event, 'column:list');
	if (selected) trackISRDeps(event, `moment:detail:${selected.id}`);

	return {
		moments,
		columns,
		columnId,
		column,
		root,
		secondary,
		selected,
		reading: !!selected || (!!secondary && !listView && !query),
		query
	};
};

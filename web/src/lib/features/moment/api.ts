import { getApi, fetchOrNull } from '$lib/shared/clients/api';
import type {
	MomentDetail,
	MomentLatestCheckResponse,
	MomentListResponse,
	MomentRelatedMoment
} from '$lib/features/moment/types';

type MomentListOptions = {
	page?: number;
	pageSize?: number;
	contentKind?: 'note' | 'article';
	columnId?: number;
};

export const getMomentList = async (
	fetcher?: typeof fetch,
	{ page = 1, pageSize = 10, contentKind = 'note', columnId }: MomentListOptions = {}
): Promise<MomentListResponse> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(page),
		pageSize: String(pageSize),
		contentKind
	});
	if (columnId) query.set('columnId', String(columnId));
	const result = await api<MomentListResponse>(`/moments?${query.toString()}`);
	return result ?? { items: [], total: 0, page, size: pageSize };
};

export const getMomentListByColumn = async (
	fetcher?: typeof fetch,
	columnSlug: string = '',
	{ page = 1, pageSize = 20 }: MomentListOptions = {}
): Promise<MomentListResponse> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(page),
		pageSize: String(pageSize)
	});
	const result = await api<MomentListResponse>(
		`/columns/short/${encodeURIComponent(columnSlug)}/moments?${query.toString()}`
	);
	return result ?? { items: [], total: 0, page, size: pageSize };
};

export const getMomentDetail = async (
	fetcher: typeof fetch | undefined,
	shortUrl: string
): Promise<MomentDetail | null> => {
	const api = getApi(fetcher);
	return fetchOrNull(() => api<MomentDetail>(`/moments/short/${shortUrl}`));
};

export const checkMomentLatest = async (
	fetcher: typeof fetch | undefined,
	id: number,
	hash: string
): Promise<MomentLatestCheckResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<MomentLatestCheckResponse>(`/moments/${id}/latest`, {
		method: 'POST',
		body: { hash }
	});
	return result ?? null;
};

export const getRecentMoments = async (fetcher?: typeof fetch): Promise<MomentListResponse> => {
	const api = getApi(fetcher);
	const result = await api<MomentListResponse>('/public/moments/recent');
	if (!result || !Array.isArray(result.items)) throw new Error('手记列表返回数据不完整');
	const recent = [...result.items]
		.sort((a, b) => Date.parse(b.createdAt) - Date.parse(a.createdAt) || b.id - a.id)
		.slice(0, 3);
	const items = await Promise.all(
		recent.map(async (moment) => {
			if (moment.cover && moment.summary) return moment;
			try {
				const detail = await getMomentDetail(fetcher, moment.shortUrl);
				if (!detail) return moment;
				const [{ parseMarkdown }, { extractPlainTextFromNodes, extractImageUrlsFromNodes }] =
					await Promise.all([import('svmarkdown'), import('$lib/shared/markdown/component-body')]);
				const { children } = parseMarkdown(detail.content);
				return {
					...moment,
					summary: moment.summary || extractPlainTextFromNodes(children).slice(0, 360),
					cover: moment.cover || extractImageUrlsFromNodes(children)[0]
				};
			} catch {
				return moment;
			}
		})
	);
	return { ...result, items };
};

type MomentSamePeriodResponse = {
	items: MomentRelatedMoment[];
};

export const getMomentSamePeriodMoments = async (
	fetcher: typeof fetch | undefined,
	id: number
): Promise<MomentRelatedMoment[]> => {
	const api = getApi(fetcher);
	const result = await api<MomentSamePeriodResponse>(`/moments/${id}/same-period-moments`);
	return result?.items ?? [];
};

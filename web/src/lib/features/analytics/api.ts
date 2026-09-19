import { getApi } from '$lib/shared/clients/api';
import type {
	ContentMetrics,
	TrackLikePayload,
	TrackLikeResponse,
	TrackViewContentType,
	TrackViewPayload,
	TrackViewResponse
} from '$lib/features/analytics/types';

export const trackContentView = async (
	fetcher: typeof fetch | undefined,
	payload: TrackViewPayload
): Promise<TrackViewResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<TrackViewResponse>('/public/analytics/view', {
		method: 'POST',
		body: payload
	});
	return result ?? null;
};

const metricsPathMap: Record<TrackViewContentType, string> = {
	moment: '/moments',
	album: '/albums'
};

export const fetchContentMetrics = async (
	contentType: TrackViewContentType,
	contentId: number
): Promise<ContentMetrics | null> => {
	const api = getApi();
	try {
		return await api<ContentMetrics>(`${metricsPathMap[contentType]}/${contentId}/metrics`);
	} catch {
		return null;
	}
};

export const trackContentLike = async (
	fetcher: typeof fetch | undefined,
	payload: TrackLikePayload
): Promise<TrackLikeResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<TrackLikeResponse>('/public/analytics/like', {
		method: 'POST',
		body: payload
	});
	return result ?? null;
};

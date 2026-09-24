import { error } from '@sveltejs/kit';
import {
	getMomentDetail,
	getMomentList,
	getMomentSamePeriodMoments
} from '$lib/features/moment/api';
import type { MomentRelatedMoment } from '$lib/features/moment/types';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch, params } = event;
	const detail = await getMomentDetail(fetch, params.slug);
	if (!detail) {
		error(404, 'Moment not found');
	}

	const matched = detail.createdAt.match(/^(\d{4})-(\d{2})-(\d{2})/);
	if (!matched) {
		error(404, 'Moment not found');
	}
	const [, year, month, day] = matched;
	if (
		params.year !== year ||
		params.month !== month ||
		params.day !== day ||
		params.slug !== detail.shortUrl
	) {
		error(404, 'Moment not found');
	}
	trackISRDeps(event, `moment:detail:${detail.id}`);

	const [relatedMoments, underlayMoments] = await Promise.all([
		getMomentSamePeriodMoments(fetch, detail.id).catch(() => [] as MomentRelatedMoment[]),
		getMomentList(fetch, { page: 1, pageSize: 20 }).catch(() => ({
			items: [],
			total: 0,
			page: 1,
			size: 20
		}))
	]);
	trackISRDeps(event, 'moment:list:page:1');

	return {
		underlayMoments,
		moment: {
			...detail,
			relatedMoments
		}
	};
};

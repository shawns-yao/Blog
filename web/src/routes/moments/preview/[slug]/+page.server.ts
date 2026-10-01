import { dev } from '$app/environment';
import { getLayoutPreviewDetail, layoutPreviewMoments } from '$lib/features/moment/layout-preview';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const moment = dev ? getLayoutPreviewDetail(params.slug) : null;
	if (!moment) error(404, 'Moment not found');
	return {
		moment,
		underlayMoments: {
			items: layoutPreviewMoments,
			total: layoutPreviewMoments.length,
			page: 1,
			size: layoutPreviewMoments.length
		}
	};
};

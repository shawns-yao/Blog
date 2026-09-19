import type { MomentSummary } from '$lib/features/moment/types';

export type Tag = {
	id: number;
	name: string;
};

export type PublicTag = {
	id: number;
	name: string;
	momentCount: number;
};

export type TagContents = {
	moments: MomentSummary[];
};

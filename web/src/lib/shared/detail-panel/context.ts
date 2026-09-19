import type { TOCNode } from '$lib/shared/types/toc';
import { createModelDataContext } from 'svatoms';

export type DetailPanelKind = 'moment' | null;

export type DetailPanelRelatedMoment = {
	id: number;
	title: string;
	shortUrl: string;
	summary: string;
	createdAt: string;
};

export type DetailPanelModel = {
	kind: DetailPanelKind;
	title: string;
	toc: TOCNode[];
	relatedMoments: DetailPanelRelatedMoment[];
	contentRoot: HTMLElement | null;
	activeAnchor: string | null;
};

export const createEmptyDetailPanelModel = (): DetailPanelModel => ({
	kind: null,
	title: '',
	toc: [],
	relatedMoments: [],
	contentRoot: null,
	activeAnchor: null
});

export const detailPanelCtx = createModelDataContext<DetailPanelModel>({
	name: 'detailPanelCtx',
	initial: createEmptyDetailPanelModel()
});
